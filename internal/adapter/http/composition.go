package http

import (
	"context"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (handler *Handler) fragment(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), 150*time.Millisecond)
	defer cancel()
	clone := request.Clone(ctx)
	clone.URL.Path = strings.TrimPrefix(request.URL.Path, "/fragments")
	handler.fragmentQuery(clone)
	clone.Header.Del("Cookie")
	clone.Header.Del("Authorization")
	clone.Header.Set("HX-Request", "true")
	clone.Header.Set("X-Shop-Fragment", "1")
	if !routePath(clone.URL.Path).fragmentAllowed() {
		handler.failure(writer, request, newHTTPError(404, "Fragment nicht gefunden"))
		return
	}

	handler.dispatch(writer, clone)
}

func (path routePath) fragmentAllowed() bool {
	if path == "/de-DE/" {
		return true
	}

	if path == "/de-DE/suche" {
		return true
	}

	if strings.HasPrefix(string(path), "/de-DE/produkt/") {
		return true
	}

	return strings.HasPrefix(string(path), "/de-DE/kategorie/")
}

func (page *Page) composition(request *http.Request) {
	if request.Header.Get("X-Shop-Composition") != "1" {
		return
	}

	if (requestView{request}).session() != "" {
		return
	}

	if !routePath(request.URL.Path).fragmentAllowed() {
		return
	}

	page.SSIHeader = template.HTML(`<!--# include virtual="/_static/header.html" -->`)
	page.SSIFooter = template.HTML(`<!--# include virtual="/_static/footer.html" -->`)
	page.Fragment.URL = "/fragments" + request.URL.RequestURI()
	if page.Detail.Product.ID != "" {
		page.rootVersion()
	}
}

func (handler *Handler) fragmentFailure(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "private,no-store")
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = writer.Write([]byte(`<section role="alert">Daten vorübergehend nicht verfügbar. <a href="/de-DE/">Erneut laden</a></section>`))
}

func (page Page) surrogateKeys() string {
	keys := []string{}
	for _, card := range page.Search.Cards {
		keys = append(keys, string(card.SKU))
	}

	for _, card := range page.Detail.Cards {
		keys = append(keys, string(card.SKU))
	}

	return strings.Join(keys, " ")
}

func (page *Page) rootVersion() {
	separator := "?"
	if strings.Contains(page.Fragment.URL, "?") {
		separator = "&"
	}

	page.Fragment.URL += separator + "rootVersion=" + strconv.FormatUint(page.Detail.Product.Revision, 10)
}

func (handler *Handler) fragmentQuery(request *http.Request) {
	if strings.Contains(request.URL.Path, "/produkt/") {
		return
	}

	query := request.URL.Query()
	query.Del("rootVersion")
	request.URL.RawQuery = query.Encode()
}
