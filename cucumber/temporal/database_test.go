package temporal_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"os"
	"time"
)

func (current *world) registerDatabase(ctx *godog.ScenarioContext) {
	ctx.Step(`^PostgreSQL den vollständigen Kaufzustand hält$`, current.database)
	ctx.Step(`^ein Gast einen echten Cartprozess öffnet$`, current.guestCart)
	ctx.Step(`^der Gast ausschließlich Workflowupdates verwendet$`, current.configureCheckout)
	ctx.Step(`^der Kauf über den Workflow verbindlich abgeschlossen wird$`, current.buy)
	ctx.Step(`^der dauerhafte Kauf und gebundene Orderprozess beobachtet werden$`, current.purchaseObserved)
}

func (current *world) database(table *godog.Table) error {
	current.schema = "temporal_test_" + current.id[len("temporal-contract-"):]
	dsn := os.Getenv("SHOP_TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://shop@127.0.0.1:15432/postgres?sslmode=disable"
	}

	connection, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	current.admin = connection
	current.databaseStore, err = postgres.Open(context.Background(), postgres.Config{DSN: dsn, Schema: current.schema, Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}})
	if err != nil {
		return err
	}

	return current.importCatalog(table.Rows[1].Cells[0].Value)
}

func (current *world) importCatalog(path string) error {
	if err := current.databaseStore.Migrate(context.Background()); err != nil {
		return err
	}

	payload, err := current.readFixture(path)
	if err != nil {
		return err
	}

	var manifest postgres.ImportManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return err
	}

	_, err = current.databaseStore.Import(context.Background(), manifest)

	return err
}

func (current *world) guestCart(table *godog.Table) error {
	store := current.databaseStore
	current.service = application.Service{Guests: store.Guests(), Carts: store.Carts(), Orders: store.Orders(), Fulfillment: store.Fulfillment(), Writer: store.Carts(), Starter: store.Carts(), Evaluator: store.Evaluator(), Purchase: store.Purchases(), Clock: current, Commands: store.Carts(), Payment: temporal.NewPaymentSimulation(true, 0)}
	if err := current.configureFault(current.rows(table)[0]); err != nil {
		return err
	}

	current.adapter = temporal.New(current.connection, current.service, store.Carts())
	current.adapter.Queue = current.id
	current.service.Processes = current.adapter
	current.guest, current.err = current.service.Open(context.Background(), "")
	if current.err != nil {
		return current.err
	}

	if err := current.startGuestCart(current.rows(table)[0]); err != nil {
		return err
	}

	return current.startCartWorker()
}

func (current *world) startCartWorker() error {
	current.id = string(current.view.Cart.ID)
	binding := application.ProcessBinding{DefinitionVersion: current.view.Cart.DefinitionVersion, RuleSetVersion: current.view.Cart.RuleSetVersion}
	if err := current.adapter.EnsureCart(context.Background(), current.view.Cart.ID, binding); err != nil {
		return err
	}

	current.runner = current.adapter.Worker()

	if err := current.runner.Start(); err != nil {
		return err
	}

	return current.sameRun()
}

func (current *world) cleanDatabase() error {
	if current.databaseStore == nil {
		return nil
	}

	current.databaseStore.Close()
	_, err := current.admin.Exec(context.Background(), "DROP SCHEMA "+pgx.Identifier{current.schema}.Sanitize()+" CASCADE")

	return errors.Join(err, current.admin.Close(context.Background()))
}

func (current *world) configureFault(row map[string]string) error {
	current.configurePurchaseFault(row["Fault"])

	if row["Fault"] == "gatedPayment" {
		current.gate = &gatedPayment{Payment: current.service.Payment, Started: make(chan struct{}), Resume: make(chan struct{})}
		current.service.Payment = current.gate
	}

	if row["PaymentDelay"] != "" {
		delay, err := time.ParseDuration(row["PaymentDelay"])
		if err != nil {
			return err
		}

		current.service.Payment = temporal.NewPaymentSimulation(true, delay)
	}

	return nil
}

func (current *world) readFixture(path string) ([]byte, error) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		return nil, err
	}

	payload, err := root.ReadFile(path)

	return payload, errors.Join(err, root.Close())
}

func (current *world) configurePurchaseFault(fault string) {
	if fault == "afterCommit" {
		current.service.Purchase = &purchaseFault{Transaction: current.databaseStore.Purchases(), Remaining: 1}
	}

	if fault == "restartAfterCommit" {
		current.commitFault = &purchaseFault{Transaction: current.databaseStore.Purchases(), Remaining: 1, Committed: make(chan struct{})}
		current.service.Purchase = current.commitFault
	}
}
