package sale_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"testing"
	"time"
)

func TestFeatures(t *testing.T) {
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @sale", Strict: true}
	status := (godog.TestSuite{Name: "sale", Options: options, ScenarioInitializer: initialize}).Run()
	if status != 0 {
		t.Fatal(status)
	}
}

func initialize(ctx *godog.ScenarioContext) {
	current := &scenario{}
	current.bindInventory(ctx)
	current.bindCart(ctx)
	current.bindCheckout(ctx)
	current.bindFulfillment(ctx)
	current.bindPrice(ctx)
	current.bindEvaluation(ctx)
	current.bindUser(ctx)
}

func rows(table *godog.Table) []map[string]string {
	result := []map[string]string{}
	for _, row := range table.Rows[1:] {
		entry := map[string]string{}
		for index, cell := range row.Cells {
			entry[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, entry)
	}

	return result
}

func number(raw string) int64 {
	result, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return result
}

func check(table *godog.Table, actual map[string]string) error {
	for _, expected := range rows(table) {
		if err := checkFields(expected, actual); err != nil {
			return err
		}
	}

	return nil
}

func checkFields(expected, actual map[string]string) error {
	for field, want := range expected {
		if actual[field] != want {
			return fmt.Errorf("%s: erwartet %s, erhalten %s", field, want, actual[field])
		}
	}

	return nil
}

func instant() time.Time {
	return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
}

func money(amount string, currency string) value.Money {
	return value.Money{MinorUnits: number(amount), Currency: value.Currency(currency)}
}

func text(amount int64) string {
	return strconv.FormatInt(amount, 10)
}

func (current *scenario) bindInventory(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Inventory fachlich angelegt wird$`, current.inventoryCreate)
	ctx.Step(`^lautet der fachliche Fehlercode$`, current.errorResult)
	ctx.Step(`^ein Inventory enthält die Ausgangswerte$`, current.inventoryGiven)
	ctx.Step(`^eine Inventoryoperation ausgeführt wird$`, current.inventoryOperate)
	ctx.Step(`^lautet das Inventoryergebnis$`, current.inventoryResult)
	ctx.Step(`^eine Geldoperation ausgeführt wird$`, current.moneyOperate)
	ctx.Step(`^lautet das Geldergebnis$`, current.moneyResult)
}

func (current *scenario) inventoryGiven(table *godog.Table) {
	row := rows(table)[0]
	current.inventory = sale.Inventory{SKU: value.SKU(row["SKU"]), Price: money(row["Price"], row["Currency"]), StockQuantity: number(row["StockQuantity"]), Revision: unsigned(row["Revision"])}
	current.inventory.Reservations = []sale.Reservation{{ID: "r-1", OrderID: "o-1", Quantity: number(row["Reserved"])}}
}

func (current *scenario) inventoryOperate(table *godog.Table) {
	row := rows(table)[0]
	operations := map[string]func(map[string]string){
		"reserve":     current.reserve,
		"setStock":    current.stock,
		"setPrice":    current.price,
		"ship":        current.ship,
		"shipUnknown": current.shipUnknown,
		"validate":    current.validateInventory,
	}

	operations[row["Operation"]](row)
}

func (current *scenario) reserve(row map[string]string) {
	current.inventory, current.err = current.inventory.Reserve("o-2", number(row["Amount"]), unsigned(row["ExpectedRevision"]))
}

func (current *scenario) stock(row map[string]string) {
	current.inventory, current.err = current.inventory.SetStock(number(row["Amount"]), unsigned(row["ExpectedRevision"]))
}

func (current *scenario) price(row map[string]string) {
	current.inventory, current.err = current.inventory.SetPrice(money(row["Amount"], row["Currency"]), unsigned(row["ExpectedRevision"]))
}

func (current *scenario) ship(row map[string]string) {
	current.inventory, current.err = current.inventory.Ship("o-1", number(row["Amount"]), unsigned(row["ExpectedRevision"]))
}

func (current *scenario) shipUnknown(row map[string]string) {
	current.inventory, current.err = current.inventory.Ship("o-unknown", number(row["Amount"]), unsigned(row["ExpectedRevision"]))
}

func (current *scenario) validateInventory(_ map[string]string) {
	current.err = current.inventory.Validate()
}

func (current *scenario) inventoryResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err), "StockQuantity": text(current.inventory.StockQuantity), "Reserved": text(current.inventory.ReservedQuantity()), "AvailableQuantity": text(current.inventory.AvailableQuantity())})
}

func (current *scenario) moneyOperate(table *godog.Table) error {
	row := rows(table)[0]
	left := money(row["Left"], row["CurrencyLeft"])
	if row["Operation"] == "add" {
		current.money, current.err = left.Add(money(row["Right"], row["CurrencyRight"]))
		return nil
	}

	if row["Operation"] == "multiply" {
		current.money, current.err = left.Multiply(number(row["Right"]))
		return nil
	}

	return fmt.Errorf("unbekannte Operation %s", row["Operation"])
}

func (current *scenario) moneyResult(table *godog.Table) error {
	amount := "absent"
	if current.err == nil {
		amount = text(current.money.MinorUnits)
	}

	return check(table, map[string]string{"Code": value.Code(current.err), "MinorUnits": amount})
}

func (current *scenario) inventoryCreate(table *godog.Table) {
	row := rows(table)[0]
	current.inventory, current.err = sale.NewInventory(value.SKU(input(row["SKU"])), money(row["Price"], row["Currency"]), number(row["StockQuantity"]))
}

func (current *scenario) errorResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err)})
}

func input(raw string) string {
	if raw == "empty" {
		return ""
	}

	return raw
}

func unsigned(raw string) uint64 {
	result, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return result
}
