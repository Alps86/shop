package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"net/http"
	"strconv"
)

func (renderer Renderer) fieldValue(form map[string][]string, key, fallback string) string {
	values, exists := form[key]
	if !exists {
		return fallback
	}

	return values[0]
}

func (renderer Renderer) itemQuantity(page Page, item sale.CartItem) string {
	if renderer.fieldValue(page.Form, "sku", "") == string(item.SKU) {
		return renderer.fieldValue(page.Form, "quantity", strconv.FormatInt(item.Quantity, 10))
	}

	return strconv.FormatInt(item.Quantity, 10)
}

func (handler *Handler) suggestion(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	query, err := handler.query(request, application.SearchSchema{})
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	result, err := handler.config.Search.Search(request.Context(), query)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	page := handler.privatePage(writer, request, "suggestions", "Suchvorschläge")
	page.Search = result
	handler.render(writer, request, page)
}
