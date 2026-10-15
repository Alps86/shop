package application_test

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"sync"
	"time"
)

type world struct {
	preview               sale.CartEvaluation
	previewRules          sale.DiscountRules
	interestData          map[value.UserID][]sale.CategoryInterest
	interestResult        []sale.CategoryInterest
	interestUser          value.UserID
	commandWrites         map[value.CommandID]application.CartWrite
	inventoryProducts     map[value.SKU]value.ProductID
	coordinationOperation string
	processUser           value.UserID
	processPayload        string
	atomicDefaults        map[string]string
	commitMutex           sync.Mutex
	orderOutbox           []value.OrderID
	atomicCommands        []application.PlaceCommand
	atomicErrors          []error
	atomicResults         []application.PurchaseResult

	initialOrder    sale.Order
	processCalls    int
	service         application.Service
	now             time.Time
	sessions        map[application.Session]application.SessionIdentity
	profiles        map[value.UserID]application.ProfileCode
	carts           map[value.CartID]sale.Cart
	orders          map[value.OrderID]sale.Order
	inventories     map[value.SKU]sale.Inventory
	result          application.GuestView
	cartView        application.CartView
	orderView       application.OrderView
	err             error
	source          recommendationSource
	recommendations application.RecommendationResult
	query           application.RecommendationQuery
	searchQuery     application.SearchQuery
	searchResult    application.SearchResult
	fault           string
	purchaseRecords map[value.CommandID]purchaseRecord
	placeCommand    application.PlaceCommand
	purchaseResult  application.PurchaseResult
	authorizations  int
	authorization   string
	fulfillment     sale.Fulfillment
	fulfillmentView application.FulfillmentView
}

type rows []map[string]string

type recommendationSource struct {
	candidates []application.RecommendationCandidate
	statistics application.PairStatistics
	fault      string
}

type orderStore struct {
	current *world
}

type purchaseRecord struct {
	command application.PurchaseCommand
	order   value.OrderID
}

type evaluator struct {
	current *world
}

type cartWriter struct {
	current *world
}

type deliveryStore struct {
	current *world
}

type cartStarter struct {
	current *world
}

type processDouble struct {
	current *world
}

type commandReader struct{ current *world }

type interestReader struct{ current *world }
