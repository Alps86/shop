package sale_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"slices"
	"strconv"
	"strings"
)

func (current *scenario) bindEvaluation(ctx *godog.ScenarioContext) {
	ctx.Step(`^besitzt die Bewertung getrennte Zuordnungen und ein Regelaggregat$`, current.aggregateResult)
	ctx.Step(`^die Bewertungsgrundlage wird explizit festgelegt$`, current.evaluationGiven)
	ctx.Step(`^eine synthetische Rabattregel gilt für Testprodukte$`, current.ruleGiven)
	ctx.Step(`^die bewertbaren Positionen lauten$`, current.itemsGiven)
	ctx.Step(`^der Warenkorb bewertet wird$`, current.evaluate)
	ctx.Step(`^lautet die Gesamtbewertung$`, current.evaluationResult)
	ctx.Step(`^eine synthetische Kombination ist bewertbar$`, current.combinationGiven)
	ctx.Step(`^eine Kombinationsaktion ausgeführt wird$`, current.combinationAction)
	ctx.Step(`^lautet die Kombinationsbewertung$`, current.combinationResult)
	ctx.Step(`^passende Rabattregeln liegen vor$`, current.rulesGiven)
	ctx.Step(`^die gemeinsame Zubehörposition bewertet wird$`, current.sharedEvaluate)
	ctx.Step(`^lautet die Rabattzuordnung$`, current.assignmentResult)
}

func (current *scenario) ruleGiven(table *godog.Table) {
	row := rows(table)[0]
	current.rules = sale.DiscountRules{{ID: value.DiscountRuleID(row["RuleID"]), Version: 1, MainProductID: value.ProductID(row["MainProductID"]), AccessoryID: value.ProductID(row["AccessoryID"]), Scope: sale.DiscountScope(row["Scope"]), Amount: money(row["Amount"], row["Currency"])}}
}

func (current *scenario) itemsGiven(table *godog.Table) {
	current.items = nil
	for _, row := range rows(table) {
		current.items = append(current.items, sale.PricedItem{SKU: value.SKU(row["SKU"]), ProductID: value.ProductID(row["ProductID"]), Quantity: number(row["Quantity"]), UnitPrice: money(row["UnitPrice"], row["Currency"])})
	}
}

func (current *scenario) evaluate(table *godog.Table) error {
	row := rows(table)[0]
	if !slices.Contains([]string{"neutral", "mobile", "home", "coffee"}, row["Profile"]) {
		return fmt.Errorf("unbekanntes Demoprofil %s", row["Profile"])
	}

	current.evaluation, current.err = current.items.Evaluate(current.rules, sale.ShippingMethod(row["ShippingMethod"]))
	if current.err == nil {
		if string(current.evaluation.Taxes[0].Rate) != row["TaxRate"] {
			return fmt.Errorf("Steuersatz %s statt %s", current.evaluation.Taxes[0].Rate, row["TaxRate"])
		}
	}

	return nil
}

func (current *scenario) evaluationResult(table *godog.Table) error {
	evaluation := current.evaluation
	var discount int64
	for _, applied := range evaluation.Discounts {
		discount += applied.Amount.MinorUnits
	}

	return check(table, map[string]string{"Subtotal": text(evaluation.Subtotal.MinorUnits), "ItemDiscountCASE": text(evaluation.Items[1].DiscountAmount()), "AggregateDiscount": text(discount), "ShippingCost": text(evaluation.ShippingCost.MinorUnits), "Total": text(evaluation.Total.MinorUnits), "Taxes": text(evaluation.Taxes[0].Amount.MinorUnits), "Currency": string(evaluation.Total.Currency)})
}

func (current *scenario) combinationGiven(table *godog.Table) {
	row := rows(table)[0]
	current.rules = sale.DiscountRules{{ID: "r-demo", Version: 1, MainProductID: "p-phone", AccessoryID: "p-case", Scope: sale.Pair, Amount: money(row["PairDiscount"], row["Currency"])}}
	current.items = sale.PricedItems{{SKU: "PHONE", ProductID: "p-phone", Quantity: number(row["MainQuantity"]), UnitPrice: money("10000", row["Currency"])}, {SKU: "CASE", ProductID: "p-case", Quantity: number(row["AccessoryQuantity"]), UnitPrice: money(row["AccessoryUnitPrice"], row["Currency"])}}
}

