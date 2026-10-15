package application_test

import (
	"context"
	"reflect"
	"slices"
	"strings"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

// Dieser Testdouble erfüllt ausschließlich den vereinbarten Transaktionsport.
// PostgreSQL-Sperren, Outbox und Temporal werden erst im Runtimepaket belegt.
func (current *world) Commit(ctx context.Context, command application.PurchaseCommand) (application.PurchaseResult, error) {
	current.commitMutex.Lock()
	defer current.commitMutex.Unlock()
	result, err := current.commitLocked(ctx, command)

	return current.commitResponse(result, err)
}

func (current *world) commitLocked(ctx context.Context, command application.PurchaseCommand) (application.PurchaseResult, error) {
	if previous, exists := current.purchaseRecords[command.CommandID]; exists {
		return previous.repeat(command)
	}

	cart := current.carts[command.CartID]
	if cart.UserID != command.UserID {
		return application.PurchaseResult{}, value.NewError("NotFound", "Cart")
	}

	if cart.State == sale.Ordered {
		return application.PurchaseResult{OrderID: orderID(cart.ID)}, value.NewError("AlreadyOrdered", "Cart")
	}

	if err := cart.CanChange(cart.Revision, command.Now); err != nil {
		return application.PurchaseResult{}, err
	}

	return current.commitNew(ctx, cart, command)
}

func (current *world) commitResponse(result application.PurchaseResult, err error) (application.PurchaseResult, error) {
	if err != nil {
		return result, err
	}

	if current.fault == "afterCommitBeforeResponse" {
		return application.PurchaseResult{}, value.NewError("DependencyUnavailable", "Response")
	}

	if current.fault == "afterCommitBeforeOrderStart" {
		return application.PurchaseResult{}, value.NewError("DependencyUnavailable", "OrderStart")
	}

	return result, nil
}

func orderID(cart value.CartID) value.OrderID {
	return value.OrderID("o-" + strings.TrimPrefix(string(cart), "c-"))
}

func (record purchaseRecord) repeat(command application.PurchaseCommand) (application.PurchaseResult, error) {
	previous := record.command
	previous.Now = command.Now
	previous.Authorization = command.Authorization
	if !reflect.DeepEqual(previous, command) {
		return application.PurchaseResult{}, value.NewError("Conflict", "CommandID")
	}

	return application.PurchaseResult{OrderID: record.order}, nil
}

func (current *world) commitNew(ctx context.Context, cart sale.Cart, command application.PurchaseCommand) (application.PurchaseResult, error) {
	if current.fault == "beforeCommit" {
		return application.PurchaseResult{}, value.NewError("DependencyUnavailable", "Transaction")
	}

	evaluation, err := (evaluator{current}).Evaluate(ctx, cart)
	if err != nil {
		return application.PurchaseResult{}, err
	}

	if !evaluation.MatchesPrices(cart.Checkout.Value.ConfirmedEvaluation.Value) {
		return current.requirePriceConfirmation(cart, evaluation)
	}

	return current.reserve(cart, command)
}

func (current *world) reserve(cart sale.Cart, command application.PurchaseCommand) (application.PurchaseResult, error) {
	reserved := map[value.SKU]sale.Inventory{}
	for _, item := range cart.Items {
		inventory, err := current.inventories[item.SKU].Reserve(orderID(cart.ID), item.Quantity, current.inventories[item.SKU].Revision)
		if err != nil {
			return current.reservationFailure(item.SKU, err)
		}

		reserved[item.SKU] = inventory
		if current.fault == "transactionRollback" {
			return application.PurchaseResult{}, value.NewError("DependencyUnavailable", "Transaction")
		}
	}

	return current.finishPurchase(cart, command, reserved)
}

func (current *world) finishPurchase(cart sale.Cart, command application.PurchaseCommand, reserved map[value.SKU]sale.Inventory) (application.PurchaseResult, error) {
	order, err := sale.NewOrder(orderID(cart.ID), cart, current.orderItems(cart), command.Now)
	if err != nil {
		return application.PurchaseResult{}, err
	}

	ordered, err := cart.MarkOrdered()
	if err != nil {
		return application.PurchaseResult{}, err
	}

	return current.persistPurchase(order, ordered, reserved, command)
}

func (current *world) persistPurchase(order sale.Order, cart sale.Cart, reserved map[value.SKU]sale.Inventory, command application.PurchaseCommand) (application.PurchaseResult, error) {
	current.orders[order.ID] = order
	current.carts[cart.ID] = cart
	for sku, inventory := range reserved {
		current.inventories[sku] = inventory
	}

	current.purchaseRecords[command.CommandID] = purchaseRecord{command: command, order: order.ID}
	current.orderOutbox = append(current.orderOutbox, order.ID)

	return application.PurchaseResult{OrderID: order.ID}, nil
}

func (current *world) orderItems(cart sale.Cart) []sale.OrderItem {
	items := []sale.OrderItem{}
	for _, item := range cart.Checkout.Value.ConfirmedEvaluation.Value.Items {
		items = append(items, sale.OrderItem{SKU: item.SKU, ProductName: "Testmodell", Quantity: item.Quantity, UnitPrice: item.UnitPrice, Discounts: item.Discounts})
	}

	return items
}

func (current *world) Authorize(_ context.Context, cart sale.Cart, id value.CommandID) (application.PaymentAuthorization, error) {
	current.authorizations++
	if current.authorization == "technical" {
		return application.PaymentAuthorization{}, value.NewError("DependencyUnavailable", "Payment")
	}

	authorization := application.PaymentAuthorization{CartID: cart.ID, CartRevision: cart.Revision, Total: cart.Checkout.Value.ConfirmedEvaluation.Value.Total, CommandID: id, Approved: current.authorization != "declined"}
	if current.authorization == "mismatch" {
		authorization.Total.MinorUnits++
	}

	return authorization, nil
}

func (current *world) requirePriceConfirmation(cart sale.Cart, evaluation sale.CartEvaluation) (application.PurchaseResult, error) {
	cart.Checkout.Value.ConfirmedEvaluation = value.Optional[sale.CartEvaluation]{}
	cart.Evaluation = value.Optional[sale.CartEvaluation]{Value: evaluation, Present: true}
	current.carts[cart.ID] = cart

	return application.PurchaseResult{Cart: value.Optional[application.CartView]{Value: application.CartView{Cart: cart}, Present: true}}, value.NewError("PriceConfirmationRequired", "Total")
}

func (current *world) reservationFailure(sku value.SKU, err error) (application.PurchaseResult, error) {
	if value.Code(err) != "Unavailable" {
		return application.PurchaseResult{}, err
	}

	result := application.PurchaseResult{UnavailableSKU: value.Optional[value.SKU]{Value: sku, Present: true}, Alternatives: current.availableAlternatives(sku)}

	return result, err
}

func (current *world) availableAlternatives(unavailable value.SKU) []application.ProductCard {
	alternatives := []application.ProductCard{}
	for sku, inventory := range current.inventories {
		if current.isAlternative(unavailable, sku, inventory) {
			alternatives = append(alternatives, application.ProductCard{ProductID: current.inventoryProducts[sku], SKU: sku, UnitPrice: inventory.Price, AvailableQuantity: inventory.AvailableQuantity()})
		}
	}

	slices.SortFunc(alternatives, func(left, right application.ProductCard) int {
		return strings.Compare(string(left.SKU), string(right.SKU))
	})

	return alternatives
}

func (current *world) isAlternative(unavailable, sku value.SKU, inventory sale.Inventory) bool {
	return sku != unavailable && current.inventoryProducts[unavailable] != "" && current.inventoryProducts[unavailable] == current.inventoryProducts[sku] && inventory.AvailableQuantity() > 0
}
