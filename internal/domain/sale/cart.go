package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

const (
	Open        CartState = "open"
	CheckingOut CartState = "checkingOut"
	Ordered     CartState = "ordered"
	Expired     CartState = "expired"
)

func NewCart(id value.CartID, user value.UserID, now time.Time) (Cart, error) {
	if value.IdentifierSyntax(string(id)).Validate() != nil {
		return Cart{}, value.NewError("InvalidInput", "CartID")
	}

	if value.IdentifierSyntax(string(user)).Validate() != nil {
		return Cart{}, value.NewError("InvalidInput", "UserID")
	}

	return Cart{ID: id, UserID: user, LastActivityAt: now, Revision: 1, State: Open, DefinitionVersion: "cart-v1", RuleSetVersion: "demo-v1"}, nil
}

func (cart Cart) CanChange(revision uint64, now time.Time) error {
	if cart.State == Ordered {
		return value.NewError("AlreadyOrdered", "Cart")
	}

	if cart.IsExpired(now) {
		return value.NewError("Expired", "Cart")
	}

	if cart.Revision != revision {
		return value.NewError("Conflict", "Revision")
	}

	return nil
}

func (cart Cart) IsExpired(now time.Time) bool {
	if cart.State == Expired {
		return true
	}

	return !now.Before(cart.LastActivityAt.Add(168 * time.Hour))
}

func (items CartItems) Quantity(sku value.SKU) int64 {
	for _, item := range items {
		if item.SKU == sku {
			return item.Quantity
		}
	}

	return 0
}

func (cart Cart) Add(sku value.SKU, quantity int64, revision uint64, now time.Time) (Cart, error) {
	if quantity <= 0 {
		return cart, value.NewError("InvalidInput", "Quantity")
	}

	return cart.SetQuantity(sku, cart.Items.Quantity(sku)+quantity, revision, now)
}

func (cart Cart) SetQuantity(sku value.SKU, quantity int64, revision uint64, now time.Time) (Cart, error) {
	if err := cart.CanChange(revision, now); err != nil {
		return cart, err
	}

	if err := cart.checkItem(sku, quantity); err != nil {
		return cart, err
	}

	cart.Items = cart.Items.without(sku)
	cart.Items = append(cart.Items, CartItem{SKU: sku, Quantity: quantity})

	return cart.changed(now), nil
}

func (cart Cart) checkItem(sku value.SKU, quantity int64) error {
	if sku == "" {
		return value.NewError("InvalidInput", "SKU")
	}

	if quantity < 1 {
		return value.NewError("InvalidInput", "Quantity")
	}

	if quantity > 99 {
		return value.NewError("InvalidInput", "Quantity")
	}

	if cart.isFullFor(sku) {
		return value.NewError("InvalidInput", "Items")
	}

	return nil
}

func (cart Cart) isFullFor(sku value.SKU) bool {
	return len(cart.Items) >= 50 && cart.Items.Quantity(sku) == 0
}

func (items CartItems) without(sku value.SKU) CartItems {
	result := CartItems{}
	for _, item := range items {
		if item.SKU != sku {
			result = append(result, item)
		}
	}

	return result
}

func (cart Cart) Remove(sku value.SKU, revision uint64, now time.Time) (Cart, error) {
	if err := cart.CanChange(revision, now); err != nil {
		return cart, err
	}

	cart.Items = cart.Items.without(sku)

	return cart.changed(now), nil
}

func (cart Cart) changed(now time.Time) Cart {
	cart.Revision++
	cart.LastActivityAt = now
	cart.Evaluation = value.Optional[CartEvaluation]{}
	cart.Checkout.Value.ConfirmedEvaluation = value.Optional[CartEvaluation]{}

	return cart
}

func (cart Cart) ConfigureCheckout(checkout Checkout, revision uint64, now time.Time) (Cart, error) {
	if err := cart.CanChange(revision, now); err != nil {
		return cart, err
	}

	if err := checkout.Validate(); err != nil {
		return cart, err
	}

	checkout.CartID = cart.ID
	checkout.ConfirmedEvaluation = value.Optional[CartEvaluation]{}
	cart.Checkout = value.Optional[Checkout]{Value: checkout, Present: true}
	cart.State = CheckingOut

	return cart.changed(now), nil
}

func (cart Cart) SetEvaluation(evaluation CartEvaluation) (Cart, error) {
	if evaluation.CartRevision != cart.Revision {
		return cart, value.NewError("Conflict", "CartRevision")
	}

	if evaluation.RuleSetVersion != cart.RuleSetVersion {
		return cart, value.NewError("Conflict", "RuleSetVersion")
	}

	cart.Evaluation = value.Optional[CartEvaluation]{Value: evaluation.Clone(), Present: true}

	return cart, nil
}

func (cart Cart) Confirm(id value.EvaluationID, revision uint64, now time.Time) (Cart, error) {
	if err := cart.CanChange(revision, now); err != nil {
		return cart, err
	}

	if !cart.confirmable(id) {
		return cart, value.NewError("PriceConfirmationRequired", "EvaluationID")
	}

	cart.Checkout.Value.ConfirmedEvaluation = value.Optional[CartEvaluation]{Value: cart.Evaluation.Value.Clone(), Present: true}
	cart.LastActivityAt = now
	cart.State = CheckingOut

	return cart, nil
}

func (cart Cart) confirmable(id value.EvaluationID) bool {
	return cart.Checkout.Present && cart.Evaluation.Present && cart.Evaluation.Value.ID == id
}

func (cart Cart) Expire(now time.Time) (Cart, error) {
	if cart.State == Ordered {
		return cart, value.NewError("AlreadyOrdered", "Cart")
	}

	if cart.IsExpired(now) {
		cart.State = Expired
		cart.Evaluation = value.Optional[CartEvaluation]{}
		cart.Checkout.Value.ConfirmedEvaluation = value.Optional[CartEvaluation]{}
	}

	return cart, nil
}

func (cart Cart) MarkOrdered() (Cart, error) {
	if err := cart.CanChange(cart.Revision, cart.LastActivityAt); err != nil {
		return cart, err
	}

	if !cart.Checkout.Value.ConfirmedEvaluation.Present {
		return cart, value.NewError("PriceConfirmationRequired", "Evaluation")
	}

	cart.State = Ordered
	cart.Revision++

	return cart, nil
}
