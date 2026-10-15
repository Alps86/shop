package application

import (
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func NewCartView(cart sale.Cart) CartView {
	actions := map[sale.CartState][]string{
		sale.Open:        {"change", "evaluate"},
		sale.CheckingOut: {"change", "evaluate"},
		sale.Ordered:     {"viewOrder", "startCart"},
		sale.Expired:     {"startCart"},
	}

	view := CartView{Cart: cart, Actions: actions[cart.State]}
	if view.canConfirm() {
		view.Actions = append(view.Actions, "confirm")
	}

	if view.canPlace() {
		view.Actions = append(view.Actions, "place")
	}

	return view
}

func (view CartView) canConfirm() bool {
	return view.Cart.State == sale.CheckingOut && view.Cart.Checkout.Present && view.Cart.Evaluation.Present && !view.Cart.Checkout.Value.ConfirmedEvaluation.Present
}

func (view CartView) canPlace() bool {
	return view.Cart.State == sale.CheckingOut && view.Cart.Checkout.Value.ConfirmedEvaluation.Present
}

func NewCartPurchaseResult(cart sale.Cart) PurchaseResult {
	return PurchaseResult{Cart: value.Optional[CartView]{Present: true, Value: NewCartView(cart)}}
}
