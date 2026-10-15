package application

import (
	"context"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) Evaluate(ctx context.Context, session Session, id value.CartID) (CartView, error) {
	cart, err := service.cart(ctx, session, id)
	if err != nil {
		return CartView{}, err
	}

	return service.Processes.EvaluateCart(ctx, EvaluateCommand{UserID: cart.UserID, CartID: id})
}

func (service Service) ApplyEvaluate(ctx context.Context, command EvaluateCommand) (CartView, error) {
	cart, err := service.cartForUser(ctx, command.UserID, command.CartID)
	if err != nil {
		return CartView{}, err
	}

	if err := cart.CanChange(cart.Revision, service.Clock.Now()); err != nil {
		return CartView{}, err
	}

	return service.evaluate(ctx, cart)
}

func (service Service) evaluate(ctx context.Context, cart sale.Cart) (CartView, error) {
	if len(cart.Items) == 0 {
		return NewCartView(cart), nil
	}

	evaluation, err := service.Evaluator.Evaluate(ctx, cart)
	if err != nil {
		return CartView{}, err
	}

	return service.saveEvaluation(ctx, cart, evaluation)
}

func (service Service) saveEvaluation(ctx context.Context, cart sale.Cart, evaluation sale.CartEvaluation) (CartView, error) {
	updated, err := cart.SetEvaluation(evaluation)
	if err != nil {
		return CartView{}, err
	}

	return service.Writer.Save(ctx, CartWrite{Cart: updated, ExpectedRevision: cart.Revision, CommandID: value.CommandID(evaluation.ID)})
}

func (service Service) Confirm(ctx context.Context, session Session, command ConfirmCommand) (CartView, error) {
	cart, err := service.cart(ctx, session, command.CartID)
	if err != nil {
		return CartView{}, err
	}

	command.UserID = cart.UserID

	return service.Processes.ConfirmEvaluation(ctx, command)
}

func (service Service) ApplyConfirm(ctx context.Context, command ConfirmCommand) (CartView, error) {
	owned, err := service.cartForUser(ctx, command.UserID, command.CartID)
	if err != nil {
		return CartView{}, err
	}

	cart, err := owned.Confirm(command.EvaluationID, command.ExpectedRevision, service.Clock.Now())
	if err != nil {
		return CartView{}, err
	}

	return service.Writer.Save(ctx, CartWrite{Cart: cart, ExpectedRevision: owned.Revision, CommandID: command.CommandID, Confirmation: value.Optional[ConfirmCommand]{Present: true, Value: command}})
}
