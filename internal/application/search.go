package application

import (
	"context"

	"github.com/Alps86/shop/internal/domain/value"
)

func (query SearchQuery) Validate(schema SearchSchema) error {
	if err := query.Context.Validate(); err != nil {
		return err
	}

	if query.invalidPageOrSort() {
		return value.NewError("InvalidFilter", "PageOrSort")
	}

	if len(query.Filters.Operators) != 0 {
		return value.NewError("InvalidFilter", "FilterOperator")
	}

	return query.validateAttributes(schema)
}

func (query SearchQuery) validateAttributes(schema SearchSchema) error {
	for attribute := range query.Filters.Options {
		if !schema.curated(attribute) {
			return value.NewError("InvalidFilter", "Attribute")
		}
	}

	for attribute := range query.Filters.Ranges {
		if !schema.curated(attribute) {
			return value.NewError("InvalidFilter", "Attribute")
		}
	}

	return schema.validateValues(query.Filters, query.Context.Currency)
}

func (query SearchQuery) invalidPageOrSort() bool {
	sorts := map[Sort]bool{"relevance": true, "priceAsc": true, "priceDesc": true, "nameAsc": true}
	return query.Page < 1 || query.Page > 100 || !sorts[query.Sort]
}

func (schema SearchSchema) curated(code value.AttributeCode) bool {
	_, exists := schema.Attributes[code]
	return exists
}

func (service SearchService) Search(ctx context.Context, query SearchQuery, schema SearchSchema) (SearchResult, error) {
	if err := query.Validate(schema); err != nil {
		return SearchResult{}, err
	}

	result, err := service.Searcher.Search(ctx, query)
	if err != nil {
		return SearchResult{}, value.NewError("SearchUnavailable", "Search")
	}

	result.PageSize = 12

	return result, nil
}