func (current *scenario) combinationAction(table *godog.Table) error {
	action := rows(table)[0]["Action"]
	if action == "removeMain" {
		current.items = current.items[1:]
	}

	if action == "removeAccessory" {
		current.items = current.items[:1]
	}

	current.evaluation, current.err = current.items.Evaluate(current.rules, sale.Standard)

	return current.err
}

func (current *scenario) combinationResult(table *godog.Table) error {
	var main, accessory, discount, gross int64
	for _, item := range current.evaluation.Items {
		if item.SKU == "PHONE" {
			main = item.Quantity
		}

		if item.SKU == "CASE" {
			accessory = item.Quantity
			discount = item.DiscountAmount()
			gross = item.Quantity*item.UnitPrice.MinorUnits - discount
		}
	}

	return check(table, map[string]string{"MainQuantity": text(main), "AccessoryQuantity": text(accessory), "Discount": text(discount), "AccessoryGrossAfterDiscount": text(gross)})
}

func (current *scenario) rulesGiven(table *godog.Table) {
	current.rules = nil
	for _, row := range rows(table) {
		current.rules = append(current.rules, sale.DiscountRule{ID: value.DiscountRuleID(row["RuleID"]), Version: 1, MainProductID: "p-phone", AccessoryID: "p-case", Scope: sale.DiscountScope(row["Scope"]), Amount: money(row["Amount"], "EUR")})
	}

	current.items = sale.PricedItems{{SKU: "PHONE", ProductID: "p-phone", Quantity: number(rows(table)[0]["MainQuantity"]), UnitPrice: money("10000", "EUR")}, {SKU: "CASE", ProductID: "p-case", Quantity: number(rows(table)[0]["AccessoryQuantity"]), UnitPrice: money("2000", "EUR")}}
}

func (current *scenario) sharedEvaluate() error {
	current.evaluation, current.err = current.items.Evaluate(current.rules, sale.Standard)

	return current.err
}

func (current *scenario) assignmentResult(table *godog.Table) error {
	for _, row := range rows(table) {
		amount, units := current.assignedRule(row["RuleID"])
		if err := compareAssignment(row, amount, units); err != nil {
			return err
		}
	}

	return nil
}

func (current *scenario) assignedRule(id string) (int64, int64) {
	var amount, units int64
	for _, discount := range current.evaluation.Discounts {
		if string(discount.RuleID) == id {
			amount += discount.Amount.MinorUnits
			units++
		}
	}

	return amount, units
}

func compareAssignment(row map[string]string, amount, units int64) error {
	if row["AppliedAmount"] != text(amount) {
		return value.NewError("UnexpectedAmount", row["RuleID"])
	}

	if row["AssignedAccessoryUnits"] != text(units) {
		return value.NewError("UnexpectedUnits", row["RuleID"])
	}

	return nil
}

func (current *scenario) evaluationGiven(table *godog.Table) {
	row := rows(table)[0]
	current.items = sale.PricedItems{{SKU: value.SKU(input(row["SKU"])), ProductID: value.ProductID(row["ProductID"]), Quantity: number(row["Quantity"]), UnitPrice: money(row["UnitPrice"], row["Currency"])}}
	current.rules = sale.DiscountRules{{ID: value.DiscountRuleID(input(row["RuleID"])), Version: unsigned(row["RuleVersion"]), MainProductID: value.ProductID(input(row["MainProductID"])), AccessoryID: value.ProductID(input(row["AccessoryID"])), Amount: money(row["RuleAmount"], row["RuleCurrency"]), Scope: sale.DiscountScope(row["Scope"])}}
}

func (current *scenario) aggregateResult(table *godog.Table) error {
	discount := current.evaluation.Discounts[0]
	skus := []string{}
	for _, sku := range discount.ItemSKUs {
		skus = append(skus, string(sku))
	}

	return check(table, map[string]string{"RuleID": string(discount.RuleID), "AggregateCount": strconv.Itoa(len(current.evaluation.Discounts)), "AggregateAmount": text(discount.Amount.MinorUnits), "AggregateSKUs": strings.Join(skus, ","), "DiscountCASE": text(current.evaluation.Items[1].DiscountAmount()), "DiscountCASE2": text(current.evaluation.Items[2].DiscountAmount()), "Total": text(current.evaluation.Total.MinorUnits)})
}
