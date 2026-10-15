package temporal

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func New(connection client.Client, service application.Service, expiration ExpirationPort) Adapter {
	return Adapter{Client: connection, Service: service, Expiration: expiration, Queue: "shop"}
}

func (adapter Adapter) Worker() worker.Worker {
	runner := worker.New(adapter.Client, adapter.Queue, worker.Options{})
	runner.RegisterWorkflow(adapter.CartWorkflow)
	runner.RegisterWorkflow(adapter.OrderWorkflow)
	runner.RegisterActivity(adapter.Change)
	runner.RegisterActivity(adapter.Evaluate)
	runner.RegisterActivity(adapter.Confirm)
	runner.RegisterActivity(adapter.Purchase)
	runner.RegisterActivity(adapter.Expire)
	runner.RegisterActivity(adapter.Delivery)

	return runner
}

func (adapter Adapter) EnsureCart(ctx context.Context, id value.CartID, binding application.ProcessBinding) error {
	return adapter.ensure(ctx, "cart/"+string(id), "CartWorkflow", Start{ID: string(id), Binding: binding})
}

func (adapter Adapter) EnsureOrder(ctx context.Context, id value.OrderID, binding application.ProcessBinding) error {
	return adapter.ensure(ctx, "order/"+string(id), "OrderWorkflow", Start{ID: string(id), Binding: binding})
}

func (adapter Adapter) ensure(ctx context.Context, id, name string, start Start) error {
	options := client.StartWorkflowOptions{ID: id, TaskQueue: adapter.Queue, WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE, WorkflowExecutionErrorWhenAlreadyStarted: true}
	_, err := adapter.Client.ExecuteWorkflow(ctx, options, name, start)
	var started *serviceerror.WorkflowExecutionAlreadyStarted
	if errors.As(err, &started) {
		return nil
	}

	return adapter.connectionError(err)
}

func (adapter Adapter) ReportDelivery(ctx context.Context, command application.DeliveryCommand) error {
	if command.CommandID != command.Update.CommandID {
		return value.NewError("InvalidInput", "CommandID")
	}

	err := adapter.Client.SignalWorkflow(ctx, "order/"+string(command.OrderID), "", "ReportDelivery", command)
	var missing *serviceerror.NotFound
	if errors.As(err, &missing) {
		return adapter.replayDelivery(ctx, command)
	}

	return adapter.connectionError(err)
}

func (adapter Adapter) connectionError(err error) error {
	if err == nil {
		return nil
	}

	if adapter.deadline(err) {
		return value.NewError("DeadlineExceeded", "Process")
	}

	return value.NewError("DependencyUnavailable", "Process")
}

func (adapter Adapter) deadline(err error) bool {
	var deadline *serviceerror.DeadlineExceeded

	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &deadline)
}
