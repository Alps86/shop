package application

import (
	"context"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) StartCart(ctx context.Context, session Session, commandID value.CommandID) (CartView, error) {
	identity, err := service.identity(ctx, session)
	if err != nil {
		return CartView{}, err
	}

	return service.Starter.Start(ctx, identity.UserID, commandID, service.Clock.Now())
}

func (service Service) ApplyChange(ctx context.Context, command ChangeCommand) (CartView, error) {
	owned, err := service.cartForUser(ctx, command.UserID, command.CartID)
	if err != nil {
		return CartView{}, err
	}

	result, found, err := service.replayChange(ctx, command)
	if service.hasReplay(found, err) {
		return result, err
	}

	cart, err := service.change(owned, command)
	if err != nil {
		return NewCartView(service.visibleCart(owned)), err
	}

	return service.saveChange(ctx, cart, command)
}

func (service Service) Change(ctx context.Context, session Session, command ChangeCommand) (CartView, error) {
	cart, err := service.cart(ctx, session, command.CartID)
	if err != nil {
		return CartView{}, err
	}

	return service.routeChange(ctx, cart, command)
}

func (service Service) change(cart sale.Cart, command ChangeCommand) (sale.Cart, error) {
	if command.Operation == "add" {
		return cart.Add(command.SKU, command.Quantity, command.ExpectedRevision, service.Clock.Now())
	}

	if command.Operation == "setQuantity" {
		return cart.SetQuantity(command.SKU, command.Quantity, command.ExpectedRevision, service.Clock.Now())
	}

	if command.Operation == "remove" {
		return cart.Remove(command.SKU, command.ExpectedRevision, service.Clock.Now())
	}

	return service.configure(cart, command)
}

func (service Service) configure(cart sale.Cart, command ChangeCommand) (sale.Cart, error) {
	checkout := cart.Checkout.Value
	if !cart.Checkout.Present {
		checkout.ShippingMethod = sale.Standard
		checkout.PaymentMethod = sale.Invoice
	}

	checkout, err := command.checkout(checkout)
	if err != nil {
		return cart, err
	}

	return cart.ConfigureCheckout(checkout, command.ExpectedRevision, service.Clock.Now())
}

func (command ChangeCommand) checkout(checkout sale.Checkout) (sale.Checkout, error) {
	if command.Operation == "address" {
		checkout.ShippingAddress = command.Address
		return checkout, nil
	}

	if command.Operation == "shipping" {
		checkout.ShippingMethod = command.ShippingMethod
		return checkout, nil
	}

	if command.Operation == "payment" {
		checkout.PaymentMethod = command.PaymentMethod
		return checkout, nil
	}

	return checkout, value.NewError("InvalidInput", "Operation")
}

func (service Service) saveChange(ctx context.Context, cart sale.Cart, command ChangeCommand) (CartView, error) {
	evaluated, err := service.evaluateChange(ctx, cart)
	if err != nil {
		return CartView{}, err
	}

	write := CartWrite{Cart: evaluated, ExpectedRevision: command.ExpectedRevision, CommandID: command.CommandID, Change: value.Optional[ChangeCommand]{Value: command, Present: true}}

	return service.Writer.Save(ctx, write)
}

func (service Service) evaluateChange(ctx context.Context, cart sale.Cart) (sale.Cart, error) {
	if len(cart.Items) == 0 {
		return cart, nil
	}

	evaluation, err := service.Evaluator.Evaluate(ctx, cart)
	if err != nil {
		return cart, err
	}

	return cart.SetEvaluation(evaluation)
}

func (service Service) replayChange(ctx context.Context, command ChangeCommand) (CartView, bool, error) {
	if service.Commands == nil {
		return CartView{}, false, nil
	}

	return service.Commands.ReplayChange(ctx, command)
}

func (service Service) hasReplay(found bool, err error) bool {
	return found || err != nil
}

func (service Service) routeChange(ctx context.Context, cart sale.Cart, command ChangeCommand) (CartView, error) {
	command.UserID = cart.UserID
	if service.terminal(cart) {
		return service.ApplyChange(ctx, command)
	}

	return service.Processes.UpdateCart(ctx, command)
}
