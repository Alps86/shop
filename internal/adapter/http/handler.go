package http

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

//go:embed template/*.html
var templateFiles embed.FS

func New(config Config) (http.Handler, error) {
	config, err := newConfig(config)
	if err != nil {
		return nil, err
	}

	renderer, err := newRenderer(config)
	if err != nil {
		return nil, err
	}

	handler := &Handler{config: config, renderer: renderer, routes: http.NewServeMux()}
	handler.register()

	return handler, nil
}

func NewBackend(config Config) (http.Handler, error) {
	public, err := New(config)
	if err != nil {
		return nil, err
	}

	handler := public.(*Handler)
	handler.backend = true

	return handler, nil
}

func (handler *Handler) register() {
	handler.routes.HandleFunc("/", handler.route)
	handler.routes.HandleFunc("/robots.txt", handler.robots)
	handler.routes.HandleFunc("/sitemap.xml", handler.sitemap)
	handler.routes.HandleFunc("/assets/", handler.asset)
	handler.routes.HandleFunc("/fragments/", handler.fragment)
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Vary", "HX-Request")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	handler.routes.ServeHTTP(writer, request)
}

func (handler *Handler) render(writer http.ResponseWriter, request *http.Request, page Page) {
	page.Routes = handler.config.Routes
	page.prepare(request)
	page.composition(request)
	page.limitCache(handler.config.Clock.Now())
	name := "page"
	if request.Header.Get("HX-Request") == "true" {
		name = "body"
	}

	var output bytes.Buffer
	if err := handler.renderer.templates.ExecuteTemplate(&output, name, page); err != nil {
		writer.Header().Set("Cache-Control", "private,no-store")
		http.Error(writer, "Ansicht vorübergehend nicht verfügbar", http.StatusServiceUnavailable)

		return
	}

	page.write(writer, output.Bytes())
}

func (handler *Handler) failure(writer http.ResponseWriter, request *http.Request, err error) {
	if request.Header.Get("X-Shop-Fragment") == "1" {
		handler.fragmentFailure(writer)
		return
	}

	code := (responseError{err}).status()
	if code == 202 {
		writer.Header().Set("Retry-After", "2")
		writer.Header().Set("Location", "/de-DE/warenkorb")
	}

	page := newPage("error", "Anfrage nicht möglich")
	page.Status = code
	page.Message = (responseError{err}).message()
	page.retry(request)
	page.Robots = "noindex,nofollow"
	page.CacheControl = "private,no-store"
	handler.render(writer, request, page)
}

func (handler *Handler) route(writer http.ResponseWriter, request *http.Request) {
	if handler.backend {
		handler.admin(writer, request)
		return
	}

	if len(request.URL.RequestURI()) > 2048 {
		handler.failure(writer, request, newHTTPError(414, "URL zu lang"))
		return
	}

	if request.URL.Path == "/" {
		http.Redirect(writer, request, "/de-DE/", http.StatusFound)
		return
	}

	handler.publicRoute(writer, request)
}

func (handler *Handler) publicRoute(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, "/backend") {
		handler.failure(writer, request, newHTTPError(404, "Nicht gefunden"))
		return
	}

	if err := handler.validateLocale(request); err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.dispatch(writer, request)
}

func (page *Page) prepare(request *http.Request) {
	page.Form = request.Form
	if page.Form == nil {
		page.Form = request.URL.Query()
	}

	page.Path = request.URL.Path
	if (requestView{request}).session() != "" {
		page.CacheControl = "private,no-store"
	}
}

func (page Page) write(writer http.ResponseWriter, body []byte) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Surrogate-Key", page.surrogateKeys())
	writer.Header().Set("X-Robots-Tag", page.Robots)
	writer.Header().Set("Cache-Control", page.CacheControl)
	writer.WriteHeader(page.Status)
	_, _ = writer.Write(body)
}

func newConfig(config Config) (Config, error) {
	if len(config.CSRFKey) < 32 {
		return config, fmt.Errorf("csrfkey benötigt mindestens 32 Bytes")
	}

	if config.Clock == nil {
		config.Clock = systemClock{}
	}

	if config.TemplateFS == nil {
		config.TemplateFS = templateFiles
	}

	if config.PublicOrigin == "" {
		config.PublicOrigin = "http://localhost:8080"
	}

	return config, config.validateRoutes()
}

func (renderer Renderer) functions() template.FuncMap {
	return template.FuncMap{"cardViews": renderer.cardViews, "suggestionURL": suggestionURL, "money": renderer.money, "text": renderer.localized, "attribute": renderer.attribute, "productURL": renderer.productURL, "categoryURL": renderer.categoryURL, "basePrice": renderer.basePrice, "field": renderer.fieldValue, "quantity": renderer.itemQuantity, "can": renderer.can}
}

func newRenderer(config Config) (Renderer, error) {
	renderer := Renderer{}
	templates, err := template.New("shop").Funcs(renderer.functions()).ParseFS(config.TemplateFS, "template/*.html")
	if err != nil {
		return renderer, fmt.Errorf("shoptemplates: %w", err)
	}

	renderer.templates = templates

	return renderer, nil
}
