package application

import (
	"context"
	"github.com/Alps86/shop/internal/domain/sale"
	"sort"
)

func (service FulfillmentService) Report(ctx context.Context, command DeliveryCommand) (FulfillmentView, error) {
	fulfillment, err := service.Store.Get(ctx, command.OrderID)
	if err != nil {
		return FulfillmentView{}, err
	}

	updated, err := fulfillment.Apply(command.Update)
	if err != nil {
		return FulfillmentView{}, err
	}

	if err := service.Store.Save(ctx, updated, command.CommandID); err != nil {
		return FulfillmentView{}, err
	}

	return NewFulfillmentView(updated), nil
}

func NewFulfillmentView(fulfillment sale.Fulfillment) FulfillmentView {
	lines := make([]DeliveryLine, 0, len(fulfillment.Purchased))
	for sku, quantity := range fulfillment.Purchased {
		lines = append(lines, DeliveryLine{SKU: sku, Purchased: quantity, Shipped: fulfillment.Shipped(sku), Delivered: fulfillment.Delivered(sku)})
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].SKU < lines[j].SKU
	})

	return FulfillmentView{State: fulfillment.State(), Lines: lines}
}
