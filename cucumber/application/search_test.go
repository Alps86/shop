package application_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerSearch(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Suche mit einem Verkaufskontext angefragt wird$`, current.searchContext)
	ctx.Step(`^ein strukturierter Suchrequest angefragt wird$`, current.structuredSearch)
	ctx.Step(`^ein Suchrequest validiert wird$`, current.validateSearch)
	ctx.Step(`^lautet das Suchvalidierungsergebnis$`, current.searchValidated)
}

func (current *world) validateSearch(table *godog.Table) error {
	row := read(table)[0]
	current.searchQuery = newSearchQuery()
	modifiers := map[string]func(string) error{"Page": current.searchPage, "Sort": current.searchSort, "Attribute": current.searchAttribute, "FilterOperator": current.searchOperator}
	modifier, exists := modifiers[row["Field"]]
	if !exists {
		return fmt.Errorf("Unbekanntes Suchfeld: %s", row["Field"])
	}

	if err := modifier(row["Value"]); err != nil {
		return err
	}

	current.searchResult, current.err = (application.SearchService{Searcher: current}).Search(context.Background(), current.searchQuery, newSearchSchema())

	return nil
}

func (current *world) searchPage(input string) error {
	page, err := strconv.Atoi(input)
	current.searchQuery.Page = page

	return err
}

func (current *world) searchSort(input string) error {
	current.searchQuery.Sort = application.Sort(input)
	return nil
}

func (current *world) searchAttribute(input string) error {
	current.searchQuery.Filters.Options = map[value.AttributeCode][]value.OptionCode{value.AttributeCode(input): {"red"}}
	return nil
}

func (current *world) searchOperator(input string) error {
	current.searchQuery.Filters.Operators = []string{input}
	return nil
}

func (current *world) searchValidated(table *godog.Table) error {
	return read(table).matches(map[string]string{"Code": value.Code(current.err), "PageSize": "12"})
}

func (current *world) Search(_ context.Context, query application.SearchQuery) (application.SearchResult, error) {
	if current.fault == "search" {
		return application.SearchResult{}, value.NewError("DependencyUnavailable", "Search")
	}

	return application.SearchResult{Page: query.Page, PageSize: 12}, nil
}

func (current *world) structuredSearch(table *godog.Table) error {
	row := read(table)[0]
	current.fault = row["Fault"]
	query := newSearchQuery()
	if err := current.setFilters(&query, row); err != nil {
		return err
	}

	current.searchResult, current.err = (application.SearchService{Searcher: current}).Search(context.Background(), query, newSearchSchema())

	return nil
}

func newSearchSchema() application.SearchSchema {
	return application.SearchSchema{Attributes: map[value.AttributeCode]catalog.AttributeDefinition{"color": {Type: "select", Options: []catalog.AttributeOption{{Code: "red"}, {Code: "blue"}}}, "storageCapacity": {Type: "number"}}}
}

func (current *world) setFilters(query *application.SearchQuery, row map[string]string) error {
	if row["Type"] == "options" {
		query.Filters.Options = map[value.AttributeCode][]value.OptionCode{value.AttributeCode(row["Attribute"]): optionCodes(row["Options"])}
	}

	if row["Type"] == "range" {
		query.Filters.Ranges = map[value.AttributeCode]application.NumericRange{value.AttributeCode(row["Attribute"]): {Minimum: decimalBound(row["Minimum"]), Maximum: decimalBound(row["Maximum"])}}
	}

	if row["Type"] == "price" {
		return current.priceFilter(query, row)
	}

	if row["Type"] == "dirtyPrice" {
		query.Filters.Price.Value.Minimum.Present = true
	}

	return nil
}

func optionCodes(input string) []value.OptionCode {
	result := []value.OptionCode{}
	if input == "[]" {
		return result
	}

	for _, code := range strings.Split(input, ",") {
		result = append(result, value.OptionCode(code))
	}

	return result
}

func decimalBound(input string) value.Optional[value.Decimal] {
	if input == "absent" {
		return value.Optional[value.Decimal]{}
	}

	if input == "dirty" {
		return value.Optional[value.Decimal]{Value: "1"}
	}

	return value.Optional[value.Decimal]{Value: value.Decimal(input), Present: true}
}

func (current *world) priceFilter(query *application.SearchQuery, row map[string]string) error {
	minimum, err := moneyBound(row["Minimum"], row["CurrencyMinimum"])
	if err != nil {
		return err
	}

	maximum, err := moneyBound(row["Maximum"], row["CurrencyMaximum"])
	if err != nil {
		return err
	}

	query.Filters.Price = value.Optional[application.MoneyRange]{Present: true, Value: application.MoneyRange{Minimum: minimum, Maximum: maximum}}

	return nil
}

func moneyBound(input, currency string) (value.Optional[value.Money], error) {
	if input == "absent" {
		return value.Optional[value.Money]{}, nil
	}

	if input == "dirty" {
		return value.Optional[value.Money]{Value: value.Money{Currency: value.Currency(currency)}}, nil
	}

	amount, err := strconv.ParseInt(input, 10, 64)

	return value.Optional[value.Money]{Present: true, Value: value.Money{MinorUnits: amount, Currency: value.Currency(currency)}}, err
}

func newSearchQuery() application.SearchQuery {
	return application.SearchQuery{Page: 1, Sort: "relevance", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
}

func (current *world) searchContext(table *godog.Table) {
	row := read(table)[0]
	query := newSearchQuery()
	query.Context = value.RequestContext{Locale: value.Locale(row["Locale"]), Currency: value.Currency(row["Currency"]), Market: value.CountryCode(row["Market"])}
	current.searchResult, current.err = (application.SearchService{Searcher: current}).Search(context.Background(), query, newSearchSchema())
}
