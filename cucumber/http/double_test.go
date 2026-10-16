package http_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
	"strings"
)

func (current *world) portError(method string) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	current.calls = append(current.calls, method)
	fault, code, found := strings.Cut(current.fault, ":")
	if !current.matchesFault(fault, method) {
		return nil
	}

	if !found {
		code = "InternalError"
	}

	return value.NewError(code, method)
}

func (current *world) matchesFault(fault, method string) bool {
	name, ordinal, found := strings.Cut(fault, "#")
	if name != method {
		return false
	}

	if !found {
		return true
	}

	expected, _ := strconv.Atoi(ordinal)
	count := 0
	for _, call := range current.calls {
		if call == method {
			count++
		}
	}

	return count == expected
}

func (double catalogDouble) Product(_ context.Context, id value.ProductID, sku value.Optional[value.SKU], _ value.RequestContext) (application.ProductDetail, error) {
	result := double.current.fixture.Detail
	if id == "p-removed" {
		return result, value.NewError("Removed", "Product")
	}

	if id != result.Product.ID {
		return result, value.NewError("NotFound", "Product")
	}

	if sku.Present {
		if _, err := result.Product.Variant(sku.Value); err != nil {
			return result, value.NewError("NotFound", "SKU")
		}

		result.ChosenSKU = sku.Value
	}

	return result, double.current.portError("Product")
}

func (double catalogDouble) Category(_ context.Context, id value.CategoryID, _ value.RequestContext) (application.CategoryView, error) {
	result := double.current.fixture.Category
	if id != result.Category.ID {
		return result, value.NewError("NotFound", "Category")
	}

	return result, double.current.portError("Category")
}

func (double catalogDouble) Snapshot(context.Context) (application.AdminCatalogSnapshot, error) {
	return double.current.fixture.AdminCatalog, double.current.portError("Snapshot")
}

func (double searchDouble) Search(_ context.Context, query application.SearchQuery) (application.SearchResult, error) {
	double.current.lastQuery = query
	return double.current.fixture.Search, double.current.portError("Search")
}

func (double recommendationDouble) Recommend(_ context.Context, query application.RecommendationQuery) (application.RecommendationResult, error) {
	double.current.lastRecommendation = query
	return double.current.fixture.Recommendations, double.current.portError("Recommend")
}

func (double catalogDouble) ProductForSKU(_ context.Context, sku value.SKU, _ value.RequestContext) (application.ProductDetail, error) {
	result := double.current.fixture.Detail
	result.ChosenSKU = sku

	return result, double.current.portError("ProductForSKU")
}
