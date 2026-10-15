package application

import (
	"context"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) Place(ctx context.Context, session Session, command PlaceCommand) (PurchaseResult, error) {
	cart, err := service.cart(ctx, session, command.CartID)
	if err != nil {
		return PurchaseResult{}, err
	}

	return service.routePlace(ctx, cart, command)
}

func (service Service) ApplyPlace(ctx context.Context, command PlaceCommand) (PurchaseResult, error) {
	cart, err := service.cartForUser(ctx, command.UserID, command.CartID)
	if err != nil {
		return PurchaseResult{}, err
	}

	if cart.State == sale.Ordered {
		return service.commit(ctx, cart, command, value.Optional[PaymentAuthorization]{})
	}

	return service.finish(ctx, cart, command)
}

func (service Service) finish(ctx context.Context, cart sale.Cart, command PlaceCommand) (PurchaseResult, error) {
	if err := service.ready(cart, command.EvaluationID); err != nil {
		return NewCartPurchaseResult(service.visibleCart(cart)), err
	}

	authorization, err := service.authorize(ctx, cart, command.CommandID)
	if err != nil {
		return PurchaseResult{}, err
	}

	return service.commit(ctx, cart, command, authorization)
}

func (service Service) ready(cart sale.Cart, evaluationID value.EvaluationID) error {
	if cart.IsExpired(service.Clock.Now()) {
		return value.NewError("Expired", "CartID")
	}

	if !cart.Checkout.Present {
		return value.NewError("PriceConfirmationRequired", "Checkout")
	}

	confirmed := cart.Checkout.Value.ConfirmedEvaluation
	if !confirmed.Present {
		return value.NewError("PriceConfirmationRequired", "EvaluationID")
	}

	if !service.confirmationMatches(cart, confirmed.Value, evaluationID) {
		return value.NewError("PriceConfirmationRequired", "EvaluationID")
	}

	return nil
}

func (service Service) confirmationMatches(cart sale.Cart, confirmed sale.CartEvaluation, id value.EvaluationID) bool {
	return confirmed.ID == id && confirmed.CartRevision == cart.Revision && confirmed.RuleSetVersion == cart.RuleSetVersion
}

func (service Service) authorize(ctx context.Context, cart sale.Cart, commandID value.CommandID) (value.Optional[PaymentAuthorization], error) {
	if cart.Checkout.Value.PaymentMethod == "invoice" {
		return value.Optional[PaymentAuthorization]{}, nil
	}

	authorization, err := service.Payment.Authorize(ctx, cart, commandID)
	if err != nil {
		return value.Optional[PaymentAuthorization]{}, err
	}

	if !authorization.matches(cart, commandID) {
		return value.Optional[PaymentAuthorization]{}, value.NewError("Conflict", "PaymentAuthorization")
	}

	if !authorization.Approved {
		return value.Optional[PaymentAuthorization]{}, value.NewError("PaymentDeclined", "Payment")
	}

	return value.Optional[PaymentAuthorization]{Value: authorization, Present: true}, nil
}

func (service Service) commit(ctx context.Context, cart sale.Cart, command PlaceCommand, authorization value.Optional[PaymentAuthorization]) (PurchaseResult, error) {
	purchase := PurchaseCommand{UserID: cart.UserID, CartID: cart.ID, CommandID: command.CommandID, EvaluationID: command.EvaluationID, Authorization: authorization, Now: service.Clock.Now()}
	return service.Purchase.Commit(ctx, purchase)
}

func (authorization PaymentAuthorization) matches(cart sale.Cart, id value.CommandID) bool {
	return authorization.CartID == cart.ID && authorization.CartRevision == cart.Revision && authorization.CommandID == id && authorization.Total == cart.Checkout.Value.ConfirmedEvaluation.Value.Total
}

func (service Service) routePlace(ctx context.Context, cart sale.Cart, command PlaceCommand) (PurchaseResult, error) {
	command.UserID = cart.UserID
	if service.terminal(cart) {
		return service.ApplyPlace(ctx, command)
	}

	return service.Processes.PlaceOrder(ctx, command)
}
