package runtime

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/hugo"
	"github.com/Alps86/shop/internal/adapter/postgres"
)

func (app *App) ExportContent(ctx context.Context, directory string) error {
	snapshot, err := app.Store.Catalog().Snapshot(ctx)
	if err != nil {
		return err
	}

	return hugo.New(snapshot, app.Config.Context.Locale).WriteDirectory(directory)
}

func (config Config) exportContent(ctx context.Context) error {
	if err := config.Validate(); err != nil {
		return err
	}

	store, err := postgres.Open(ctx, postgres.Config{DSN: config.DSN, Schema: config.Schema, Context: config.Context, ReplicationSlot: config.Slot})
	if err != nil {
		return err
	}

	defer store.Close()
	app := &App{Config: config, Store: store}

	return app.prepareContent(ctx)
}

func (app *App) prepareContent(ctx context.Context) error {
	if err := app.Store.Migrate(ctx); err != nil {
		return err
	}

	if err := app.importSeed(ctx); err != nil {
		return err
	}

	return app.ExportContent(ctx, app.Config.ContentSource)
}
