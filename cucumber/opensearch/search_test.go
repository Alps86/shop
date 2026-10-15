package opensearch_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) searchSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die kuratierten Suchdefinitionen stehen bereit$`, current.definitions)
	ctx.Step(`^die öffentliche Suche ausgeführt wird$`, current.publicSearch)
	ctx.Step(`^lautet das Suchergebnis$`, current.searchObserved)
	ctx.Step(`^lauten die kontextuellen Attributzählungen$`, current.facetsObserved)
	ctx.Step(`^lautet die Modellkarte$`, current.cardObserved)
}

func (current *world) definitions(table *godog.Table) {
	current.config.Schema = application.SearchSchema{Attributes: map[value.AttributeCode]catalog.AttributeDefinition{}}
	category := catalog.Category{ID: current.template.PrimaryCategoryID}
	for _, data := range read(table) {
		definition := data.definition()
		current.config.Schema.Attributes[definition.Code] = definition
		category.FilterAttributes = append(category.FilterAttributes, definition.Code)
		if data["Top"] == "true" {
			category.TopAttributes = append(category.TopAttributes, definition.Code)
		}
	}

	current.config.Categories = map[value.CategoryID]catalog.Category{category.ID: category}
	current.adapter = opensearch.New(current.config)
}

func (data row) definition() catalog.AttributeDefinition {
	definition := catalog.AttributeDefinition{Code: value.AttributeCode(data["Code"]), Type: catalog.AttributeType(data["Type"]), Level: "variant", Unit: data["Unit"], Minimum: data.optionalNumber("Minimum"), Maximum: data.optionalNumber("Maximum")}
	for _, option := range strings.Split(data["Options"], ",") {
		if option != "[]" {
			definition.Options = append(definition.Options, catalog.AttributeOption{Code: value.OptionCode(option)})
		}
	}

	return definition
}

func (data row) optionalNumber(key string) value.Optional[value.Decimal] {
	input := data[key]
	if input == "absent" {
		return value.Optional[value.Decimal]{}
	}

	return value.Optional[value.Decimal]{Present: true, Value: value.Decimal(input)}
}

func (current *world) publicSearch(table *godog.Table) error {
	data := read(table)[0]
	page, err := strconv.Atoi(data["Page"])
	if err != nil {
		return err
	}

	current.query = application.SearchQuery{Page: page, Sort: application.Sort(data["Sort"]), Context: current.config.Context, Filters: data.filters()}
	if err := current.price(data); err != nil {
		return err
	}

	data.rangeFilter(&current.query.Filters)
	current.result, current.err = current.adapter.Search(context.Background(), current.query)

	return nil
}

func (data row) filters() application.FilterSet {
	filters := application.FilterSet{Options: map[value.AttributeCode][]value.OptionCode{}, Ranges: map[value.AttributeCode]application.NumericRange{}}
	if data.filterPresent("Brand") {
		filters.Brands = strings.Split(data["Brand"], ",")
	}

	if data.filterPresent("Color") {
		filters.Options["color"] = []value.OptionCode{value.OptionCode(data["Color"])}
	}

	if data.filterPresent("StorageGB") {
		bound := value.Optional[value.Decimal]{Present: true, Value: value.Decimal(data["StorageGB"])}
		filters.Ranges["storageCapacity"] = application.NumericRange{Minimum: bound, Maximum: bound}
	}

	return filters
}

func (data row) filterPresent(key string) bool {
	return data[key] != "" && data[key] != "any" && data[key] != "[]"
}

func (current *world) price(data row) error {
	if !data.filterPresent("MaxPrice") {
		return nil
	}

	price, err := strconv.ParseInt(data["MaxPrice"], 10, 64)
	if err != nil {
		return err
	}

	current.query.Filters.Price = value.Optional[application.MoneyRange]{Present: true, Value: application.MoneyRange{Maximum: value.Optional[value.Money]{Present: true, Value: value.Money{MinorUnits: price, Currency: current.config.Context.Currency}}}}

	current.minimumPrice(data)

	return nil
}

func (current *world) searchObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	actual := map[string]string{"TotalProducts": strconv.FormatInt(current.result.TotalProducts, 10), "ProductIDs": "[]", "SelectedZeroChip": "none", "BrandAlpha": "0", "BrandBeta": "0", "BrandGamma": "0"}
	ids := []string{}
	for _, card := range current.result.Cards {
		ids = append(ids, string(card.ProductID))
	}

	if len(ids) > 0 {
		actual["ProductIDs"] = strings.Join(ids, ",")
	}

	current.brandCounts(actual)

	return read(table)[0].matches(actual)
}

func (current *world) brandCounts(actual map[string]string) {
	for _, facet := range current.result.Facets {
		if facet.Code != "brand" {
			continue
		}

		for _, item := range facet.Values {
			actual["Brand"+item.Value] = strconv.FormatInt(item.ProductCount, 10)
			if current.selectedZero(item) {
				actual["SelectedZeroChip"] = item.Value
			}
		}
	}
}

func (current *world) selectedZero(item application.FacetValue) bool {
	return item.Selected && item.ProductCount == 0
}

func (current *world) facetsObserved(table *godog.Table) error {
	for _, expected := range read(table) {
		count := current.facetCount(expected)
		if strconv.FormatInt(count, 10) != expected["ProductCount"] {
			return fmt.Errorf("Facette %s %s: %d statt %s", expected["Facet"], expected["Value"], count, expected["ProductCount"])
		}
	}

	return current.err
}

func (current *world) facetCount(expected row) int64 {
	for _, facet := range current.result.Facets {
		if facet.Code == expected["Facet"] {
			return expected.valueCount(facet.Values)
		}
	}

	return 0
}

func (expected row) valueCount(values []application.FacetValue) int64 {
	for _, item := range values {
		if item.Value == expected["Value"] {
			return item.ProductCount
		}
	}

	return 0
}

func (current *world) cardObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	if len(current.result.Cards) == 0 {
		return fmt.Errorf("keine Modellkarte")
	}

	card := current.result.Cards[0]

	return read(table)[0].matches(map[string]string{"ProductID": string(card.ProductID), "ChosenSKU": string(card.SKU), "VariantCount": strconv.Itoa(card.VariantCount), "CardCount": strconv.Itoa(len(current.result.Cards))})
}

func (expected row) matches(actual map[string]string) error {
	for field, want := range expected {
		if actual[field] != want {
			return fmt.Errorf("%s: %q statt %q", field, actual[field], want)
		}
	}

	return nil
}

func (data row) rangeFilter(filters *application.FilterSet) {
	if data["RangeCode"] == "" {
		return
	}

	filters.Ranges[value.AttributeCode(data["RangeCode"])] = application.NumericRange{Minimum: data.optionalNumber("RangeMin"), Maximum: data.optionalNumber("RangeMax")}
}

func (current *world) minimumPrice(data row) {
	if data["MinPrice"] == "" {
		return
	}

	amount, err := strconv.ParseInt(data["MinPrice"], 10, 64)
	if err != nil {
		current.err = err
		return
	}

	current.query.Filters.Price.Value.Minimum = value.Optional[value.Money]{Present: true, Value: value.Money{MinorUnits: amount, Currency: current.config.Context.Currency}}
}
