package value_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "progress", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @value", Strict: true}
	suite := godog.TestSuite{Name: "value", Options: options, ScenarioInitializer: current.initialize}
	if suite.Run() != 0 {
		t.Fatal("Werteverträge fehlgeschlagen")
	}
}

func (current *world) initialize(context *godog.ScenarioContext) {
	current.comparisonSteps(context)
	context.Step(`^ein Wert konstruiert wird$`, current.construct)
	context.Step(`^eine skalierte Zahl angefragt wird$`, current.scaled)
	context.Step(`^ein JSON-Wert geprüft wird$`, current.jsonValue)
	context.Step(`^ein technischer Fehler eingeordnet wird$`, current.technicalError)
	context.Step(`^lautet das Wertergebnis$`, current.result)
	context.Step(`^ein Verkaufskontext gewählt wird$`, current.chooseContext)
	context.Step(`^lautet das Kontextergebnis$`, current.contextResult)
}

func (current *world) construct(table *godog.Table) error {
	row := table.Rows[1].Cells
	current.canonical = "absent"
	current.err = nil
	input := row[1].Value
	if input == "empty" {
		input = ""
	}

	current.parse(row[0].Value, input)

	return nil
}

func (current *world) parse(kind, input string) {
	parsers := map[string]func(string) (string, error){"Money": parseMoney, "Decimal": parseDecimal, "Locale": parseLocale, "Currency": parseCurrency, "CountryCode": parseCountry}
	if parser, exists := parsers[kind]; exists {
		result, err := parser(input)
		current.err = err
		if err == nil {
			current.canonical = result
		}

		return
	}

	current.err = value.IdentifierSyntax(input).Validate()
	if current.err == nil {
		current.canonical = input
	}
}

func parseDecimal(input string) (string, error) {
	result, err := value.ParseDecimal(input)
	return string(result), err
}

func parseLocale(input string) (string, error) {
	result, err := value.ParseLocale(input)
	return string(result), err
}

func parseCurrency(input string) (string, error) {
	result, err := value.ParseCurrency(input)
	return string(result), err
}

func parseCountry(input string) (string, error) {
	result, err := value.ParseCountryCode(input)
	return string(result), err
}

func (current *world) result(table *godog.Table) error {
	row := table.Rows[1].Cells
	if current.err != nil {
		if current.err.Error() == "" {
			return fmt.Errorf("leere Fehlerdarstellung")
		}
	}

	if value.Code(current.err) != row[0].Value {
		return fmt.Errorf("Code %s statt %s", value.Code(current.err), row[0].Value)
	}

	if current.canonical != row[1].Value {
		return fmt.Errorf("Canonical %s statt %s", current.canonical, row[1].Value)
	}

	return nil
}

func (current *world) chooseContext(table *godog.Table) error {
	row := table.Rows[1].Cells
	current.context = value.RequestContext{Locale: value.Locale(row[0].Value), Currency: value.Currency(row[1].Value), Market: value.CountryCode(row[2].Value)}
	current.err = current.context.Validate()

	return nil
}

func (current *world) contextResult(table *godog.Table) error {
	expected := table.Rows[1].Cells[0].Value
	if value.Code(current.err) != expected {
		return fmt.Errorf("Code %s statt %s", value.Code(current.err), expected)
	}

	return nil
}

func (current *world) jsonValue(table *godog.Table) error {
	row := table.Rows[1].Cells
	current.canonical = "absent"
	if row[0].Value == "required" {
		current.err = value.RequiredJSON([]byte(row[1].Value)).Validate()
		if current.err == nil {
			current.canonical = row[1].Value
		}

		return nil
	}

	result, err := value.ParseOptionalReplacement[string]([]byte(row[1].Value))
	current.err = err
	if err == nil {
		current.canonical = result.Value
	}

	return nil
}

func (current *world) technicalError(table *godog.Table) error {
	current.err = fmt.Errorf("%s", table.Rows[1].Cells[0].Value)
	current.canonical = "absent"

	return nil
}

func (current *world) scaled(table *godog.Table) error {
	number, err := value.Decimal(table.Rows[1].Cells[0].Value).Scaled()
	current.err = err
	current.canonical = "absent"
	if err == nil {
		current.canonical = strconv.FormatInt(number, 10)
	}

	return nil
}

func parseMoney(input string) (string, error) {
	parts := strings.Split(input, ":")
	amount, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return "", err
	}

	money, err := value.NewMoney(amount, value.Currency(parts[1]))

	return strconv.FormatInt(money.MinorUnits, 10) + ":" + string(money.Currency), err
}
