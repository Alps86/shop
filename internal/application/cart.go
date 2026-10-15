package application

import (
	"context"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) cart(ctx context.Context, session Session, id value.CartID) (sale.Cart, error) {
	identity, err := service.identity(ctx, session)
	if err != nil {
		return sale.Cart{}, err
	}

	return service.cartForUser(ctx, identity.UserID, id)
}

func (service Service) cartForUser(ctx context.Context, user value.UserID, id value.CartID) (sale.Cart, error) {
	cart, err := service.Carts.Get(ctx, id)
	if err != nil {
		return sale.Cart{}, err
	}

	if cart.UserID != user {
		return sale.Cart{}, value.NewError("NotFound", "CartID")
	}

	return cart, nil
}

func (service Service) GetCart(ctx context.Context, session Session, id value.CartID) (CartView, error) {
	cart, err := service.cart(ctx, session, id)
	if err != nil {
		return CartView{}, err
	}

	if cart.State == sale.Ordered {
		return service.orderedCart(ctx, cart)
	}

	if cart.IsExpired(service.Clock.Now()) {
		cart.State = sale.Expired
		return NewCartView(cart), value.NewError("Expired", "CartID")
	}

	return NewCartView(cart), nil
}

func (service Service) ownedOrder(ctx context.Context, session Session, id value.OrderID) (sale.Order, error) {
	identity, err := service.identity(ctx, session)
	if err != nil {
		return sale.Order{}, err
	}

	order, err := service.Orders.Get(ctx, id)
	if err != nil {
		return sale.Order{}, err
	}

	if order.UserID != identity.UserID {
		return sale.Order{}, value.NewError("NotFound", "OrderID")
	}

	return order, nil
}

func (service Service) GetOrder(ctx context.Context, session Session, id value.OrderID) (OrderView, error) {
	order, err := service.ownedOrder(ctx, session, id)
	if err != nil {
		return OrderView{}, err
	}

	fulfillment, err := service.Fulfillment.Get(ctx, id)
	if err != nil {
		return OrderView{}, err
	}

	return OrderView{Purchase: order.Snapshot(), Fulfillment: NewFulfillmentView(fulfillment)}, nil
}

func (service Service) orderedCart(ctx context.Context, cart sale.Cart) (CartView, error) {
	id, err := service.Orders.ForCart(ctx, cart.ID)
	if err != nil {
		return CartView{}, err
	}

	view := NewCartView(cart)
	view.OrderID = value.Optional[value.OrderID]{Value: id, Present: true}

	return view, nil
}
