package http_test

import (
	"bytes"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

type world struct {
	fragmentFault      atomic.Bool
	engineClients      map[string]*http.Client
	sharedFrame        string
	lastCSRF           string
	cache              shophttp.CacheInvalidator
	cacheServer        *httptest.Server
	cacheGlobalStatus  int
	cacheStatus        int
	cacheRequests      []cacheRequest
	mutex              sync.Mutex
	options            map[string]string
	err                error
	engine             *compositionEngine
	response           *httptest.ResponseRecorder
	fixture            fixture
	cookies            []*http.Cookie
	fault              string
	calls              []string
	lastRecommendation application.RecommendationQuery
	lastQuery          application.SearchQuery
	lastChange         application.ChangeCommand
	lastConfirm        application.ConfirmCommand
	lastPlace          application.PlaceCommand
	lastInventory      application.InventoryCommand
	lastDelivery       application.DeliveryCommand
	handler            http.Handler
}

type fixture struct {
	Guests          map[application.Session]application.GuestView
	Carts           map[value.CartID]application.CartView
	Routes          []shophttp.PublicRoute
	Diagnostics     map[string]string
	FrameContent    string
	Preview         sale.CartEvaluation
	Now             time.Time
	Interests       []sale.CategoryInterest
	Detail          application.ProductDetail
	Category        application.CategoryView
	Search          application.SearchResult
	Guest           application.GuestView
	Cart            application.CartView
	Order           application.OrderView
	Purchase        application.PurchaseResult
	Recommendations application.RecommendationResult
	AdminCatalog    application.AdminCatalogSnapshot
	AdminOrders     application.AdminOrderPage
}

type catalogDouble struct {
	current *world
}

type searchDouble struct {
	current *world
}

type guestDouble struct {
	current *world
}

type recommendationDouble struct {
	current *world
}

type inventoryDouble struct {
	current *world
}

type adminOrderDouble struct {
	current *world
}

type processDouble struct {
	current *world
}

type rows []map[string]string

type compositionEngine struct {
	source    string
	directory string
	url       string
	origin    *httptest.Server
	nginx     *exec.Cmd
	varnish   *exec.Cmd
	output    engineLog
}

type engineLog struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

type compositionOrigin struct {
	current *world
	handler http.Handler
}

type frameSeed struct {
	Kind        string
	ID          string
	Revision    uint64
	Path        string
	Title       string
	Description string
	Markdown    string
}

type diagnosticDouble struct {
	current *world
}

type cacheOrigin struct {
	current *world
}

type cacheRequest struct {
	PublicSearch string
	Method       string
	SKU          string
	Key          string
}
