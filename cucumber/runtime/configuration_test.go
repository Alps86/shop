package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"net/url"
	"time"
)

func (current *world) configurationSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine abweichende Laufzeitkonfiguration ausgeführt wird$`, current.changedConfiguration)
	ctx.Step(`^lautet das Ergebnis der Laufzeitoperation$`, current.operationObserved)
}

func (current *world) changedConfiguration(table *godog.Table) error {
	data := read(table)[0]
	config := current.config
	if err := current.changeConfiguration(&config, data); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return current.configurationOperation(ctx, config, data["Operation"])
}

func (current *world) configurationOperation(ctx context.Context, config runtime.Config, operation string) error {
	commands := map[string][]string{"content": {"content"}, "healthcheck": {"healthcheck"}, "run": nil}
	if arguments, exists := commands[operation]; exists {
		current.err = config.Execute(ctx, arguments)
		return nil
	}

	current.tryOpen(ctx, config)

	return nil
}

func (current *world) changeConfiguration(config *runtime.Config, data map[string]string) error {
	target, exists := current.configurationFields(config)[data["Field"]]
	if !exists {
		return fmt.Errorf("unbekanntes Konfigurationsfeld: %s", data["Field"])
	}

	*target = data["Value"]
	if *target == "empty" {
		*target = ""
	}

	if *target == "diagnostic-listener" {
		endpoint, err := url.Parse(current.server.URL)
		*target = endpoint.Host

		return err
	}

	return nil
}

func (current *world) configurationFields(config *runtime.Config) map[string]*string {
	fields := current.frontendFields(config)
	for key, target := range map[string]*string{"Schema": &config.Schema, "DSN": &config.DSN, "TemporalAddress": &config.TemporalAddress, "TemporalNamespace": &config.TemporalNamespace, "SearchEndpoint": &config.SearchEndpoint, "Listen": &config.Listen, "SeedFile": &config.SeedFile} {
		fields[key] = target
	}

	return fields
}

func (current *world) tryOpen(ctx context.Context, config runtime.Config) {
	app, err := runtime.Open(ctx, config)
	current.err = err
	if app != nil {
		current.err = errors.Join(err, app.Close(ctx))
	}
}

func (current *world) operationObserved(table *godog.Table) error {
	expected := read(table)[0]["Code"]
	if value.Code(current.err) != expected {
		return fmt.Errorf("Laufzeitoperation: %w (%s), erwartet %s", current.err, value.Code(current.err), expected)
	}

	return nil
}
