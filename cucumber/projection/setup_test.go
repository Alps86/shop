package projection_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
)

func (current *world) connect(table *godog.Table) error {
	dsn := current.environment("SHOP_TEST_POSTGRES_DSN", "postgres://shop@127.0.0.1:15432/postgres?sslmode=disable")
	connection, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	current.Admin = connection
	current.Config = cdc.Config{DSN: dsn, Schema: current.Alias, Publication: current.Alias + "_pub", Slot: current.Alias + "_slot"}
	current.Store, err = postgres.Open(context.Background(), postgres.Config{DSN: dsn, Schema: current.Config.Schema, Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}})
	if err != nil {
		return err
	}

	if err := current.Store.Migrate(context.Background()); err != nil {
		return err
	}

	return current.connectSearch(table)
}

func (current *world) connectSearch(table *godog.Table) error {
	row := current.rows(table)[0]
	current.Config.Store = current.Store
	current.Source = &gatedSource{Source: cdc.NewReindexSource(current.Config)}
	current.Endpoint = current.environment("SHOP_TEST_OPENSEARCH_URL", "http://127.0.0.1:19200")
	requestContext := value.RequestContext{Locale: value.Locale(row["Locale"]), Currency: value.Currency(row["Currency"]), Market: value.CountryCode(row["Market"])}
	current.Search = opensearch.New(opensearch.Config{Endpoint: current.Endpoint, Index: current.Alias, Context: requestContext, Source: current.Source})
	if err := current.Search.Initialize(context.Background()); err != nil {
		return err
	}

	current.Config.Target = current.Search

	return current.startConsumer()
}

func (current *world) startConsumer() error {
	current.Consumer = cdc.New(current.Config)
	if err := current.Consumer.Prepare(context.Background()); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	current.Cancel = cancel
	current.Stopped = make(chan error, 1)
	go func() {
		current.Stopped <- current.Consumer.Stream(ctx)
	}()

	return nil
}

func (current *world) stopConsumer() error {
	if current.Cancel == nil {
		return nil
	}

	current.Cancel()
	err := <-current.Stopped
	current.Cancel = nil
	if errors.Is(err, context.Canceled) {
		err = nil
	}

	return errors.Join(err, current.Consumer.Close(context.Background()))
}
