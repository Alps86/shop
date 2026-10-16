package demo_test

import (
	"context"
	"errors"
	"github.com/cucumber/godog"
	"net/http"
	"net/url"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @demo-http", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "demo-http", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Öffentliche Demoabnahme fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Before(current.before)
	ctx.After(current.after)
	ctx.Step(`^die echte öffentliche Demolaufzeit ist bereit$`, current.open)
	ctx.Step(`^ein eigener HTTPgast wird verwendet$`, current.selectClient)
	ctx.Step(`^der Gast ruft die öffentliche Ansicht auf$`, current.request)
	ctx.Step(`^die öffentliche Antwort enthält$`, current.observe)
	ctx.Step(`^der Gast sendet das sichtbare Formular$`, current.submit)
	ctx.Step(`^die aktuelle Formularantwort wird begrenzt abgewartet$`, current.await)
	ctx.Step(`^die Formularanfrage wird für einen Wiederholungsversuch gespeichert$`, current.rememberForm)
	ctx.Step(`^der unveränderte Abschluss wird erneut gesendet$`, current.replay)
	ctx.Step(`^die Bestellung ist über echte Kaufports beobachtbar$`, current.orderObserved)
	ctx.Step(`^der persistierte Gastwarenkorb lautet$`, current.cartObserved)
	current.registerOperations(ctx)
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	*current = world{clients: map[string]*http.Client{}, runs: map[string]string{}, prepared: map[string]url.Values{}}
	return ctx, nil
}

func (current *world) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.deployment == nil {
		return ctx, nil
	}

	cleanup := current.cleanupSearchImport()

	return ctx, errors.Join(cleanup, current.deployment.stopApp(), current.stopWorkflows(), current.deployment.close())
}

func read(table *godog.Table) []map[string]string {
	result := []map[string]string{}
	for _, row := range table.Rows[1:] {
		data := map[string]string{}
		for index, cell := range row.Cells {
			data[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, data)
	}

	return result
}
