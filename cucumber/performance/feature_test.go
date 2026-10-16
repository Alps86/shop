package performance_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"os/exec"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @demo-performance", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "performance", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Echte Performanceabnahme fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.After(current.close)
	ctx.Step(`^der dokumentierte Messdatensatz ist geladen$`, current.open)
	ctx.Step(`^die vereinbarten Anzeigerouten gemessen werden$`, current.measure)
	ctx.Step(`^die direkten Go-Grenzen getrennt gemessen werden$`, current.direct)
	ctx.Step(`^muss der Messbericht die Aussagegrenze enthalten$`, current.observe)
}

func (current *world) close(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.temporal != nil {
		current.temporal.Close()
	}

	if current.store != nil {
		current.store.Close()
	}

	return ctx, nil
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

func configureTool(command *exec.Cmd, variable string) {
	if configured := os.Getenv(variable); configured != "" {
		command.Path = configured
		command.Err = nil
	}
}

func require(err error, message string) error {
	if err != nil {
		return errors.Join(fmt.Errorf("%s", message), err)
	}

	return nil
}
