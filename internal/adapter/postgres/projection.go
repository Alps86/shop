package postgres

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (store *Store) project(ctx context.Context, tx pgx.Tx, id value.ProductID) error {
	var revision uint64
	if err := tx.QueryRow(ctx, "UPDATE projection_versions SET revision=revision+1 WHERE product_id=$1 RETURNING revision", id).Scan(&revision); err != nil {
		return databaseError(err)
	}

	projection, err := store.projection(ctx, tx, id, revision)
	if err != nil {
		return err
	}

	envelope := event.ProjectionEnvelope{EventID: "e-" + rand.Text(), SchemaVersion: 1, OccurredAt: time.Now().UTC(), Kind: "upsert", Context: store.config.Context, Projection: projection}

	return store.insertProjection(ctx, tx, envelope)
}

func (store *Store) projection(ctx context.Context, db database, id value.ProductID, revision uint64) (application.ProjectionEvent, error) {
	product := catalog.Product{}
	if err := newRecord("product", string(id), &product).read(ctx, db); err != nil {
		return application.ProjectionEvent{}, err
	}

	inventories, err := store.productInventory(ctx, db, product)

	return application.ProjectionEvent{ProductID: id, Revision: revision, Product: value.Optional[catalog.Product]{Value: product, Present: true}, Inventory: inventories}, err
}

func (store *Store) insertProjection(ctx context.Context, tx pgx.Tx, envelope event.ProjectionEnvelope) error {
	_, err := tx.Exec(ctx, "INSERT INTO outbox(envelope) VALUES($1)", envelope)
	return databaseError(err)
}

func (store *Store) process(ctx context.Context, tx pgx.Tx, kind string, cart sale.Cart) error {
	envelope := event.ProcessEnvelope{EventID: "e-" + rand.Text(), Kind: kind, CartID: cart.ID, Binding: application.ProcessBinding{DefinitionVersion: cart.DefinitionVersion, RuleSetVersion: cart.RuleSetVersion}}
	return store.insertProcess(ctx, tx, envelope)
}

func (store *Store) insertProcess(ctx context.Context, tx pgx.Tx, envelope event.ProcessEnvelope) error {
	_, err := tx.Exec(ctx, "INSERT INTO process_outbox(envelope) VALUES($1)", envelope)
	return databaseError(err)
}

func (store *Store) envelope(projection application.ProjectionEvent, kind string) event.ProjectionEnvelope {
	return event.ProjectionEnvelope{EventID: "e-" + rand.Text(), SchemaVersion: 1, OccurredAt: time.Now().UTC(), Kind: kind, Context: store.config.Context, Projection: projection}
}

func (store *Store) projectSKU(ctx context.Context, tx pgx.Tx, sku value.SKU) error {
	product, err := store.productForSKU(ctx, tx, sku)
	if err != nil {
		return err
	}

	return store.project(ctx, tx, product.ID)
}
