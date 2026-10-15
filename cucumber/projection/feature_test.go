package projection_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"strings"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @projection", Strict: true, Format: "pretty", TestingT: t}
	suite := godog.TestSuite{Name: "projection", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Reale Projektionsverträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Before(current.before)
	ctx.After(current.after)
	ctx.Step(`^PostgreSQL und OpenSearch über echten WAL-CDC verbunden sind$`, current.connect)
	ctx.Step(`^ein veröffentlichter Produktdatensatz committed wird$`, current.publish)
	ctx.Step(`^die öffentliche Suche den committed Zustand erhält$`, current.eventually)
	ctx.Step(`^Preis und Bestand atomar geändert werden$`, current.inventoryChange)
	current.registerRebuild(ctx)
	current.registerReplay(ctx)
	ctx.Step(`^eine Inventoryänderung an der tatsächlichen Outboxgrenze zurückrollt$`, current.rollbackInventory)
	ctx.Step(`^Quelle und Outbox den bestätigten Zustand behalten$`, current.rollbackObserved)
	ctx.Step(`^der echte Replikationspfad dauerhaft quittiert ist$`, current.walObserved)
	ctx.Step(`^der Suchalias genau den neuen physischen Index adressiert$`, current.aliasObserved)
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	suffix := strings.ToLower(rand.Text()[:10])
	*current = world{Alias: "projection_" + suffix}

	return ctx, nil
}

func (current *world) rows(table *godog.Table) []map[string]string {
	rows := []map[string]string{}
	for _, record := range table.Rows[1:] {
		row := map[string]string{}
		for index, cell := range record.Cells {
			row[table.Rows[0].Cells[index].Value] = cell.Value
		}

		rows = append(rows, row)
	}

	return rows
}

func (current *world) environment(key, fallback string) string {
	actual := os.Getenv(key)
	if actual != "" {
		return actual
	}

	return fallback
}

func (current *world) matches(actual map[string]string, expected map[string]string) error {
	for key, wanted := range expected {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}
