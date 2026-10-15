package temporal

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (adapter Adapter) replayDelivery(ctx context.Context, command application.DeliveryCommand) error {
	fulfillment, err := adapter.Service.Fulfillment.Get(ctx, command.OrderID)
	if err != nil {
		return err
	}

	if fulfillment.State() != "delivered" {
		return value.NewError("NotFound", "Process")
	}

	return adapter.replayReceipt(ctx, fulfillment, command)
}

func (adapter Adapter) replayReceipt(ctx context.Context, fulfillment sale.Fulfillment, command application.DeliveryCommand) error {
	_, err := fulfillment.Apply(command.Update)
	if err != nil {
		return err
	}

	_, err = (application.FulfillmentService{Store: adapter.Service.Fulfillment}).Report(ctx, command)

	return err
}
