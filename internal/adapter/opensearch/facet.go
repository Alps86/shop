package opensearch

import (
	"github.com/Alps86/shop/internal/domain/value"
	"maps"
	"slices"
)

func (adapter Adapter) facets(plan FilterPlan) JSON {
	aggregations := JSON{}
	plan.Omit = "brand"
	aggregations["facet_brand"] = JSON{"filter": plan.selection(plan.variant(), false), "aggs": JSON{"values": JSON{"terms": JSON{"field": "brand", "size": 100}}}}
	plan.Omit = "price"
	aggregations["facet_price"] = plan.nestedFacet("variants.price")
	for _, code := range adapter.facetCodes(plan) {
		plan.Omit = string(code)
		aggregations["facet_"+string(code)] = plan.nestedFacet(adapter.facetField(code))
	}

	return aggregations
}

func (adapter Adapter) facetCodes(plan FilterPlan) []value.AttributeCode {
	if plan.Query.CategoryID.Present {
		return adapter.config.Categories[plan.Query.CategoryID.Value].FilterAttributes
	}

	codes := slices.Sorted(maps.Keys(adapter.config.Schema.Attributes))

	return codes[:min(4, len(codes))]
}

func (adapter Adapter) facetField(code value.AttributeCode) string {
	kind := adapter.config.Schema.Attributes[code].Type
	if kind == "number" {
		return "variants.numbers." + string(code)
	}

	if kind == "measurement" {
		return "variants.numbers." + string(code)
	}

	return "variants.options." + string(code)
}

func (plan FilterPlan) nestedFacet(field string) JSON {
	values := JSON{"terms": JSON{"field": field, "size": 100}, "aggs": JSON{"products": JSON{"reverse_nested": JSON{}}}}
	matching := JSON{"filter": plan.variant(), "aggs": JSON{"values": values}}
	variants := JSON{"nested": JSON{"path": "variants"}, "aggs": JSON{"matching": matching}}

	return JSON{"filter": plan.selection(plan.variant(), false), "aggs": JSON{"variants": variants}}
}
