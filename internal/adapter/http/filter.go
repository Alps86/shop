package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
	"strings"
)

func (input queryInput) filterQuery(query *application.SearchQuery, key string, selected []string, schema application.SearchSchema) error {
	code := strings.TrimSuffix(strings.TrimSuffix(key, ".min"), ".max")
	if _, found := schema.Attributes[value.AttributeCode(code)]; !found {
		return newHTTPError(422, "Unbekannter Filter")
	}

	if key != code {
		return input.rangeFilter(query, key, selected, code)
	}

	for _, option := range selected {
		query.Filters.Options[value.AttributeCode(key)] = append(query.Filters.Options[value.AttributeCode(key)], value.OptionCode(option))
	}

	return nil
}

func (input queryInput) rangeFilter(query *application.SearchQuery, key string, selected []string, code string) error {
	if len(selected) != 1 {
		return newHTTPError(422, "Mehrdeutiges Intervall")
	}

	decimal, err := value.ParseDecimal(selected[0])
	if err != nil {
		return newHTTPError(422, "Ungültige Zahl")
	}

	input.storeRange(query, key, code, decimal)

	return nil
}

func (input queryInput) priceFilter(query *application.SearchQuery) error {
	minimum, err := input.priceBound("priceMin")
	if err != nil {
		return err
	}

	maximum, err := input.priceBound("priceMax")
	if err != nil {
		return err
	}

	if !minimum.Present {
		if !maximum.Present {
			return nil
		}
	}

	query.Filters.Price = value.Optional[application.MoneyRange]{Present: true, Value: application.MoneyRange{Minimum: minimum, Maximum: maximum}}

	return nil
}

func (input queryInput) priceBound(key string) (value.Optional[value.Money], error) {
	values := input.values
	if !values.Has(key) {
		return value.Optional[value.Money]{}, nil
	}

	amount, err := strconv.ParseInt(values.Get(key), 10, 64)
	if err != nil {
		return value.Optional[value.Money]{}, newHTTPError(422, "Ungültiger Preis in Cent")
	}

	return value.Optional[value.Money]{Present: true, Value: value.Money{MinorUnits: amount, Currency: "EUR"}}, nil
}

func (input queryInput) canonical() {
	values := input.values
	defaults := map[string]string{"page": "1", "sort": "relevance", "view": "grid"}
	for key, selected := range values {
		if querySelection(selected).matches(defaults[key]) {
			values.Del(key)
		}
	}
}

func (input queryInput) storeRange(query *application.SearchQuery, key, code string, decimal value.Decimal) {
	interval := query.Filters.Ranges[value.AttributeCode(code)]
	if strings.HasSuffix(key, ".min") {
		interval.Minimum = value.Optional[value.Decimal]{Value: decimal, Present: true}
	}

	if strings.HasSuffix(key, ".max") {
		interval.Maximum = value.Optional[value.Decimal]{Value: decimal, Present: true}
	}

	query.Filters.Ranges[value.AttributeCode(code)] = interval
}

func (input queryInput) standardKey(key string) bool {
	allowed := map[string]bool{"q": true, "page": true, "sort": true, "view": true, "brand": true, "priceMin": true, "priceMax": true}
	return allowed[key]
}

func (input queryInput) attributeCount() int {
	attributes := map[string]bool{}
	for key := range input.values {
		if input.standardKey(key) {
			continue
		}

		code := strings.TrimSuffix(strings.TrimSuffix(key, ".min"), ".max")
		attributes[code] = true
	}

	return len(attributes)
}
