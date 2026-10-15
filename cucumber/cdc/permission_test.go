package cdc_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
)

func (current *world) restrictedDatabase(input *godog.Table) error {
	current.foreignDatabase = current.config.Schema + "_restricted"
	current.restrictedRole = current.config.Schema + "_role"
	if err := current.createRestrictedDatabase(); err != nil {
		return err
	}

	root, limited, err := current.permissionConfigurations()
	if err != nil {
		return err
	}

	if err := current.openRestrictedStore(limited); err != nil {
		return err
	}

	current.config.DSN = root

	return current.configureRestriction(current.fields(input)["Permission"], root, limited)
}

func (current *world) createRestrictedDatabase() error {
	_, err := current.admin.Exec(context.Background(), "CREATE ROLE "+pgx.Identifier{current.restrictedRole}.Sanitize()+" LOGIN PASSWORD 'temporary-contract'")
	if err != nil {
		return err
	}

	_, err = current.admin.Exec(context.Background(), "CREATE DATABASE "+pgx.Identifier{current.foreignDatabase}.Sanitize()+" OWNER "+pgx.Identifier{current.restrictedRole}.Sanitize())

	return err
}

func (current *world) permissionConfigurations() (string, string, error) {
	config, err := url.Parse(current.config.DSN)
	if err != nil {
		return "", "", err
	}

	config.Path = "/" + current.foreignDatabase
	root := config.String()
	config.User = url.UserPassword(current.restrictedRole, "temporary-contract")

	return root, config.String(), nil
}

func (current *world) configureRestriction(permission, root, limited string) error {
	if permission == "Replication" {
		current.config.DSN = limited
		current.source = cdc.NewReindexSource(current.config)

		return nil
	}

	connection, err := pgx.Connect(context.Background(), root)
	if err != nil {
		return err
	}

	if err := current.verifyOwnDatabase(connection); err != nil {
		return errors.Join(err, connection.Close(context.Background()))
	}

	_, err = connection.Exec(context.Background(), "REVOKE SELECT ON pg_catalog.pg_replication_slots FROM PUBLIC")
	current.source = cdc.NewReindexSource(current.config)

	return errors.Join(err, connection.Close(context.Background()))
}

func (current *world) dropRestrictedRole(ctx context.Context) error {
	if current.restrictedRole == "" {
		return nil
	}

	_, err := current.admin.Exec(ctx, "DROP ROLE "+pgx.Identifier{current.restrictedRole}.Sanitize())

	return err
}

func (current *world) verifyOwnDatabase(connection *pgx.Conn) error {
	var database string
	if err := connection.QueryRow(context.Background(), "SELECT current_database()").Scan(&database); err != nil {
		return err
	}

	if database != current.foreignDatabase {
		return fmt.Errorf("Testdatenbank %s statt %s", database, current.foreignDatabase)
	}

	return nil
}

func (current *world) openRestrictedStore(limited string) error {
	current.store.Close()
	store, err := postgres.Open(context.Background(), postgres.Config{DSN: limited, Schema: current.config.Schema, Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}})
	if err != nil {
		return err
	}

	current.store = store
	current.config.Store = store

	return store.Migrate(context.Background())
}
