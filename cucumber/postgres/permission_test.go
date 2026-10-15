package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/jackc/pgx/v5"
)

func (current *world) isolatedDigestPermission() error {
	current.permissionDatabase = current.schema + "_restricted"
	current.permissionRole = current.schema + "_reader"
	if err := current.createPermissionDatabase(); err != nil {
		return err
	}

	connection, err := current.connectPermissionDatabase()
	if err != nil {
		return err
	}

	err = current.denyDigest(connection)

	return errors.Join(err, connection.Close(context.Background()))
}

func (current *world) createPermissionDatabase() error {
	if _, err := current.admin.Exec(context.Background(), "CREATE ROLE "+pgx.Identifier{current.permissionRole}.Sanitize()+" NOLOGIN"); err != nil {
		return err
	}

	_, err := current.admin.Exec(context.Background(), "CREATE DATABASE "+pgx.Identifier{current.permissionDatabase}.Sanitize()+" OWNER "+pgx.Identifier{current.permissionRole}.Sanitize())

	return err
}

func (current *world) connectPermissionDatabase() (*pgx.Conn, error) {
	dsn, err := url.Parse(os.Getenv("SHOP_TEST_POSTGRES_DSN"))
	if err != nil {
		return nil, err
	}

	dsn.Path = "/" + current.permissionDatabase
	if err := current.openPermissionStore(dsn); err != nil {
		return nil, err
	}

	return pgx.Connect(context.Background(), dsn.String())
}

func (current *world) openPermissionStore(dsn *url.URL) error {
	limited := *dsn
	query := limited.Query()
	query.Set("options", "-crole="+current.permissionRole)
	limited.RawQuery = query.Encode()
	current.store.Close()
	store, err := postgres.Open(context.Background(), postgres.Config{DSN: limited.String(), Schema: current.schema, Context: current.requestContext(nil)})
	if err != nil {
		return err
	}

	current.store = store

	return store.Migrate(context.Background())
}

func (current *world) denyDigest(connection *pgx.Conn) error {
	var database string
	if err := connection.QueryRow(context.Background(), "SELECT current_database()").Scan(&database); err != nil {
		return err
	}

	if database != current.permissionDatabase {
		return fmt.Errorf("Testdatenbank %s statt %s", database, current.permissionDatabase)
	}

	_, err := connection.Exec(context.Background(), "REVOKE EXECUTE ON FUNCTION pg_catalog.sha256(bytea) FROM PUBLIC")

	return err
}

func (current *world) cleanupPermission(ctx context.Context) error {
	if current.permissionDatabase == "" {
		return nil
	}

	_, databaseErr := current.admin.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{current.permissionDatabase}.Sanitize())
	_, roleErr := current.admin.Exec(ctx, "DROP ROLE IF EXISTS "+pgx.Identifier{current.permissionRole}.Sanitize())

	return errors.Join(databaseErr, roleErr)
}
