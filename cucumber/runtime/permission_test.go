package runtime_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"net/url"
)

func (current *world) deniedCatalog() (err error) {
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, current.config.DSN)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, connection.Close(ctx))
	}()

	return current.withRole(ctx, connection)
}

func (current *world) withRole(ctx context.Context, connection *pgx.Conn) (err error) {
	role := pgx.Identifier{current.config.Schema + "_reader"}.Sanitize()
	if err := current.prepareRole(ctx, connection, role); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, releaseRole(ctx, connection, role))
	}()

	return current.openRestricted(ctx)
}

func (current *world) openRestricted(ctx context.Context) error {
	config, err := current.restrictedConfiguration()
	if err != nil {
		return err
	}

	current.tryOpen(ctx, config)

	return nil
}

func (current *world) restrictedConfiguration() (runtime.Config, error) {
	config := current.config
	endpoint, err := url.Parse(config.DSN)
	if err != nil {
		return config, err
	}

	query := endpoint.Query()
	query.Set("role", current.config.Schema+"_reader")
	endpoint.RawQuery = query.Encode()
	config.DSN = endpoint.String()

	return config, nil
}

func releaseRole(ctx context.Context, connection *pgx.Conn, role string) error {
	_, ownedError := connection.Exec(ctx, "DROP OWNED BY "+role)
	_, roleError := connection.Exec(ctx, "DROP ROLE "+role)

	return errors.Join(ownedError, roleError)
}

func (current *world) prepareRole(ctx context.Context, connection *pgx.Conn, role string) error {
	var database string
	if err := connection.QueryRow(ctx, "SELECT current_database()").Scan(&database); err != nil {
		return err
	}

	_, err := connection.Exec(ctx, "CREATE ROLE "+role+" NOLOGIN")
	if err != nil {
		return err
	}

	_, databaseError := connection.Exec(ctx, "GRANT CREATE ON DATABASE "+pgx.Identifier{database}.Sanitize()+" TO "+role)
	_, schemaError := connection.Exec(ctx, "GRANT USAGE, CREATE ON SCHEMA "+pgx.Identifier{current.config.Schema}.Sanitize()+" TO "+role)
	_, insertError := connection.Exec(ctx, "GRANT INSERT ON ALL TABLES IN SCHEMA "+pgx.Identifier{current.config.Schema}.Sanitize()+" TO "+role)

	return errors.Join(databaseError, schemaError, insertError)
}

func (current *world) permissionSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine neue Laufzeit ohne Leseberechtigung für den Katalog geöffnet wird$`, current.deniedCatalog)
}
