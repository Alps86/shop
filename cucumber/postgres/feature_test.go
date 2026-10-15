package postgres_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @postgres", Format: "pretty", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "postgres", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("PostgreSQL-Verträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	current.purchaseSteps(ctx)
	current.bundleSteps(ctx)
	current.changeSteps(ctx)
	current.raceSteps(ctx)
	current.readModelSteps(ctx)
	current.processSteps(ctx)
	current.faultSteps(ctx)
	current.cartSteps(ctx)
	current.purchaseBoundarySteps(ctx)
	current.cancellationSteps(ctx)
	ctx.Step(`^die technische Speichergrenze mit folgender Konfiguration aufgerufen wird$`, current.construction)
	ctx.Before(current.before)
	ctx.After(current.after)
	ctx.Step(`^ein kohärenter Demodatensatz in echter PostgreSQL vorbereitet ist$`, current.prepare)
	ctx.Step(`^der Datensatz veröffentlicht wird$`, current.publish)
	ctx.Step(`^die maßgebliche Katalogsicht gelesen wird$`, current.catalog)
	ctx.Step(`^lautet das persistente Importergebnis$`, current.importObserved)
}

func (current *world) readModelSteps(ctx *godog.ScenarioContext) {
	current.readSteps(ctx)
	current.importSteps(ctx)
	current.publicationSteps(ctx)
	current.accessorySteps(ctx)
	current.readFaultSteps(ctx)
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	*current = world{schema: "shop_test_" + strings.ToLower(rand.Text()[:10]), observed: map[string]string{}}
	dsn := os.Getenv("SHOP_TEST_POSTGRES_DSN")
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return ctx, err
	}

	current.admin = connection
	store, err := postgres.Open(ctx, postgres.Config{DSN: dsn, Schema: current.schema, Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}})
	if err != nil {
		return ctx, err
	}

	current.store = store

	return ctx, store.Migrate(ctx)
}

func (current *world) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.store != nil {
		current.store.Close()
	}

	if current.admin == nil {
		return ctx, nil
	}

	_, err := current.admin.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{current.schema}.Sanitize()+" CASCADE")
	closeErr := current.admin.Close(ctx)
	if err != nil {
		return ctx, err
	}

	return ctx, closeErr
}

func table(table *godog.Table) map[string]string {
	result := map[string]string{}
	for index, cell := range table.Rows[1].Cells {
		result[table.Rows[0].Cells[index].Value] = cell.Value
	}

	return result
}

func check(expected, actual map[string]string) error {
	for key, wanted := range expected {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %q statt %q", key, actual[key], wanted)
		}
	}

	return nil
}
