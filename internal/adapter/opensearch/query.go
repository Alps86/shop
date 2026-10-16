package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"strings"
)

func (adapter Adapter) searchBody(query application.SearchQuery) (JSON, error) {
	plan, err := NewFilterPlan(query)
	if err != nil {
		return nil, err
	}

	variant := plan.variant()
	selection := plan.selection(variant, false)
	inner := plan.selection(variant, true)
	base := adapter.baseQuery(query)
	base["should"] = []JSON{inner}
	base["minimum_should_match"] = 0
	body := JSON{"_source": false, "stored_fields": []string{"_none_"}, "docvalue_fields": []string{"product_id"}, "from": (query.Page - 1) * 12, "size": 12, "track_total_hits": true, "query": (Ranking{Interests: query.Interests}).query(JSON{"bool": base}), "post_filter": selection, "sort": plan.productSort(variant), "aggs": adapter.facets(plan)}
	adapter.suggestBody(body, query)

	return body, nil
}

func (adapter Adapter) baseQuery(query application.SearchQuery) JSON {
	filters := []JSON{{"term": JSON{"deleted": false}}, {"term": JSON{"published": true}}}
	if query.CategoryID.Present {
		filters = append(filters, JSON{"term": JSON{"category_ids": query.CategoryID.Value}})
	}

	text := JSON{"match_all": JSON{}}
	if strings.TrimSpace(query.Text) != "" {
		text = JSON{"multi_match": JSON{"query": query.Text, "fields": []string{"name^3", "brand.text^2", "description"}}}
	}

	return JSON{"must": []JSON{text}, "filter": filters}
}

func (plan FilterPlan) productSort(variant JSON) []JSON {
	if plan.Query.Sort == "nameAsc" {
		return []JSON{{"name.sort": "asc"}, {"product_id": "asc"}}
	}

	if plan.priceSorted() {
		return plan.priceSort(variant)
	}

	return []JSON{{"_score": "desc"}, {"product_id": "asc"}}
}

func (plan FilterPlan) priceSorted() bool {
	return plan.Query.Sort == "priceAsc" || plan.Query.Sort == "priceDesc"
}

func (plan FilterPlan) priceSort(variant JSON) []JSON {
	order := "asc"
	mode := "min"
	if plan.Query.Sort == "priceDesc" {
		order = "desc"
		mode = "max"
	}

	sorting := JSON{"order": order, "mode": mode, "nested": JSON{"path": "variants", "filter": variant}}

	return []JSON{{"variants.price": sorting}, {"product_id": "asc"}}
}
