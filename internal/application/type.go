package application

import (
	"context"
	"time"

	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

type Session string

type ProfileCode string

type SessionIdentity struct {
	UserID   value.UserID
	IssuedAt time.Time
	CartID   value.Optional[value.CartID]
	Profile  ProfileCode
}

type GuestView struct {
	UserID  value.UserID
	CartID  value.Optional[value.CartID]
	Profile ProfileCode
	Session Session
}

type GuestStore interface {
	Resolve(context.Context, Session) (SessionIdentity, error)
	Open(context.Context) (GuestView, error)
	Profile(context.Context, value.UserID, ProfileCode) (GuestView, error)
}

type CartStore interface {
	Get(context.Context, value.CartID) (sale.Cart, error)
}

type OrderStore interface {
	Get(context.Context, value.OrderID) (sale.Order, error)
	ForCart(context.Context, value.CartID) (value.OrderID, error)
}

type Clock interface {
	Now() time.Time
}

type Service struct {
	Guests      GuestStore
	Carts       CartStore
	Orders      OrderStore
	Clock       Clock
	Purchase    PurchaseTransaction
	Payment     PaymentPort
	Processes   ProcessPort
	Writer      CartWriter
	Evaluator   EvaluationPort
	Fulfillment FulfillmentStore
	Starter     CartStarter
}

type CartView struct {
	OrderID value.Optional[value.OrderID]
	Cart    sale.Cart
	Actions []string
}

type OrderView struct {
	Purchase    sale.Order
	Fulfillment FulfillmentView
}

type FulfillmentView struct {
	State string
	Lines []DeliveryLine
}

type DeliveryLine struct {
	SKU       value.SKU
	Purchased int64
	Shipped   int64
	Delivered int64
}

type ChangeCommand struct {
	UserID           value.UserID
	CartID           value.CartID
	CommandID        value.CommandID
	ExpectedRevision uint64
	Operation        string
	SKU              value.SKU
	Quantity         int64
	ShippingMethod   sale.ShippingMethod
	PaymentMethod    sale.PaymentMethod
	Address          sale.Address
}

type PlaceCommand struct {
	UserID       value.UserID
	CartID       value.CartID
	CommandID    value.CommandID
	EvaluationID value.EvaluationID
}

type PurchaseCommand struct {
	UserID        value.UserID
	CartID        value.CartID
	CommandID     value.CommandID
	EvaluationID  value.EvaluationID
	Authorization value.Optional[PaymentAuthorization]
	Now           time.Time
}

type PurchaseResult struct {
	OrderID          value.OrderID
	ProcessReference string
	Pending          bool
	Cart             value.Optional[CartView]
	UnavailableSKU   value.Optional[value.SKU]
	Alternatives     []ProductCard
}

type PurchaseTransaction interface {
	Commit(context.Context, PurchaseCommand) (PurchaseResult, error)
}

type PaymentAuthorization struct {
	CartID       value.CartID
	CartRevision uint64
	Total        value.Money
	CommandID    value.CommandID
	Approved     bool
}

type PaymentPort interface {
	Authorize(context.Context, sale.Cart, value.CommandID) (PaymentAuthorization, error)
}

type ProcessBinding struct {
	DefinitionVersion string
	RuleSetVersion    string
}

type ProcessPort interface {
	EnsureCart(context.Context, value.CartID, ProcessBinding) error
	UpdateCart(context.Context, ChangeCommand) (CartView, error)
	EvaluateCart(context.Context, EvaluateCommand) (CartView, error)
	ConfirmEvaluation(context.Context, ConfirmCommand) (CartView, error)
	PlaceOrder(context.Context, PlaceCommand) (PurchaseResult, error)
	EnsureOrder(context.Context, value.OrderID, ProcessBinding) error
	ReportDelivery(context.Context, DeliveryCommand) error
}

type DeliveryCommand struct {
	OrderID   value.OrderID
	CommandID value.CommandID
	Update    sale.FulfillmentUpdate
}

type ProductCard struct {
	ProductID         value.ProductID
	URL               string
	Name              string
	SKU               value.SKU
	Image             catalog.Image
	UnitPrice         value.Money
	AvailableQuantity int64
	VariantCount      int
	TopAttributes     map[value.AttributeCode]catalog.AttributeValue
	Compliance        catalog.ProductCompliance
}

type ProductDetail struct {
	Product   catalog.Product
	ChosenSKU value.SKU
	Cards     []ProductCard
}

type CatalogReader interface {
	Product(context.Context, value.ProductID, value.Optional[value.SKU], value.RequestContext) (ProductDetail, error)
	Category(context.Context, value.CategoryID, value.RequestContext) (CategoryView, error)
}

type Sort string

type FilterSet struct {
	Operators []string
	Brands    []string
	Options   map[value.AttributeCode][]value.OptionCode
	Ranges    map[value.AttributeCode]NumericRange
	Price     value.Optional[MoneyRange]
}

type NumericRange struct {
	Minimum value.Optional[value.Decimal]
	Maximum value.Optional[value.Decimal]
}

type MoneyRange struct {
	Minimum value.Optional[value.Money]
	Maximum value.Optional[value.Money]
}

type SearchQuery struct {
	Text       string
	CategoryID value.Optional[value.CategoryID]
	Filters    FilterSet
	Sort       Sort
	Page       int
	Context    value.RequestContext
	Interests  []sale.CategoryInterest
}

type FacetValue struct {
	Value        string
	ProductCount int64
	Selected     bool
}

type Facet struct {
	Code   string
	Values []FacetValue
}

type Suggestion struct {
	Text       string
	Correction bool
}

type SearchResult struct {
	TotalProducts int64
	Cards         []ProductCard
	Facets        []Facet
	Page          int
	PageSize      int
	Suggestions   []Suggestion
	Degraded      bool
}

type SearchPort interface {
	Search(context.Context, SearchQuery) (SearchResult, error)
}

type ProjectionEvent struct {
	ProductID value.ProductID
	Revision  uint64
	Product   value.Optional[catalog.Product]
	Inventory []sale.Inventory
}

type ProjectionPort interface {
	Apply(context.Context, ProjectionEvent) error
	Rebuild(context.Context, string, time.Time) error
}

type InventoryCommand struct {
	SKU              value.SKU
	ExpectedRevision uint64
	Price            value.Optional[value.Money]
	Stock            value.Optional[int64]
	CommandID        value.CommandID
}

type InventoryService interface {
	Change(context.Context, InventoryCommand) (sale.Inventory, error)
}

type RecommendationQuery struct {
	ProductID value.ProductID
	SKU       value.SKU
	Context   value.RequestContext
	Interests []sale.CategoryInterest
	Limit     int
}

type RecommendationResult struct {
	Cards    []ProductCard
	Label    string
	Degraded bool
}

type RecommendationPort interface {
	Recommend(context.Context, RecommendationQuery) (RecommendationResult, error)
}

type RecommendationSource interface {
	Candidates(context.Context, RecommendationQuery) ([]RecommendationCandidate, error)
	Statistics(context.Context, value.ProductID) (PairStatistics, error)
}

type RecommendationCandidate struct {
	Card       ProductCard
	CategoryID value.CategoryID
	Compatible bool
}

type PairStatistic struct {
	SourceProductID value.ProductID
	TargetProductID value.ProductID
	BasketCount     int64
	WindowStart     time.Time
	WindowEnd       time.Time
	DatasetVersion  string
}

type PairStatistics []PairStatistic

type RecommendationService struct {
	Source RecommendationSource
}

type rankedRecommendation struct {
	Candidate RecommendationCandidate
	Score     int64
}

type rankedRecommendations []rankedRecommendation

type SearchService struct {
	Searcher SearchPort
}

type CartWrite struct {
	Change           value.Optional[ChangeCommand]
	Confirmation     value.Optional[ConfirmCommand]
	Cart             sale.Cart
	ExpectedRevision uint64
	CommandID        value.CommandID
}

type CartWriter interface {
	Save(context.Context, CartWrite) (CartView, error)
}

type EvaluationPort interface {
	Evaluate(context.Context, sale.Cart) (sale.CartEvaluation, error)
}

type ConfirmCommand struct {
	UserID           value.UserID
	CartID           value.CartID
	EvaluationID     value.EvaluationID
	ExpectedRevision uint64
	CommandID        value.CommandID
}

type FulfillmentStore interface {
	Get(context.Context, value.OrderID) (sale.Fulfillment, error)
	Save(context.Context, sale.Fulfillment, value.CommandID) error
}

type FulfillmentService struct {
	Store FulfillmentStore
}

type CartStarter interface {
	Start(context.Context, value.UserID, value.CommandID, time.Time) (CartView, error)
}

type EvaluateCommand struct {
	UserID value.UserID
	CartID value.CartID
}

type SearchSchema struct {
	Attributes map[value.AttributeCode]catalog.AttributeDefinition
}

type CategoryView struct {
	Category          catalog.Category
	Breadcrumb        []catalog.Category
	Children          []catalog.Category
	FilterDefinitions []catalog.AttributeDefinition
	TopDefinitions    []catalog.AttributeDefinition
}

type AdminOrderPage struct {
	Orders   []OrderView
	Page     int
	PageSize int
	Total    int64
}

type AdminOrderReader interface {
	List(context.Context, int) (AdminOrderPage, error)
	Get(context.Context, value.OrderID) (OrderView, error)
}

type AdminCatalogSnapshot struct {
	Products      []catalog.Product
	Families      []catalog.ProductFamily
	Categories    []catalog.Category
	DiscountRules []sale.DiscountRule
	Inventory     []sale.Inventory
}

type AdminCatalogReader interface {
	Snapshot(context.Context) (AdminCatalogSnapshot, error)
}
