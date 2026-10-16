package postgres_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/cucumber/godog"
)

func (current *world) diagnosticPool(row map[string]string) error {
	if row["PoolMaxConns"] == "" {
		return nil
	}

	dsn, err := url.Parse(os.Getenv("SHOP_TEST_POSTGRES_DSN"))
	if err != nil {
		return err
	}

	query := dsn.Query()
	query.Set("pool_max_conns", row["PoolMaxConns"])
	dsn.RawQuery = query.Encode()
	current.store.Close()
	current.store, err = postgres.Open(context.Background(), postgres.Config{DSN: dsn.String(), Schema: current.schema, Context: current.requestContext(nil)})

	return err
}

func (current *world) poolObserved(input *godog.Table) error {
	for _, row := range input.Rows[1:] {
		if err := current.poolMinimum(row.Cells[0].Value, row.Cells[1].Value); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) poolMinimum(field, minimum string) error {
	actual, err := strconv.ParseInt(current.observed[field], 10, 64)
	if err != nil {
		return fmt.Errorf("%s ist keine ganze Zahl: %w", field, err)
	}

	limit, err := strconv.ParseInt(minimum, 10, 64)
	if err != nil {
		return err
	}

	if actual < limit {
		return fmt.Errorf("%s: %d unter %d", field, actual, limit)
	}

	return nil
}
