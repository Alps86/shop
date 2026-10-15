package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @http", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "http", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("HTTP-Verträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		*current = world{}
		return ctx, nil
	})
	ctx.Step(`^explizite HTTPdaten liegen vor$`, current.prepare)
	ctx.Step(`^HTTPportdaten werden ersetzt$`, current.override)
	ctx.Step(`^die HTTPkonfiguration verwendet$`, current.configure)
	ctx.Step(`^der HTTPadapter wird initialisiert$`, current.construct)
	ctx.Step(`^die Adapterinitialisierung meldet$`, current.initializationResult)
	current.registerComposition(ctx)
	ctx.Step(`^der Shop liefert eine HTTPansicht$`, current.request)
	ctx.Step(`^die HTTPantwort enthält die sichtbaren Vertragswerte$`, current.expect)
	ctx.Step(`^der verwendete Applicationport erhält$`, current.expectPort)
}

func (current *world) prepare(table *godog.Table) error {
	fields := read(table)[0]
	data, err := os.ReadFile("../../" + fields["Fixture"])
	if err != nil {
		return err
	}

	current.fault = fields["Fault"]

	return json.Unmarshal(data, &current.fixture)
}

func read(table *godog.Table) rows {
	result := rows{}
	for _, row := range table.Rows[1:] {
		fields := map[string]string{}
		for index, cell := range row.Cells {
			fields[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, fields)
	}

	return result
}

func (expected rows) matches(actual map[string]string) error {
	for key, wanted := range expected[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %q statt %q", key, actual[key], wanted)
		}
	}

	return nil
}

func (current *world) override(table *godog.Table) error {
	for _, fields := range read(table) {
		if err := json.Unmarshal([]byte(fields["JSON"]), &current.fixture); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) registerComposition(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Standardengines werden mit Hugo gestartet$`, current.startComposition)
	ctx.Step(`^die zusammengesetzte HTTPansicht wird angefordert$`, current.composedRequest)
	ctx.After(current.stopComposition)
	ctx.After(current.stopCache)
	ctx.Step(`^die Cacheinvalidierung wird eingerichtet$`, current.prepareCache)
	ctx.Step(`^die SKUcacheinvalidierung ausgeführt wird$`, current.invalidateCache)
	ctx.Step(`^die Cachegrenze beobachtet$`, current.cacheResult)
	ctx.Step(`^der komponierte Cache invalidiert wird$`, current.invalidateComposition)
	ctx.Step(`^die Standardenginelogs enthalten keine Fähigkeiten$`, current.inspectLogs)
}
