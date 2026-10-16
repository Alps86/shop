package postgres_test

import (
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

type world struct {
	previousGuest      application.GuestView
	permissionDatabase string
	permissionRole     string
	originalOrder      sale.Order
	processCarts       map[value.CartID]bool
	processOrders      map[value.OrderID]bool
	processFailure     string
	observed           map[string]string
	paymentOutcome     string
	commands           []application.PlaceCommand
	results            []application.PurchaseResult
	errors             []error
	store              *postgres.Store
	admin              *pgx.Conn
	schema             string
	manifest           postgres.ImportManifest
	imported           postgres.ImportResult
	service            application.Service
	guest              application.GuestView
	cart               application.CartView
	result             application.PurchaseResult
	inventoryPrice     int64
	err                error
}

type processObserver struct{ current *world }
