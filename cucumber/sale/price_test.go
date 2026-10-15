package sale_test

import (
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/cucumber/godog"
)

func (current *scenario) bindPrice(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Bewertungsänderung fachlich verglichen wird$`, current.comparePrices)
	ctx.Step(`^lautet der Preisvergleich$`, current.priceComparison)
}

func (current *scenario) comparePrices(table *godog.Table) {
	row := rows(table)[0]
	changed := current.evaluation.Clone()
	if !specialPriceChange(&changed, row) {
		mutatePrice(&changed, row)
	}

	current.matched = current.evaluation.MatchesPrices(changed)
}

func specialPriceChange(evaluation *sale.CartEvaluation, row map[string]string) bool {
	operations := map[string]func(){
		"none": func() {
		},
		"Items": func() {
			evaluation.Items = nil
		},
		"InventoryRevision": func() {
			evaluation.InventoryRevisions["PHONE"] = unsigned(row["Value"])
		},
		"reverseItems": func() {
			slices.Reverse(evaluation.Items)
		},
	}
	operation, exists := operations[row["Field"]]
	if exists {
		operation()
	}

	return exists
}

func mutatePrice(evaluation *sale.CartEvaluation, row map[string]string) {
	target := priceField(reflect.ValueOf(evaluation).Elem(), strings.Split(row["Field"], "."))
	operations := map[reflect.Kind]func(){
		reflect.String: func() {
			target.SetString(row["Value"])
		},
		reflect.Int64: func() {
			target.SetInt(number(row["Value"]))
		},
		reflect.Uint64: func() {
			target.SetUint(unsigned(row["Value"]))
		},
	}
	operations[target.Kind()]()
}

func priceField(root reflect.Value, path []string) reflect.Value {
	if len(path) == 0 {
		return root
	}

	if root.Kind() == reflect.Slice {
		index, err := strconv.Atoi(path[0])
		if err != nil {
			panic(err)
		}

		return priceField(root.Index(index), path[1:])
	}

	return priceField(root.FieldByName(path[0]), path[1:])
}

func (current *scenario) priceComparison(table *godog.Table) error {
	return check(table, map[string]string{"MatchesPrices": strconv.FormatBool(current.matched)})
}
