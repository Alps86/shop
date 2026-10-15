package postgres

import (
	"context"
	"errors"

	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func newRecord(kind, id string, destination any) record {
	return record{kind: kind, id: id, value: destination}
}

func (record record) read(ctx context.Context, db database) error {
	return databaseError(db.QueryRow(ctx, "SELECT payload FROM records WHERE kind=$1 AND id=$2", record.kind, record.id).Scan(record.value))
}

func (record record) lock(ctx context.Context, db database) error {
	return databaseError(db.QueryRow(ctx, "SELECT payload FROM records WHERE kind=$1 AND id=$2 FOR UPDATE", record.kind, record.id).Scan(record.value))
}

func (record record) save(ctx context.Context, db database) error {
	_, err := db.Exec(ctx, "INSERT INTO records(kind,id,payload) VALUES($1,$2,COALESCE($3::jsonb,'null'::jsonb)) ON CONFLICT(kind,id) DO UPDATE SET payload=EXCLUDED.payload", record.kind, record.id, record.value)
	return databaseError(err)
}

func (store *Store) documents(ctx context.Context, kind string, destination any) error {
	return databaseError(store.pool.QueryRow(ctx, "SELECT COALESCE(jsonb_agg(payload ORDER BY id),'[]'::jsonb) FROM records WHERE kind=$1", kind).Scan(destination))
}

func databaseError(err error) error {
	if value.Code(err) != "InternalError" {
		return err
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return value.NewError("NotFound", "Resource")
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return errors.Join(postgresFailure(postgresError), err)
	}

	return errors.Join(value.NewError("DependencyUnavailable", "PostgreSQL"), err)
}

func postgresFailure(err *pgconn.PgError) error {
	if err.Code == "23505" {
		return value.NewError("Conflict", err.ConstraintName)
	}

	return value.NewError("DependencyUnavailable", "PostgreSQL")
}
