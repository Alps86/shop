package postgres

import (
	"context"
	"io"
	"time"

	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	ReplicationSlot string
	Entropy         io.Reader
	DSN             string
	Schema          string
	Context         value.RequestContext
}

type Store struct {
	pool   *pgxpool.Pool
	config Config
}

type database interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type record struct {
	kind  string
	id    string
	value any
}

type GuestRepository struct {
	store *Store
}

type CartRepository struct {
	store *Store
}

type OrderRepository struct {
	store *Store
}

type FulfillmentRepository struct {
	store *Store
}

type CatalogRepository struct {
	store *Store
}

type InventoryRepository struct {
	store *Store
}

type EvaluationRepository struct {
	store *Store
}

type PurchaseRepository struct {
	store *Store
}

type RecommendationRepository struct {
	store *Store
}

type AdminOrderRepository struct {
	store *Store
}

type ImportManifest struct {
	DatasetVersion       string
	SourceURL            string
	RetrievedAt          time.Time
	LicenseNote          string
	Products             []catalog.Product
	Families             []catalog.ProductFamily
	Categories           []catalog.Category
	Inventories          []sale.Inventory
	DiscountRules        []sale.DiscountRule
	CoPurchaseStatistics application.PairStatistics
}

type ImportResult struct {
	Code           string
	DatasetVersion string
	ProductCount   int
}

type inventoryState map[value.SKU]sale.Inventory

type purchaseAttempt struct {
	orderID    value.OrderID
	ordered    sale.Cart
	repository PurchaseRepository
	tx         pgx.Tx
	command    application.PurchaseCommand
	cart       sale.Cart
	evaluation sale.CartEvaluation
	inventory  inventoryState
	products   map[value.SKU]catalog.Product
	result     application.PurchaseResult
	rejection  error
}

type commandRecord struct {
	scope   string
	id      value.CommandID
	payload any
	result  any
}

type ProcessDispatcher struct {
	Store     *Store
	Processes application.ProcessPort
}

type pendingProcess struct {
	sequence int64
	envelope event.ProcessEnvelope
}

type versionState struct {
	Version string
}

type startCommand struct {
	UserID value.UserID
}

type placeFingerprint struct {
	UserID       value.UserID
	CartID       value.CartID
	EvaluationID value.EvaluationID
}

type projectionVersion struct {
	id       value.ProductID
	revision uint64
	deleted  bool
}

type ReplicationState struct {
	Exists    bool
	Restart   uint64
	Confirmed uint64
	Status    string
}
