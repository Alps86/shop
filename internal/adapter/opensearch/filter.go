package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
)

func (plan FilterPlan) selection(variant JSON, inner bool) JSON {
	if inner {
		return plan.rankedInner(variant)
	}

	nested := JSON{"path": "variants", "query": variant, "score_mode": "none"}
	filters := []JSON{{"nested": nested}}
	if plan.includesBrand() {
		filters = append(filters, JSON{"terms": JSON{"brand": plan.Query.Filters.Brands}})
	}

	return JSON{"bool": JSON{"filter": filters}}
}

func (plan FilterPlan) includesBrand() bool {
	return plan.Omit != "brand" && len(plan.Query.Filters.Brands) > 0
}

func NewFilterPlan(query application.SearchQuery) (FilterPlan, error) {
	plan := FilterPlan{Query: query, Predicates: map[string]JSON{}}
	for code, options := range query.Filters.Options {
		plan.Predicates[string(code)] = JSON{"terms": JSON{"variants.options." + string(code): options}}
	}

	for code, interval := range query.Filters.Ranges {
		bounds, err := NewNumberRange(interval)
		if err != nil {
			return FilterPlan{}, value.NewError("InvalidFilter", string(code))
		}

		plan.Predicates[string(code)] = JSON{"range": JSON{"variants.numbers." + string(code): bounds}}
	}

	if plan.includesPrice() {
		plan.Predicates["price"] = plan.priceFilter()
	}

	return plan, nil
}

func (plan FilterPlan) variant() JSON {
	filters := []JSON{}
	for code, predicate := range plan.Predicates {
		if code == plan.Omit {
			continue
		}

		filters = append(filters, predicate)
	}

	return JSON{"bool": JSON{"filter": filters}}
}

func NewNumberRange(interval application.NumericRange) (JSON, error) {
	bounds := JSON{}
	if err := bounds.decimalBound("gte", interval.Minimum); err != nil {
		return nil, err
	}

	if err := bounds.decimalBound("lte", interval.Maximum); err != nil {
		return nil, err
	}

	return bounds, nil
}

func (bounds JSON) decimalBound(key string, optional value.Optional[value.Decimal]) error {
	if !optional.Present {
		return nil
	}

	number, err := optional.Value.Scaled()
	bounds[key] = number

	return err
}

func (plan FilterPlan) includesPrice() bool {
	return plan.Omit != "price" && plan.Query.Filters.Price.Present
}

func (plan FilterPlan) priceFilter() JSON {
	interval := plan.Query.Filters.Price.Value
	bounds := JSON{}
	if interval.Minimum.Present {
		bounds["gte"] = interval.Minimum.Value.MinorUnits
	}

	if interval.Maximum.Present {
		bounds["lte"] = interval.Maximum.Value.MinorUnits
	}

	return JSON{"range": JSON{"variants.price": bounds}}
}

func (plan FilterPlan) variantSort() []JSON {
	if plan.Query.Sort == "priceAsc" {
		return []JSON{{"variants.price": "asc"}, {"variants.sku": "asc"}}
	}

	if plan.Query.Sort == "priceDesc" {
		return []JSON{{"variants.price": "desc"}, {"variants.sku": "asc"}}
	}

	return []JSON{{"variants.available": "desc"}, {"variants.price": "asc"}, {"variants.sku": "asc"}}
}

func (plan FilterPlan) rankedInner(variant JSON) JSON {
	functions := []JSON{{"field_value_factor": JSON{"field": "variants.margin", "factor": 0.00000005, "missing": 0}}}
	ranking := JSON{"function_score": JSON{"query": variant, "functions": functions, "boost_mode": "replace", "score_mode": "sum"}}
	nested := JSON{"path": "variants", "query": ranking, "score_mode": "max", "inner_hits": JSON{"name": "chosen", "size": 1, "sort": plan.variantSort()}}

	return JSON{"nested": nested}
}
