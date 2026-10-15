package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"strconv"
)

func (handler *Handler) listing(writer http.ResponseWriter, request *http.Request) {
	handler.searchPage(writer, request, application.CategoryView{})
}

func (handler *Handler) category(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	category, err := handler.config.Catalog.Category(request.Context(), value.CategoryID(routePath(request.URL.Path).id()), newRequestContext())
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.categoryRoute(writer, request, category)
}

func (handler *Handler) searchPage(writer http.ResponseWriter, request *http.Request, category application.CategoryView) {
	query, err := handler.categoryQuery(request, category)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	result, err := handler.config.Search.Search(request.Context(), query)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	page := newPage("listing", "Technik entdecken")
	page.Category = category
	page.Query = query
	page.Search = result
	page.Canonical = request.URL.Path
	page.searchMetadata(request)
	handler.render(writer, request, page)
}

func (page *Page) searchMetadata(request *http.Request) {
	if page.Category.Category.ID != "" {
		page.Title = page.Category.Category.Name[page.Context.Locale]
	}

	page.searchIndexing(request)
	if (requestView{request}).session() != "" {
		page.CacheControl = "private,no-store"
	}

	page.pagination(request)
}

func (handler *Handler) product(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	detail, err := handler.productDetail(request)
	if err != nil {
		writer.Header().Set("Retry-After", "2")
		handler.failure(writer, request, err)

		return
	}

	handler.productRoute(writer, request, detail)
}

func (handler *Handler) productRecommendations(request *http.Request, page *Page) {
	query := application.RecommendationQuery{ProductID: page.Detail.Product.ID, SKU: page.Detail.ChosenSKU, Context: page.Context, Limit: 4}
	page.Recommendations = handler.recommendations(request, query)
}

func (handler *Handler) categoryQuery(request *http.Request, category application.CategoryView) (application.SearchQuery, error) {
	schema := newCategorySchema(category)

	query, err := handler.query(request, schema)
	if category.Category.ID != "" {
		query.CategoryID = value.Optional[value.CategoryID]{Value: category.Category.ID, Present: true}
	}

	if err != nil {
		return query, err
	}

	query.Interests, err = handler.requestInterests(request)

	return query, err
}

func (handler *Handler) productDetail(request *http.Request) (application.ProductDetail, error) {
	sku := value.Optional[value.SKU]{Value: value.SKU(request.URL.Query().Get("sku")), Present: request.URL.Query().Has("sku")}
	if sku.Present {
		if err := value.IdentifierSyntax(sku.Value).Validate(); err != nil {
			return application.ProductDetail{}, newHTTPError(422, "Ungültige Variante")
		}
	}

	for key := range request.URL.Query() {
		if !handler.allowedProductKey(key, request) {
			return application.ProductDetail{}, newHTTPError(422, "Unbekannte Produktquery")
		}
	}

	return handler.config.Catalog.Product(request.Context(), value.ProductID(routePath(request.URL.Path).id()), sku, newRequestContext())
}

func (page *Page) pagination(request *http.Request) {
	count := (page.Search.TotalProducts + 11) / 12
	first := max(1, page.Query.Page-2)
	last := min(int(count), page.Query.Page+2)
	for number := first; number <= last; number++ {
		query := request.URL.Query()
		query.Set("page", strconv.Itoa(number))
		(queryInput{query}).canonical()
		target := request.URL.Path
		if encoded := query.Encode(); encoded != "" {
			target += "?" + encoded
		}

		page.Pagination = append(page.Pagination, PaginationLink{URL: target, Number: number, Current: number == page.Query.Page})
	}
}

func (page *Page) variantRoles() error {
	roles, err := page.Detail.Product.ResolveRoles(page.Detail.ChosenSKU)
	if err != nil {
		return err
	}

	page.Manufacturer = roles.Manufacturer
	page.ResponsiblePerson = roles.ResponsiblePerson

	return nil
}

func (page *Page) searchIndexing(request *http.Request) {
	page.Canonical = request.URL.Path
	page.Robots = "noindex,follow"
	if request.URL.RawQuery == "" {
		page.Robots = "index,follow"
	}

	if request.URL.RawQuery == "page="+request.URL.Query().Get("page") {
		page.Canonical = request.URL.String()
		page.Robots = "index,follow"
	}

	if request.URL.Path == "/de-DE/suche" {
		page.Robots = "noindex,follow"
		page.Canonical = request.URL.String()
	}
}

func (handler *Handler) showProduct(writer http.ResponseWriter, request *http.Request, detail application.ProductDetail) {
	page := newPage("product", handler.renderer.localized(detail.Product.Name))
	page.Detail = detail
	page.Canonical = handler.renderer.productURL(detail.Product)
	page.Description = handler.renderer.localized(detail.Product.Description)
	if err := page.variantRoles(); err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.productRecommendations(request, &page)
	handler.render(writer, request, page)
}

func (handler *Handler) allowedProductKey(key string, request *http.Request) bool {
	if key == "sku" {
		return true
	}

	if key != "rootVersion" {
		return false
	}

	return request.Header.Get("X-Shop-Fragment") == "1"
}

func (handler *Handler) categoryRoute(writer http.ResponseWriter, request *http.Request, category application.CategoryView) {
	if _, err := handler.query(request, newCategorySchema(category)); err != nil {
		handler.failure(writer, request, err)
		return
	}

	if handler.redirectSlug(writer, request, handler.renderer.categoryURL(category.Category)) {
		return
	}

	frame := frameReference{Kind: "category", ID: string(category.Category.ID), Revision: category.Category.Version}
	if handler.accelerate(writer, request, frame) {
		return
	}

	handler.searchPage(writer, request, category)
}

func newCategorySchema(category application.CategoryView) application.SearchSchema {
	schema := application.SearchSchema{Attributes: map[value.AttributeCode]catalog.AttributeDefinition{}}
	for _, definition := range category.FilterDefinitions {
		schema.Attributes[definition.Code] = definition
	}

	return schema
}

func (handler *Handler) productRoute(writer http.ResponseWriter, request *http.Request, detail application.ProductDetail) {
	if handler.redirectSlug(writer, request, handler.renderer.productURL(detail.Product)) {
		return
	}

	frame := frameReference{Kind: "product", ID: string(detail.Product.ID), Revision: detail.Product.Revision}
	if handler.accelerate(writer, request, frame) {
		return
	}

	handler.showProduct(writer, request, detail)
}

func (handler *Handler) recommendations(request *http.Request, query application.RecommendationQuery) application.RecommendationResult {
	if handler.config.Recommendations == nil {
		return application.RecommendationResult{}
	}

	interests, err := handler.requestInterests(request)
	if err != nil {
		return application.RecommendationResult{Degraded: true}
	}

	query.Interests = interests
	result, err := handler.config.Recommendations.Recommend(request.Context(), query)
	if err != nil {
		return application.RecommendationResult{Degraded: true}
	}

	return result
}
