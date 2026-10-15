package temporal

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
)

func (adapter Adapter) CartWorkflow(ctx workflow.Context, start Start) error {
	process := cartProcess{Start: start, Status: Status{State: "running", Binding: start.Binding}}
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	if err := process.register(ctx); err != nil {
		return err
	}

	return process.run(ctx)
}

func NewActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{StartToCloseTimeout: 10 * time.Second, ScheduleToCloseTimeout: 2 * time.Minute, RetryPolicy: &temporal.RetryPolicy{InitialInterval: time.Second, MaximumInterval: 10 * time.Second, BackoffCoefficient: 2, MaximumAttempts: 5}}
}

func (process *cartProcess) register(ctx workflow.Context) error {
	handlers := []updateHandler{
		{"ChangeCart", process.change}, {"EvaluateCart", process.evaluate}, {"ConfirmEvaluation", process.confirm}, {"PlaceOrder", process.place},
	}

	for _, handler := range handlers {
		if err := workflow.SetUpdateHandler(ctx, handler.Name, handler.Handler); err != nil {
			return err
		}
	}

	return workflow.SetQueryHandler(ctx, "ProcessStatus", process.status)
}

func (process *cartProcess) status() (Status, error) {
	return process.Status, nil
}

func (process *cartProcess) run(ctx workflow.Context) error {
	for !process.Terminal {
		if err := process.tick(ctx); err != nil {
			return err
		}
	}

	return process.finish(ctx)
}

func (process *cartProcess) tick(ctx workflow.Context) error {
	var cart sale.Cart
	err := workflow.ExecuteActivity(ctx, "Expire", value.CartID(process.Start.ID), workflow.Now(ctx)).Get(ctx, &cart)
	process.attention(err)
	if err != nil {
		return workflow.Await(ctx, process.recovered)
	}

	return process.pause(ctx, cart)
}

func (process *cartProcess) pause(ctx workflow.Context, cart sale.Cart) error {
	if cart.State == sale.Expired {
		process.Terminal = true
		return nil
	}

	if cart.State == sale.Ordered {
		process.Terminal = true
		return nil
	}

	duration := cart.LastActivityAt.Add(168 * time.Hour).Sub(workflow.Now(ctx))
	_, err := workflow.AwaitWithTimeout(ctx, duration, func() bool {
		return process.Terminal
	})

	return err
}

func (process *cartProcess) attention(err error) {
	process.Status.State = "running"
	if err != nil {
		process.Status.State = "needsAttention"
	}
}

func (process *cartProcess) change(ctx workflow.Context, command application.ChangeCommand) (cartResult, error) {
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	var result cartResult
	err := workflow.ExecuteActivity(ctx, "Change", command).Get(ctx, &result)
	process.finished(err, result.Failure.Present)

	return result, err
}

func (process *cartProcess) evaluate(ctx workflow.Context, command application.EvaluateCommand) (cartResult, error) {
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	var result cartResult
	err := workflow.ExecuteActivity(ctx, "Evaluate", command).Get(ctx, &result)
	process.finished(err, result.Failure.Present)

	return result, err
}

func (process *cartProcess) confirm(ctx workflow.Context, command application.ConfirmCommand) (cartResult, error) {
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	var result cartResult
	err := workflow.ExecuteActivity(ctx, "Confirm", command).Get(ctx, &result)
	process.finished(err, result.Failure.Present)

	return result, err
}

func (process *cartProcess) place(ctx workflow.Context, command application.PlaceCommand) (purchaseResult, error) {
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	var result purchaseResult
	err := workflow.ExecuteActivity(ctx, "Purchase", command).Get(ctx, &result)
	process.finished(err, result.Failure.Present)
	if result.Result.OrderID != "" {
		process.Terminal = true
	}

	return result, err
}

func (adapter Adapter) OrderWorkflow(ctx workflow.Context, start Start) error {
	process := orderProcess{Start: start, Status: Status{State: "running", Binding: start.Binding}}
	ctx = workflow.WithActivityOptions(ctx, NewActivityOptions())
	if err := workflow.SetQueryHandler(ctx, "ProcessStatus", process.status); err != nil {
		return err
	}

	return process.run(ctx)
}

func (process *orderProcess) status() (Status, error) {
	return process.Status, nil
}

func (process *orderProcess) run(ctx workflow.Context) error {
	channel := workflow.GetSignalChannel(ctx, "ReportDelivery")
	for !process.Terminal {
		var command application.DeliveryCommand
		channel.Receive(ctx, &command)
		var result application.FulfillmentView
		err := workflow.ExecuteActivity(ctx, "Delivery", command).Get(ctx, &result)
		process.Status.State = "running"
		if err != nil {
			process.Status.State = "needsAttention"
		}

		process.Terminal = result.State == "delivered"
	}

	return nil
}

func (process *cartProcess) finish(ctx workflow.Context) error {
	return workflow.Await(ctx, func() bool {
		return workflow.AllHandlersFinished(ctx)
	})
}

func (process *cartProcess) recovered() bool {
	return process.Status.State == "running" || process.Terminal
}

func (process *cartProcess) finished(err error, rejected bool) {
	if rejected {
		return
	}

	process.attention(err)
}
