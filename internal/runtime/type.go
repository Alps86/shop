package runtime

import (
	"github.com/Alps86/shop/internal/adapter/cdc"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/adapter/postgres"
	shoptemporal "github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/client"
	"net/http"
	"sync"
)

type Config struct {
	DSN               string
	Schema            string
	TemporalAddress   string
	TemporalNamespace string
	TaskQueue         string
	SearchEndpoint    string
	SearchIndex       string
	Publication       string
	Slot              string
	Listen            string
	SeedFile          string
	ContentSource     string
	Frontend          FrontendConfig
	Context           value.RequestContext
}

type App struct {
	Config           Config
	Store            *postgres.Store
	Client           client.Client
	Processes        shoptemporal.Adapter
	Service          application.Service
	Search           opensearch.Adapter
	Consumer         *cdc.Consumer
	Handler          http.Handler
	Backend          http.Handler
	Seed             Seed
	cache            shophttp.CacheInvalidator
	mutex            sync.Mutex
	dispatchError    error
	dispatchStarted  bool
	projectionStatus string
}

type Seed struct {
	postgres.ImportManifest
	PurchaseCosts map[value.SKU]value.Money
}

type FrontendConfig struct {
	PublicOrigin     string
	ContentDirectory string
	CacheEndpoint    string
	CacheKey         string
	CSRFKey          string
	BackendListen    string
}

type cacheProjection struct {
	opensearch.Adapter
	cache shophttp.CacheInvalidator
	skus  []value.SKU
}
