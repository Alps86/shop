package runtime

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"go.temporal.io/sdk/client"
)

func Open(ctx context.Context, config Config) (*App, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	store, err := postgres.Open(ctx, postgres.Config{DSN: config.DSN, Schema: config.Schema, Context: config.Context, ReplicationSlot: config.Slot})
	if err != nil {
		return nil, err
	}

	app := &App{Config: config, Store: store}
	if err := app.connect(ctx); err != nil {
		return nil, errors.Join(err, app.Close(ctx))
	}

	return app, nil
}

func (app *App) connect(ctx context.Context) error {
	if err := app.Store.Migrate(ctx); err != nil {
		return err
	}

	if err := app.importSeed(ctx); err != nil {
		return err
	}

	snapshot, err := app.Store.Catalog().Snapshot(ctx)
	if err != nil {
		return err
	}

	return app.connectTemporal(ctx, snapshot)
}

func (app *App) connectTemporal(ctx context.Context, snapshot application.AdminCatalogSnapshot) error {
	connection, err := client.DialContext(ctx, client.Options{HostPort: app.Config.TemporalAddress, Namespace: app.Config.TemporalNamespace})
	if err != nil {
		return err
	}

	app.Client = connection

	return app.bind(snapshot)
}

func (app *App) Close(ctx context.Context) error {
	var err error
	if app.Consumer != nil {
		err = app.Consumer.Close(ctx)
	}

	if app.Client != nil {
		app.Client.Close()
	}

	app.Store.Close()

	return err
}
