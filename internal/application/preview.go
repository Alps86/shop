package application

import (
	"context"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) PreviewAccessory(ctx context.Context, session Session, id value.CartID, sku value.SKU) (sale.CartEvaluation, error) {
	cart, err := service.cart(ctx, session, id)
	if err != nil {
		return sale.CartEvaluation{}, err
	}

	cart.Items = append(sale.CartItems{}, cart.Items...)
	preview, err := cart.Add(sku, 1, cart.Revision, service.Clock.Now())
	if err != nil {
		return sale.CartEvaluation{}, err
	}

	return service.Evaluator.Evaluate(ctx, preview)
}
