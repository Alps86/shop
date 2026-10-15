package temporal

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"time"
)

func (adapter Adapter) UpdateCart(ctx context.Context, command application.ChangeCommand) (application.CartView, error) {
	return adapter.cartUpdate(ctx, workflowUpdate{ID: string(command.CartID), UserID: command.UserID, Name: "ChangeCart", Command: command})
}

func (adapter Adapter) EvaluateCart(ctx context.Context, command application.EvaluateCommand) (application.CartView, error) {
	return adapter.cartUpdate(ctx, workflowUpdate{ID: string(command.CartID), UserID: command.UserID, Name: "EvaluateCart", Command: command})
}

func (adapter Adapter) ConfirmEvaluation(ctx context.Context, command application.ConfirmCommand) (application.CartView, error) {
	return adapter.cartUpdate(ctx, workflowUpdate{ID: string(command.CartID), UserID: command.UserID, Name: "ConfirmEvaluation", Command: command})
}

func (adapter Adapter) PlaceOrder(ctx context.Context, command application.PlaceCommand) (application.PurchaseResult, error) {
	var result purchaseResult
	pending, err := adapter.update(ctx, workflowUpdate{ID: string(command.CartID), UserID: command.UserID, Name: "PlaceOrder", Command: command}, &result)
	if pending {
		return application.PurchaseResult{Pending: true, ProcessReference: "cart/" + string(command.CartID)}, nil
	}

	if result.Failure.Present {
		return result.Result, result.Failure.Value
	}

	return result.Result, err
}

func (adapter Adapter) cartUpdate(ctx context.Context, request workflowUpdate) (application.CartView, error) {
	var result cartResult
	pending, err := adapter.update(ctx, request, &result)
	if pending {
		return result.View, value.NewError("Pending", "cart/"+request.ID)
	}

	if result.Failure.Present {
		return result.View, result.Failure.Value
	}

	return result.View, err
}

func (adapter Adapter) update(ctx context.Context, request workflowUpdate, result any) (bool, error) {
	bounded, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	handle, err := adapter.accept(bounded, request)
	if err != nil {
		return false, err
	}

	return adapter.wait(bounded, handle, result)
}

func (adapter Adapter) accept(ctx context.Context, request workflowUpdate) (client.WorkflowUpdateHandle, error) {
	options := client.UpdateWorkflowOptions{WorkflowID: "cart/" + request.ID, UpdateName: request.Name, Args: []any{request.Command}, WaitForStage: client.WorkflowUpdateStageAccepted}
	handle, err := adapter.Client.UpdateWorkflow(ctx, options)
	var missing *serviceerror.NotFound
	if !errors.As(err, &missing) {
		return handle, adapter.connectionError(err)
	}

	if err := adapter.recoverCart(ctx, request); err != nil {
		return nil, err
	}

	handle, err = adapter.Client.UpdateWorkflow(ctx, options)

	return handle, adapter.connectionError(err)
}

func (adapter Adapter) recoverCart(ctx context.Context, request workflowUpdate) error {
	if adapter.Service.Carts == nil {
		return value.NewError("DependencyUnavailable", "CartStore")
	}

	cart, err := adapter.Service.Carts.Get(ctx, value.CartID(request.ID))
	if err != nil {
		return err
	}

	if cart.UserID != request.UserID {
		return value.NewError("NotFound", "CartID")
	}

	return adapter.EnsureCart(ctx, cart.ID, application.ProcessBinding{DefinitionVersion: cart.DefinitionVersion, RuleSetVersion: cart.RuleSetVersion})
}

func (adapter Adapter) wait(ctx context.Context, handle client.WorkflowUpdateHandle, result any) (bool, error) {
	err := handle.Get(ctx, result)
	pending := ctx.Err() != nil || adapter.deadline(err)
	if pending {
		err = nil
	}

	return pending, adapter.connectionError(err)
}
