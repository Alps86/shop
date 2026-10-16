package http_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (double inventoryDouble) Change(_ context.Context, command application.InventoryCommand) (sale.Inventory, error) {
	double.current.lastInventory = command
	return sale.Inventory{}, double.current.portError("Inventory")
}

func (double adminOrderDouble) List(context.Context, int) (application.AdminOrderPage, error) {
	return double.current.fixture.AdminOrders, double.current.portError("ListOrders")
}

func (double adminOrderDouble) Get(context.Context, value.OrderID) (application.OrderView, error) {
	return double.current.fixture.Order, double.current.portError("AdminGetOrder")
}

func (double processDouble) ReportDelivery(_ context.Context, command application.DeliveryCommand) error {
	double.current.lastDelivery = command
	return double.current.portError("ReportDelivery")
}

func (double processDouble) EnsureCart(context.Context, value.CartID, application.ProcessBinding) error {
	return nil
}

func (double processDouble) EnsureOrder(context.Context, value.OrderID, application.ProcessBinding) error {
	return nil
}

func (double processDouble) UpdateCart(context.Context, application.ChangeCommand) (application.CartView, error) {
	return application.CartView{}, nil
}

func (double processDouble) EvaluateCart(context.Context, application.EvaluateCommand) (application.CartView, error) {
	return application.CartView{}, nil
}

func (double processDouble) ConfirmEvaluation(context.Context, application.ConfirmCommand) (application.CartView, error) {
	return application.CartView{}, nil
}

func (double processDouble) PlaceOrder(context.Context, application.PlaceCommand) (application.PurchaseResult, error) {
	return application.PurchaseResult{}, nil
}

func (double diagnosticDouble) Snapshot(context.Context) (map[string]string, error) {
	return double.current.fixture.Diagnostics, double.current.portError("Diagnostics")
}
