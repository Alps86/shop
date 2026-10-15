package application

import "github.com/Alps86/shop/internal/domain/sale"

func (service Service) terminal(cart sale.Cart) bool {
	return cart.State == sale.Ordered || cart.IsExpired(service.Clock.Now())
}

func (service Service) visibleCart(cart sale.Cart) sale.Cart {
	if cart.State == sale.Ordered {
		return cart
	}

	if cart.IsExpired(service.Clock.Now()) {
		cart.State = sale.Expired
	}

	return cart
}
