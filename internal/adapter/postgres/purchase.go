package postgres

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository PurchaseRepository) Commit(ctx context.Context, command application.PurchaseCommand) (application.PurchaseResult, error) {
	attempt := purchaseAttempt{orderID: value.OrderID("o-" + rand.Text()), repository: repository, command: command, inventory: inventoryState{}, products: map[value.SKU]catalog.Product{}}
	err := repository.store.transact(ctx, func(tx pgx.Tx) error {
		attempt.tx = tx
		return attempt.transaction(ctx)
	})

	if err != nil {
		return attempt.result, err
	}

	return attempt.result, attempt.rejection
}

func (attempt *purchaseAttempt) transaction(ctx context.Context) error {
	err := attempt.execute(ctx)
	if value.Code(err) == "PriceConfirmationRequired" {
		attempt.rejection = err
		return nil
	}

	return err
}

func (attempt *purchaseAttempt) execute(ctx context.Context) error {
	cart, err := attempt.repository.store.Carts().load(ctx, attempt.tx, attempt.command.CartID, true)
	if err != nil {
		return err
	}

	attempt.cart = cart
	if cart.UserID != attempt.command.UserID {
		return value.NewError("NotFound", "CartID")
	}

	repeated, err := attempt.receipt().previous(ctx, attempt.tx)
	if repeated {
		return err
	}

	if err != nil {
		return err
	}

	return attempt.checkAndCommit(ctx)
}

func (attempt *purchaseAttempt) receipt() commandRecord {
	fingerprint := placeFingerprint{UserID: attempt.command.UserID, CartID: attempt.command.CartID, EvaluationID: attempt.command.EvaluationID}
	return commandRecord{scope: "purchase/" + string(attempt.command.CartID), id: attempt.command.CommandID, payload: fingerprint, result: &attempt.result}
}

func (attempt *purchaseAttempt) checkAndCommit(ctx context.Context) error {
	if attempt.cart.State == sale.Ordered {
		id, err := attempt.repository.store.Orders().ForCart(ctx, attempt.cart.ID)
		attempt.result.OrderID = id
		if err != nil {
			return err
		}

		return value.NewError("AlreadyOrdered", "Cart")
	}

	attempt.command.Now = time.Now().UTC()
	if err := attempt.validateCart(); err != nil {
		return err
	}

	return attempt.currentPurchase(ctx)
}

func (attempt *purchaseAttempt) validateCart() error {
	if attempt.authorizationStale() {
		return value.NewError("Conflict", "PaymentAuthorization")
	}

	if err := attempt.cart.CanChange(attempt.cart.Revision, attempt.command.Now); err != nil {
		return err
	}

	confirmed := attempt.cart.Checkout.Value.ConfirmedEvaluation
	if !confirmed.Present {
		return value.NewError("PriceConfirmationRequired", "Evaluation")
	}

	if !attempt.confirmationMatches(confirmed.Value) {
		return value.NewError("PriceConfirmationRequired", "Evaluation")
	}

	return nil
}

func (attempt *purchaseAttempt) confirmationMatches(confirmed sale.CartEvaluation) bool {
	return confirmed.ID == attempt.command.EvaluationID && confirmed.CartRevision == attempt.cart.Revision && confirmed.RuleSetVersion == attempt.cart.RuleSetVersion
}

func (attempt *purchaseAttempt) currentPurchase(ctx context.Context) error {
	if err := attempt.lockInventory(ctx); err != nil {
		return err
	}

	evaluation, err := attempt.repository.store.Evaluator().evaluate(ctx, attempt.tx, attempt.cart)
	if err != nil {
		return err
	}

	attempt.evaluation = evaluation
	if !evaluation.MatchesPrices(attempt.cart.Checkout.Value.ConfirmedEvaluation.Value) {
		return attempt.requireConfirmation(ctx)
	}

	if err := attempt.authorization(); err != nil {
		return err
	}

	return attempt.persist(ctx)
}

