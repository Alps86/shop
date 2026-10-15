package catalog_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "progress", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @catalog", Strict: true}
	suite := godog.TestSuite{Name: "catalog", Options: options, ScenarioInitializer: current.initialize}
	if suite.Run() != 0 {
		t.Fatal("Katalogverträge fehlgeschlagen")
	}
}

func (current *world) initialize(context *godog.ScenarioContext) {
	current.attributeSteps(context)
	current.productSteps(context)
	current.complianceSteps(context)
	current.roleReadSteps(context)
}

func tableRow(table *godog.Table) row {
	result := row{}
	for index, cell := range table.Rows[0].Cells {
		result[cell.Value] = table.Rows[1].Cells[index].Value
	}

	return result
}

func (data row) truth(key string) bool {
	return data[key] == "true"
}

func list(input string) []string {
	if input == "[]" {
		return nil
	}

	if input == "" {
		return nil
	}

	return strings.Split(input, ",")
}

func decimal(input string) value.Decimal {
	return value.Decimal(input)
}

func integer(input string) (int, error) {
	return strconv.Atoi(input)
}

func timestamp(input string) (time.Time, error) {
	return time.Parse(time.RFC3339, input)
}

func (current *world) checkCode(data row) error {
	actual := value.Code(current.err)
	if actual != data["Code"] {
		return fmt.Errorf("Code %s statt %s (%w)", actual, data["Code"], current.err)
	}

	return nil
}

func equal(actual, expected string) error {
	if actual != expected {
		return fmt.Errorf("%s statt %s", actual, expected)
	}

	return nil
}
