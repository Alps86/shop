package commerce_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/runtime"
)

type row map[string]string

type world struct {
	answers  map[string]map[string]string
	cartID   string
	app      *runtime.App
	config   runtime.Config
	cancel   context.CancelFunc
	finished chan error
	guest    application.GuestView
	view     application.CartView
	purchase application.PurchaseResult
	codes    map[string]string
	err      error
}
