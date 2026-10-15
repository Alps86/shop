package temporal

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	sdk "go.temporal.io/sdk/temporal"
	"time"
)

func (adapter Adapter) Change(ctx context.Context, command application.ChangeCommand) (cartResult, error) {
	view, err := adapter.Service.ApplyChange(ctx, command)

	return adapter.cartResult(view, err)
}

func (adapter Adapter) Evaluate(ctx context.Context, command application.EvaluateCommand) (cartResult, error) {
	view, err := adapter.Service.ApplyEvaluate(ctx, command)

	return adapter.cartResult(view, err)
}

func (adapter Adapter) Confirm(ctx context.Context, command application.ConfirmCommand) (cartResult, error) {
	view, err := adapter.Service.ApplyConfirm(ctx, command)

	return adapter.cartResult(view, err)
}

func (adapter Adapter) Purchase(ctx context.Context, command application.PlaceCommand) (purchaseResult, error) {
	result, err := adapter.Service.ApplyPlace(ctx, command)
	failure, technical := adapter.failure(err)

	return purchaseResult{Result: result, Failure: failure}, technical
}

func (adapter Adapter) Expire(ctx context.Context, id value.CartID, now time.Time) (sale.Cart, error) {
	cart, err := adapter.Expiration.Expire(ctx, id, now)
	if value.Code(err) == "Expired" {
		return cart, nil
	}

	return cart, adapter.activityError(err)
}

func (adapter Adapter) Delivery(ctx context.Context, command application.DeliveryCommand) (application.FulfillmentView, error) {
	view, err := (application.FulfillmentService{Store: adapter.Service.Fulfillment}).Report(ctx, command)

	return view, adapter.activityError(err)
}

func (adapter Adapter) cartResult(view application.CartView, err error) (cartResult, error) {
	failure, technical := adapter.failure(err)

	return cartResult{View: view, Failure: failure}, technical
}

func (adapter Adapter) failure(err error) (value.Optional[value.Error], error) {
	var failure value.Error
	if errors.As(err, &failure) {
		if !adapter.technical(failure) {
			return value.Optional[value.Error]{Value: failure, Present: true}, nil
		}
	}

	return value.Optional[value.Error]{}, err
}

func (adapter Adapter) activityError(err error) error {
	failure, technical := adapter.failure(err)
	if failure.Present {
		return sdk.NewNonRetryableApplicationError(failure.Value.Error(), failure.Value.Code, nil)
	}

	return technical
}

func (adapter Adapter) technical(err value.Error) bool {
	return err.Code == "DependencyUnavailable" || err.Code == "DeadlineExceeded"
}
