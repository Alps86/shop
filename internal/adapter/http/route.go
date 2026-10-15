package http

import (
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"strings"
)

func (handler *Handler) dispatch(writer http.ResponseWriter, request *http.Request) {
	routes := map[string]http.HandlerFunc{"/de-DE/": handler.home, "/de-DE/vorschlag": handler.suggestion, "/de-DE/suche": handler.listing, "/de-DE/marken": handler.editorial, "/de-DE/service": handler.editorial, "/de-DE/warenkorb": handler.cart, "/de-DE/kasse": handler.checkout, "/de-DE/profil": handler.profile, "/de-DE/gast": handler.openGuest, "/de-DE/warenkorb/aktion": handler.cartAction, "/de-DE/warenkorb/badge": handler.badge}
	if route, found := routes[request.URL.Path]; found {
		route(writer, request)
		return
	}

	handler.catalogRoute(writer, request)
}

func (handler *Handler) home(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page := newPage("home", "Technik für deinen Alltag")
	if request.Header.Get("X-Shop-Fragment") == "1" {
		page.Kind = "offers"
	}

	page.Canonical = "/de-DE/"
	page.CacheControl = "public,max-age=5"
	handler.supplementSearch(request, &page)

	handler.render(writer, request, page)
}

func (handler *Handler) editorial(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page := newPage("editorial", "Service und Kontakt")
	page.Canonical = request.URL.Path
	page.CacheControl = "public,max-age=3600"
	if request.URL.Path == "/de-DE/marken" {
		page.Title = "Markenwelt"
	}

	handler.supplementSearch(request, &page)

	handler.render(writer, request, page)
}

func (handler *Handler) readOnly(writer http.ResponseWriter, request *http.Request) bool {
	if request.Method != "GET" {
		handler.failure(writer, request, newHTTPError(405, "Nur lesende Navigation erlaubt"))
		return false
	}

	return true
}

func (path routePath) id() string {
	parts := strings.Split(string(path), "--")
	return parts[len(parts)-1]
}

func (handler *Handler) redirectSlug(writer http.ResponseWriter, request *http.Request, target string) bool {
	if request.URL.Path == target {
		return false
	}

	query := request.URL.Query().Encode()
	if query != "" {
		target += "?" + query
	}

	// #nosec G710 -- Fester eigener Routenpräfix aus Katalogdaten; Query wird URLkodiert.
	http.Redirect(writer, request, target, http.StatusMovedPermanently)

	return true
}

func newRequestContext() value.RequestContext {
	return value.RequestContext{Locale: "de-DE", Market: "DE", Currency: "EUR"}
}

func (handler *Handler) catalogRoute(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/de-DE/produkt/") {
		handler.product(writer, request)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/de-DE/kategorie/") {
		handler.category(writer, request)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/de-DE/bestellung/") {
		handler.order(writer, request)
		return
	}

	handler.failure(writer, request, newHTTPError(404, "Nicht gefunden"))
}

func (handler *Handler) supplementSearch(request *http.Request, page *Page) {
	if handler.config.Search == nil {
		return
	}

	result, err := handler.config.Search.Search(request.Context(), newSearchQuery(page.Context))
	if err != nil {
		page.Message = "Angebote vorübergehend nicht verfügbar. Bitte erneut laden."
		return
	}

	page.Search = result
}
