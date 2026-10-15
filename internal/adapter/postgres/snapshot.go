package postgres

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (store *Store) ExportedSnapshot(ctx context.Context, name string) ([]application.ProjectionEvent, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, "SET TRANSACTION SNAPSHOT "+"'"+strings.ReplaceAll(name, "'", "''")+"'"); err != nil {
		return nil, store.rollback(ctx, tx, err)
	}

	events, err := store.snapshotEvents(ctx, tx)
	if err != nil {
		return nil, store.rollback(ctx, tx, err)
	}

	return events, tx.Commit(ctx)
}

func (store *Store) snapshotEvents(ctx context.Context, tx pgx.Tx) ([]application.ProjectionEvent, error) {
	rows, err := tx.Query(ctx, "SELECT product_id,revision,deleted FROM projection_versions ORDER BY product_id")
	if err != nil {
		return nil, err
	}

	versions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (projectionVersion, error) {
		version := projectionVersion{}
		err := row.Scan(&version.id, &version.revision, &version.deleted)

		return version, err
	})
	if err != nil {
		return nil, err
	}

	return store.versionEvents(ctx, tx, versions)
}

func (store *Store) versionEvents(ctx context.Context, tx pgx.Tx, versions []projectionVersion) ([]application.ProjectionEvent, error) {
	events := []application.ProjectionEvent{}
	for _, version := range versions {
		projection, err := store.versionEvent(ctx, tx, version)
		if err != nil {
			return nil, err
		}

		events = append(events, projection)
	}

	return events, nil
}

func (store *Store) DeleteProduct(ctx context.Context, id value.ProductID) error {
	return store.transact(ctx, func(tx pgx.Tx) error {
		var revision uint64
		if err := tx.QueryRow(ctx, "UPDATE projection_versions SET revision=revision+1,deleted=true WHERE product_id=$1 RETURNING revision", id).Scan(&revision); err != nil {
			return databaseError(err)
		}

		envelope := store.envelope(application.ProjectionEvent{ProductID: id, Revision: revision}, "delete")
		if _, err := tx.Exec(ctx, "DELETE FROM records WHERE kind='product' AND id=$1", id); err != nil {
			return err
		}

		return store.insertProjection(ctx, tx, envelope)
	})
}

func (store *Store) versionEvent(ctx context.Context, tx pgx.Tx, version projectionVersion) (application.ProjectionEvent, error) {
	if version.deleted {
		return application.ProjectionEvent{ProductID: version.id, Revision: version.revision}, nil
	}

	return store.projection(ctx, tx, version.id, version.revision)
}
