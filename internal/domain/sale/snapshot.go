package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

func (discount AppliedDiscount) clone() AppliedDiscount {
	discount.ItemSKUs = append([]value.SKU{}, discount.ItemSKUs...)

	return discount
}

func (discounts AppliedDiscounts) Clone() []AppliedDiscount {
	result := make([]AppliedDiscount, len(discounts))
	for index, discount := range discounts {
		result[index] = discount.clone()
	}

	return result
}

func (evaluation CartEvaluation) Clone() CartEvaluation {
	evaluation.Items = append([]EvaluatedCartItem{}, evaluation.Items...)
	for index, item := range evaluation.Items {
		evaluation.Items[index].Discounts = AppliedDiscounts(item.Discounts).Clone()
	}

	revisions := map[value.SKU]uint64{}
	for sku, revision := range evaluation.InventoryRevisions {
		revisions[sku] = revision
	}

	evaluation.InventoryRevisions = revisions
	evaluation.Discounts = AppliedDiscounts(evaluation.Discounts).Clone()
	evaluation.Taxes = append([]TaxAmount{}, evaluation.Taxes...)

	return evaluation
}

func NewOrder(id value.OrderID, cart Cart, items []OrderItem, now time.Time) (Order, error) {
	if err := cart.validatePurchase(id, items, now); err != nil {
		return Order{}, err
	}

	return cart.purchaseOrder(id, items, now), nil
}

func NewPurchase(id value.OrderID, cart Cart, items []OrderItem, now time.Time) (Order, Cart, error) {
	if err := cart.validatePurchase(id, items, now); err != nil {
		return Order{}, cart, err
	}

	order := cart.purchaseOrder(id, items, now)
	cart.State = Ordered
	cart.Revision++

	return order, cart, nil
}

func (cart Cart) purchaseOrder(id value.OrderID, items []OrderItem, now time.Time) Order {
	checkout := cart.Checkout.Value
	evaluation := checkout.ConfirmedEvaluation.Value
	order := Order{ID: id, CartID: cart.ID, UserID: cart.UserID, PlacedAt: now, Items: items,
		DefinitionVersion: "order-v1", RuleSetVersion: cart.RuleSetVersion, ShippingAddress: checkout.ShippingAddress,
		ShippingMethod: checkout.ShippingMethod, PaymentMethod: checkout.PaymentMethod, Subtotal: evaluation.Subtotal,
		Discounts: evaluation.Discounts, ShippingCost: evaluation.ShippingCost, Taxes: evaluation.Taxes, Total: evaluation.Total}

	return order.Snapshot()
}

func (cart Cart) validatePurchase(id value.OrderID, items []OrderItem, now time.Time) error {
	if err := cart.CanChange(cart.Revision, now); err != nil {
		return err
	}

	if value.IdentifierSyntax(string(id)).Validate() != nil {
		return value.NewError("InvalidInput", "OrderID")
	}

	if len(items) == 0 {
		return value.NewError("InvalidInput", "Items")
	}

	if !cart.Checkout.Value.ConfirmedEvaluation.Present {
		return value.NewError("PriceConfirmationRequired", "Evaluation")
	}

	return cart.validatePurchaseItems(items)
}

func (cart Cart) validatePurchaseItems(items []OrderItem) error {
	evaluation := cart.Checkout.Value.ConfirmedEvaluation.Value
	if len(items) != len(evaluation.Items) {
		return value.NewError("Conflict", "Items")
	}

	for index, item := range items {
		if !item.matchesEvaluation(evaluation.Items[index]) {
			return value.NewError("Conflict", "Items")
		}
	}

	return nil
}

func (item OrderItem) matchesEvaluation(evaluated EvaluatedCartItem) bool {
	return item.SKU == evaluated.SKU && item.Quantity == evaluated.Quantity && item.UnitPrice == evaluated.UnitPrice && item.matchesDiscounts(evaluated.Discounts)
}

func (item OrderItem) matchesDiscounts(discounts []AppliedDiscount) bool {
	return AppliedDiscounts(item.Discounts).Matches(discounts)
}

func (order Order) Snapshot() Order {
	order.Items = append([]OrderItem{}, order.Items...)
	for index, item := range order.Items {
		order.Items[index].VariantAttributes = item.VariantAttributes.Clone()
		order.Items[index].Discounts = AppliedDiscounts(item.Discounts).Clone()
	}

	order.Discounts = AppliedDiscounts(order.Discounts).Clone()
	order.Taxes = append([]TaxAmount{}, order.Taxes...)

	return order
}
