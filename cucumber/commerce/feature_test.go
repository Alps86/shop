package commerce_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @commerce-runtime", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "commerce-runtime", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Commerceintegration fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.After(current.cleanup)
	ctx.Step(`^ein nicht gespeicherter Warenkorb direkt beim Prozessport angefragt wird$`, current.missingCart)
	ctx.Step(`^die vollständige Commerce-Laufzeit mit echten Diensten läuft$`, current.open)
	ctx.Step(`^ein Gast unmittelbar nach dem Start seinen Warenkorb ändert$`, current.immediateChange)
	ctx.Step(`^lautet die unmittelbare Warenkorbantwort$`, current.changeObserved)
	ctx.Step(`^nur OpenSearch ist nicht erreichbar$`, current.searchDown)
	ctx.Step(`^die öffentlichen Ports aufgerufen werden$`, current.ports)
	ctx.Step(`^lauten die Portantworten$`, current.responses)
}

func read(table *godog.Table) []row {
	rows := []row{}
	for _, record := range table.Rows[1:] {
		data := row{}
		for index, cell := range record.Cells {
			data[table.Rows[0].Cells[index].Value] = cell.Value
		}

		rows = append(rows, data)
	}

	return rows
}

func (data row) matches(actual map[string]string) error {
	for key, want := range data {
		if actual[key] != want {
			return fmt.Errorf("%s: %q statt %q", key, actual[key], want)
		}
	}

	return nil
}

func (current *world) searchDown(table *godog.Table) error {
	return current.open(table)
}

func (current *world) responses(table *godog.Table) error {
	for _, data := range read(table) {
		actual := current.answers[data["Operation"]]
		if actual == nil {
			actual = map[string]string{}
		}

		actual["Code"] = current.codes[data["Operation"]]
		delete(data, "Operation")
		if err := data.matches(actual); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) changeObserved(table *godog.Table) error {
	actual := map[string]string{"Code": current.codes["Change"]}
	if len(current.view.Cart.Items) > 0 {
		actual["SKU"] = string(current.view.Cart.Items[0].SKU)
	}

	return read(table)[0].matches(actual)
}

func (current *world) ports(table *godog.Table) error {
	for _, data := range read(table) {
		if err := current.port(context.Background(), data); err != nil {
			return err
		}
	}

	return nil
}
