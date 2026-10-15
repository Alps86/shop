package http

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"
)

func (route PublicRoute) Validate() error {
	parsed, err := url.ParseRequestURI(route.Path)
	if err != nil {
		return fmt.Errorf("öffentliche Route: %w", err)
	}

	if parsed.RawQuery != "" {
		return fmt.Errorf("öffentliche Route enthält eine Query")
	}

	if parsed.Host != "" {
		return fmt.Errorf("öffentliche Route muss lokal sein")
	}

	if !route.publicPath() {
		return fmt.Errorf("öffentliche Route ist nicht indexierbar")
	}

	return nil
}

func (route PublicRoute) publicPath() bool {
	if route.Path == "/de-DE/" {
		return true
	}

	if route.Path == "/de-DE/service" {
		return true
	}

	if route.Path == "/de-DE/marken" {
		return true
	}

	if strings.HasPrefix(route.Path, "/de-DE/produkt/") {
		return true
	}

	return strings.HasPrefix(route.Path, "/de-DE/kategorie/")
}

func (config Config) validateRoutes() error {
	for _, route := range config.Routes {
		if err := route.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) sitemap(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Header().Set("Cache-Control", "public,max-age=3600")
	_, _ = fmt.Fprint(writer, `<?xml version="1.0" encoding="utf-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, route := range handler.config.Routes {
		_, _ = fmt.Fprintf(writer, "<url><loc>%s%s</loc></url>", html.EscapeString(handler.config.PublicOrigin), html.EscapeString(route.Path))
	}

	_, _ = fmt.Fprint(writer, "</urlset>")
}

func (handler *Handler) robots(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.Header().Set("Cache-Control", "public,max-age=3600")
	_, _ = fmt.Fprint(writer, "User-agent: *\nDisallow: /backend\nDisallow: /de-DE/warenkorb\nDisallow: /de-DE/kasse\nDisallow: /de-DE/bestellung/\nDisallow: /de-DE/profil\nDisallow: /de-DE/gast\nSitemap: "+handler.config.PublicOrigin+"/sitemap.xml\n")
}
