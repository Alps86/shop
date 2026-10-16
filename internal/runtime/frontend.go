package runtime

import (
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/adapter/hugo"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"os"
	"strings"
)

func environmentFrontend() FrontendConfig {
	return FrontendConfig{PublicOrigin: os.Getenv("SHOP_PUBLIC_ORIGIN"), ContentDirectory: os.Getenv("SHOP_CONTENT_DIRECTORY"), CacheEndpoint: os.Getenv("SHOP_CACHE_ENDPOINT"), CacheKey: os.Getenv("SHOP_CACHE_KEY"), CSRFKey: os.Getenv("SHOP_CSRF_KEY"), BackendListen: os.Getenv("SHOP_BACKEND_LISTEN")}
}

func (app *App) bindFrontend(snapshot application.AdminCatalogSnapshot) error {
	cache, err := shophttp.NewCacheInvalidator(app.Config.Frontend.CacheEndpoint, app.Config.Frontend.CacheKey)
	if err != nil {
		return value.NewError("InvalidInput", "FrontendConfiguration")
	}

	app.cache = cache

	return app.installFrontend(app.frontendConfig(snapshot))
}

func (app *App) installFrontend(config shophttp.Config) error {
	handlers := []http.Handler{}
	for _, factory := range []func(shophttp.Config) (http.Handler, error){shophttp.New, shophttp.NewBackend} {
		handler, err := factory(config)
		if err != nil {
			return value.NewError("InvalidInput", "FrontendConfiguration")
		}

		handlers = append(handlers, handler)
	}

	app.installHandlers(handlers)

	return nil
}

func (app *App) installHandlers(handlers []http.Handler) {
	mux := app.healthHandler()
	mux.Handle("/", handlers[0])
	app.Handler = mux
	app.Backend = handlers[1]
}

func (app *App) frontendConfig(snapshot application.AdminCatalogSnapshot) shophttp.Config {
	return shophttp.Config{
		Preview: app.Service, Guests: app.Service, Interests: app.Service,
		Catalog: app.Store.Catalog(), Variants: app.Store.Catalog(), AdminCatalog: app.Store.Catalog(),
		Search: app.Search, Recommendations: application.RecommendationService{Source: app.Store.Recommendations()},
		Inventory: app.Store.Inventory(), AdminOrders: app.Store.AdminOrders(), Processes: app.Processes,
		Clock: opensearch.Clock{}, Diagnostics: app, Cache: app.cache,
		CSRFKey: []byte(app.Config.Frontend.CSRFKey), SecureCookie: strings.HasPrefix(app.Config.Frontend.PublicOrigin, "https://"),
		ContentDirectory: app.Config.Frontend.ContentDirectory, PublicOrigin: app.Config.Frontend.PublicOrigin,
		Routes: hugo.New(snapshot, app.Config.Context.Locale).Routes(),
	}
}