func (attempt *purchaseAttempt) requireConfirmation(ctx context.Context) error {
	attempt.cart.Evaluation = value.Optional[sale.CartEvaluation]{Present: true, Value: attempt.evaluation}
	attempt.cart.Checkout.Value.ConfirmedEvaluation = value.Optional[sale.CartEvaluation]{}
	if err := attempt.repository.store.Carts().update(ctx, attempt.tx, attempt.cart); err != nil {
		return err
	}

	attempt.result = application.NewCartPurchaseResult(attempt.cart)

	return value.NewError("PriceConfirmationRequired", "Evaluation")
}

func (attempt *purchaseAttempt) authorization() error {
	if attempt.cart.Checkout.Value.PaymentMethod == sale.Invoice {
		return nil
	}

	if !attempt.authorizationMatches() {
		return value.NewError("Conflict", "PaymentAuthorization")
	}

	if !attempt.command.Authorization.Value.Approved {
		return value.NewError("PaymentDeclined", "PaymentAuthorization")
	}

	return nil
}

func (attempt *purchaseAttempt) authorizationMatches() bool {
	authorization := attempt.command.Authorization.Value
	if !attempt.command.Authorization.Present {
		return false
	}

	if authorization.CartID != attempt.cart.ID {
		return false
	}

	return authorization.CartRevision == attempt.cart.Revision && authorization.Total == attempt.evaluation.Total && authorization.CommandID == attempt.command.CommandID
}

func (attempt *purchaseAttempt) persist(ctx context.Context) error {
	order, ordered, err := sale.NewPurchase(attempt.orderID, attempt.cart, attempt.orderItems(), attempt.command.Now)
	if err != nil {
		return err
	}

	attempt.ordered = ordered
	if err := attempt.reserve(ctx); err != nil {
		return err
	}

	if err := attempt.insertOrder(ctx, order); err != nil {
		return err
	}

	return attempt.finish(ctx, order)
}

func (attempt *purchaseAttempt) insertOrder(ctx context.Context, order sale.Order) error {
	fulfillment := sale.NewFulfillment(order)
	_, err := attempt.tx.Exec(ctx, "INSERT INTO orders(id,cart_id,user_id,payload,fulfillment) VALUES($1,$2,$3,$4,$5)", order.ID, order.CartID, order.UserID, order, fulfillment)

	return databaseError(err)
}

func (attempt *purchaseAttempt) finish(ctx context.Context, order sale.Order) error {
	if err := attempt.repository.store.Carts().update(ctx, attempt.tx, attempt.ordered); err != nil {
		return err
	}

	envelope := event.ProcessEnvelope{EventID: "e-" + rand.Text(), Kind: "EnsureOrder", OrderID: order.ID, Binding: application.ProcessBinding{DefinitionVersion: order.DefinitionVersion, RuleSetVersion: order.RuleSetVersion}}
	if err := attempt.repository.store.insertProcess(ctx, attempt.tx, envelope); err != nil {
		return err
	}

	attempt.result.OrderID = order.ID

	return attempt.receipt().save(ctx, attempt.tx)
}

func (attempt *purchaseAttempt) orderItems() []sale.OrderItem {
	result := []sale.OrderItem{}
	for _, item := range attempt.cart.Checkout.Value.ConfirmedEvaluation.Value.Items {
		result = append(result, attempt.orderItem(item))
	}

	return result
}

func (attempt *purchaseAttempt) orderItem(item sale.EvaluatedCartItem) sale.OrderItem {
	product := attempt.products[item.SKU]
	attributes := catalog.Attributes{}
	for _, variant := range product.Variants {
		if variant.SKU == item.SKU {
			attributes = variant.Attributes.Clone()
		}
	}

	name, _ := product.Name.Resolve(attempt.repository.store.config.Context.Locale)

	return sale.OrderItem{SKU: item.SKU, ProductName: name, VariantAttributes: attributes, Quantity: item.Quantity, UnitPrice: item.UnitPrice, Discounts: item.Discounts}
}

func (attempt *purchaseAttempt) authorizationStale() bool {
	authorization := attempt.command.Authorization
	return authorization.Present && authorization.Value.CartRevision != attempt.cart.Revision
}
