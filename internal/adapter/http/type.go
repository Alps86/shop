package http

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"time"
)

type GuestService interface {
	Open(context.Context, application.Session) (application.GuestView, error)
	SelectProfile(context.Context, application.Session, application.ProfileCode) (application.GuestView, error)
	GetCart(context.Context, application.Session, value.CartID) (application.CartView, error)
	StartCart(context.Context, application.Session, value.CommandID) (application.CartView, error)
	Change(context.Context, application.Session, application.ChangeCommand) (application.CartView, error)
	Evaluate(context.Context, application.Session, value.CartID) (application.CartView, error)
	Confirm(context.Context, application.Session, application.ConfirmCommand) (application.CartView, error)
	Place(context.Context, application.Session, application.PlaceCommand) (application.PurchaseResult, error)
	GetOrder(context.Context, application.Session, value.OrderID) (application.OrderView, error)
}

type Config struct {
	Preview          AccessoryPreview
	Variants         VariantReader
	Routes           []PublicRoute
	PublicOrigin     string
	Interests        InterestReader
	Clock            application.Clock
	Diagnostics      Diagnostics
	Cache            CacheInvalidation
	TemplateFS       fs.FS
	Guests           GuestService
	Catalog          application.CatalogReader
	Search           application.SearchPort
	Recommendations  application.RecommendationPort
	AdminCatalog     application.AdminCatalogReader
	AdminOrders      application.AdminOrderReader
	Inventory        application.InventoryService
	Processes        application.ProcessPort
	CSRFKey          []byte
	SecureCookie     bool
	ContentDirectory string
}

type Handler struct {
	config   Config
	renderer Renderer
	routes   *http.ServeMux
	backend  bool
}

type Page struct {
	Routes            []PublicRoute
	ProfileChoice     application.ProfileCode
	Preview           value.Optional[sale.CartEvaluation]
	PreviewSKU        value.SKU
	Products          []application.ProductDetail
	Diagnostics       map[string]string
	RetryAction       string
	RetryFields       []InputField
	Manufacturer      catalog.EconomicOperator
	ResponsiblePerson value.Optional[catalog.EconomicOperator]
	SSIHeader         template.HTML
	SSIFooter         template.HTML
	Fragment          Fragment
	Pagination        []PaginationLink
	Alternatives      []application.ProductCard
	UnavailableSKU    value.Optional[value.SKU]
	Kind              string
	Title             string
	Description       string
	Canonical         string
	Robots            string
	CacheControl      string
	Status            int
	Message           string
	Path              string
	Form              map[string][]string
	Context           value.RequestContext
	Search            application.SearchResult
	Query             application.SearchQuery
	Category          application.CategoryView
	Detail            application.ProductDetail
	Recommendations   application.RecommendationResult
	Cart              application.CartView
	Order             application.OrderView
	Guest             application.GuestView
	CSRF              string
	CommandID         value.CommandID
	AdminCatalog      application.AdminCatalogSnapshot
	AdminOrders       application.AdminOrderPage
}

type PaginationLink struct {
	URL     string
	Number  int
	Current bool
}

type Fragment struct {
	URL string
}

type CacheInvalidation interface {
	Invalidate(context.Context, []value.SKU) error
}

type CacheInvalidator struct {
	request *http.Request
	key     string
	client  *http.Client
}

type InputField struct {
	Name   string
	Value  string
	Label  string
	Hidden bool
}

type Diagnostics interface {
	Snapshot(context.Context) (map[string]string, error)
}

type systemClock struct{}

type InterestReader interface {
	GetInterests(context.Context, application.Session) ([]sale.CategoryInterest, error)
}

type PublicRoute struct {
	Path        string
	Title       string
	Description string
	Kind        string
}

type VariantReader interface {
	ProductForSKU(context.Context, value.SKU, value.RequestContext) (application.ProductDetail, error)
}

type AccessoryPreview interface {
	PreviewAccessory(context.Context, application.Session, value.CartID, value.SKU) (sale.CartEvaluation, error)
}

type frameReference struct {
	Kind     string
	ID       string
	Revision uint64
}

type querySelection []string

type Renderer struct {
	templates *template.Template
}

type requestView struct {
	request *http.Request
}

type routePath string

type responseError struct {
	cause error
}

type cacheWindow struct {
	now time.Time
}

type queryInput struct {
	values url.Values
}

type inventoryInput struct {
	values url.Values
}

type cardView struct {
	application.ProductCard
	ImageAlt   string
	Price      string
	BasePrice  string
	Attributes []cardAttributeView
}

type cardAttributeView struct {
	Code value.AttributeCode
	Text string
}
