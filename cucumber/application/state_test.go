package application_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (port evaluator) Evaluate(_ context.Context, cart sale.Cart) (sale.CartEvaluation, error) {
	if port.current.fault == "evaluationRead" {
		return sale.CartEvaluation{}, value.NewError("DependencyUnavailable", "Evaluation")
	}

	items := port.priced(cart)

	evaluation, err := items.Evaluate(nil, port.shipping(cart))
	evaluation.ID = "e-a"
	evaluation.CartRevision = cart.Revision
	evaluation.RuleSetVersion = cart.RuleSetVersion
	if port.current.fault == "evaluationRevision" {
		evaluation.CartRevision++
	}

	return evaluation, err
}

func (port cartWriter) Save(_ context.Context, write application.CartWrite) (application.CartView, error) {
	if port.current.fault == "cartWrite" {
		return application.CartView{}, value.NewError("DependencyUnavailable", "Cart")
	}

	port.current.carts[write.Cart.ID] = write.Cart

	return application.NewCartView(write.Cart), nil
}

func (port evaluator) priced(cart sale.Cart) sale.PricedItems {
	items := sale.PricedItems{}
	for _, item := range cart.Items {
		inventory := port.current.inventories[item.SKU]
		items = append(items, sale.PricedItem{SKU: item.SKU, ProductID: "p-phone", Quantity: item.Quantity, UnitPrice: inventory.Price, InventoryRevision: inventory.Revision})
	}

	return items
}

func (port evaluator) shipping(cart sale.Cart) sale.ShippingMethod {
	if !cart.Checkout.Present {
		return sale.Standard
	}

	return cart.Checkout.Value.ShippingMethod
}
