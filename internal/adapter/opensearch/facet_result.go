package opensearch

import (
	"encoding/json"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

func (response SearchResponse) facets(query application.SearchQuery) ([]application.Facet, error) {
	facets := []application.Facet{}
	for _, name := range slices.Sorted(maps.Keys(response.Aggregations)) {
		code := strings.TrimPrefix(name, "facet_")
		facet, err := response.facet(code, response.Aggregations[name], query)
		if err != nil {
			return nil, err
		}

		facets = append(facets, facet)
	}

	return facets, nil
}

func (response SearchResponse) facet(code string, raw json.RawMessage, query application.SearchQuery) (application.Facet, error) {
	var decoded FacetResponse
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return application.Facet{}, value.NewError("SearchUnavailable", "Facet")
	}

	return response.facetBuckets(code, decoded.buckets(code), query)
}

func (response SearchResponse) facetBuckets(code string, buckets []FacetBucket, query application.SearchQuery) (application.Facet, error) {
	facet := application.Facet{Code: code}
	for _, bucket := range buckets {
		text, err := bucket.value(code)
		if err != nil {
			return application.Facet{}, err
		}

		count := bucket.productCount(code)
		facet.Values = append(facet.Values, application.FacetValue{Value: text, ProductCount: count})
	}

	return (FilterPlan{Query: query}).retainSelected(facet), nil
}

func (bucket FacetBucket) value(code string) (string, error) {
	if strings.HasPrefix(string(bucket.Key), "\"") {
		var text string
		err := json.Unmarshal(bucket.Key, &text)

		return text, err
	}

	var number int64
	if err := json.Unmarshal(bucket.Key, &number); err != nil {
		return "", value.NewError("SearchUnavailable", "FacetKey")
	}

	if code == "price" {
		return strconv.FormatInt(number, 10), nil
	}

	canonical, err := value.ParseDecimal(big.NewRat(number, 1000000).FloatString(6))

	return string(canonical), err
}

func (plan FilterPlan) retainSelected(facet application.Facet) application.Facet {
	selected := plan.selectedValues(facet.Code)
	seen := map[string]bool{}
	for index, item := range facet.Values {
		seen[item.Value] = true
		facet.Values[index].Selected = slices.Contains(selected, item.Value)
	}

	for _, text := range selected {
		if !seen[text] {
			facet.Values = append(facet.Values, application.FacetValue{Value: text, Selected: true})
		}
	}

	slices.SortFunc(facet.Values, func(left, right application.FacetValue) int {
		return strings.Compare(left.Value, right.Value)
	})

	return facet
}

func (plan FilterPlan) selectedValues(code string) []string {
	if code == "brand" {
		return plan.Query.Filters.Brands
	}

	selected := []string{}
	for _, option := range plan.Query.Filters.Options[value.AttributeCode(code)] {
		selected = append(selected, string(option))
	}

	return selected
}

func (bucket FacetBucket) productCount(code string) int64 {
	if code == "brand" {
		return bucket.Count
	}

	return bucket.Products.Count
}

func (decoded FacetResponse) buckets(code string) []FacetBucket {
	if code == "brand" {
		return decoded.Values.Buckets
	}

	return decoded.Variants.Matching.Values.Buckets
}
