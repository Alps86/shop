package postgres

import (
	"context"
	_ "embed"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var migration string

func Open(ctx context.Context, config Config) (*Store, error) {
	poolConfig, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, err
	}

	if config.Schema == "" {
		config.Schema = "public"
	}

	poolConfig.ConnConfig.RuntimeParams["search_path"] = pgx.Identifier{config.Schema}.Sanitize()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)

	return &Store{pool: pool, config: config}, err
}

func (store *Store) Close() {
	store.pool.Close()
}

func (store *Store) Migrate(ctx context.Context) error {
	if _, err := store.pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{store.config.Schema}.Sanitize()); err != nil {
		return err
	}

	_, err := store.pool.Exec(ctx, migration)

	return err
}

func (store *Store) transact(ctx context.Context, operation func(pgx.Tx) error) error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return databaseError(err)
	}

	if err := operation(tx); err != nil {
		return databaseError(errors.Join(err, tx.Rollback(ctx)))
	}

	return databaseError(tx.Commit(ctx))
}

func (store *Store) Guests() GuestRepository {
	return GuestRepository{store: store}
}

func (store *Store) Carts() CartRepository {
	return CartRepository{store: store}
}

func (store *Store) Orders() OrderRepository {
	return OrderRepository{store: store}
}

func (store *Store) Fulfillment() FulfillmentRepository {
	return FulfillmentRepository{store: store}
}

func (store *Store) Catalog() CatalogRepository {
	return CatalogRepository{store: store}
}

func (store *Store) Inventory() InventoryRepository {
	return InventoryRepository{store: store}
}

func (store *Store) Evaluator() EvaluationRepository {
	return EvaluationRepository{store: store}
}

func (store *Store) Purchases() PurchaseRepository {
	return PurchaseRepository{store: store}
}

func (store *Store) Recommendations() RecommendationRepository {
	return RecommendationRepository{store: store}
}

func (store *Store) AdminOrders() AdminOrderRepository {
	return AdminOrderRepository{store: store}
}

func (store *Store) rollback(ctx context.Context, tx pgx.Tx, cause error) error {
	return errors.Join(cause, tx.Rollback(ctx))
}
