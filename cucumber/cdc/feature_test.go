package cdc_test

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"os"
	"strings"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @pgoutput", Format: "pretty", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "cdc", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("pgoutput-Verträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	current.streamSteps(ctx)
	current.validationSteps(ctx)
	current.protocolSteps(ctx)
	current.envelopeSteps(ctx)
	current.databaseFaultSteps(ctx)
	ctx.Step(`^eine eigene Testdatenbank ohne folgende Berechtigung besteht$`, current.restrictedDatabase)
	ctx.Before(current.before)
	ctx.After(current.after)
	ctx.Step(`^ein exportierter PostgreSQL-Snapshot mit eigenem pgoutput-Slot besteht$`, current.snapshotReady)
	ctx.Step(`^danach ein vollständiges Löschereignis committed wird$`, current.commitEvent)
	ctx.Step(`^der WAL-Nachlauf bis zur aktuellen Grenze erfolgt$`, current.catchup)
	ctx.Step(`^lautet die tatsächliche Projektionsbeobachtung$`, current.observed)
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	*current = world{available: true}
	suffix := strings.ToLower(rand.Text()[:10])
	current.config = cdc.Config{DSN: os.Getenv("SHOP_TEST_POSTGRES_DSN"), Schema: "cdc_test_" + suffix, Publication: "cdc_pub_" + suffix, Slot: "cdc_slot_" + suffix, Target: current}
	admin, err := pgx.Connect(ctx, current.config.DSN)
	if err != nil {
		return ctx, err
	}

	current.admin = admin
	store, err := postgres.Open(ctx, postgres.Config{DSN: current.config.DSN, Schema: current.config.Schema, Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}})
	if err != nil {
		return ctx, err
	}

	current.store = store
	current.config.Store = store
	current.source = cdc.NewReindexSource(current.config)

	return ctx, store.Migrate(ctx)
}

func (current *world) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.consumer != nil {
		if err := current.closeStream(); err != nil {
			return ctx, err
		}
	}

	if current.store != nil {
		current.store.Close()
	}

	if current.admin == nil {
		return ctx, nil
	}

	return ctx, current.cleanup(ctx)
}

func (current *world) cleanup(ctx context.Context) error {
	_, slotErr := current.admin.Exec(ctx, "SELECT pg_drop_replication_slot(slot_name) FROM pg_replication_slots WHERE slot_name=$1", current.config.Slot)
	_, publicationErr := current.admin.Exec(ctx, "DROP PUBLICATION IF EXISTS "+pgx.Identifier{current.config.Publication}.Sanitize())
	_, schemaErr := current.admin.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{current.config.Schema}.Sanitize()+" CASCADE")

	databaseErr := current.dropForeignDatabase(ctx)
	roleErr := current.dropRestrictedRole(ctx)

	return errors.Join(slotErr, publicationErr, schemaErr, databaseErr, roleErr, current.admin.Close(ctx))
}

func (current *world) observed(input *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	actual := fmt.Sprint(len(current.events))
	if actual != input.Rows[1].Cells[0].Value {
		return fmt.Errorf("Events: %s", actual)
	}

	if string(current.events[0].ProductID) != input.Rows[1].Cells[1].Value {
		return fmt.Errorf("ProductID: %s", current.events[0].ProductID)
	}

	return nil
}

func (current *world) dropForeignDatabase(ctx context.Context) error {
	if current.foreignDatabase == "" {
		return nil
	}

	_, err := current.admin.Exec(ctx, "DROP DATABASE "+pgx.Identifier{current.foreignDatabase}.Sanitize())

	return err
}
