package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) processSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die getrennte Prozessoutbox ausgeliefert wird$`, current.dispatch)
	ctx.Step(`^lautet die dauerhafte Prozessbeobachtung$`, current.processObserved)
}

func (current *world) dispatch(input *godog.Table) {
	current.processFailure = table(input)["FailureKind"]
	if current.processCarts == nil {
		current.processCarts = map[value.CartID]bool{}
		current.processOrders = map[value.OrderID]bool{}
	}

	current.err = (postgres.ProcessDispatcher{Store: current.store, Processes: processObserver{current}}).Dispatch(context.Background())
}

func (current *world) processObserved(input *godog.Table) error {
	var pending, checkpoints int
	if err := current.admin.QueryRow(context.Background(), "SELECT (SELECT count(*) FROM "+current.schema+".process_outbox WHERE NOT acknowledged), (SELECT count(*) FROM "+current.schema+".cdc_checkpoints)").Scan(&pending, &checkpoints); err != nil {
		return err
	}

	return check(table(input), map[string]string{"Code": value.Code(current.err), "CartCount": strconv.Itoa(len(current.processCarts)), "OrderCount": strconv.Itoa(len(current.processOrders)), "Pending": strconv.Itoa(pending), "SearchCheckpoints": strconv.Itoa(checkpoints)})
}

func (observer processObserver) EnsureCart(_ context.Context, id value.CartID, _ application.ProcessBinding) error {
	if observer.current.processFailure == "EnsureCart" {
		return value.NewError("DependencyUnavailable", "Temporal")
	}

	observer.current.processCarts[id] = true

	return nil
}

func (observer processObserver) EnsureOrder(_ context.Context, id value.OrderID, _ application.ProcessBinding) error {
	if observer.current.processFailure == "EnsureOrder" {
		return value.NewError("DependencyUnavailable", "Temporal")
	}

	observer.current.processOrders[id] = true

	return nil
}

func (observer processObserver) UpdateCart(context.Context, application.ChangeCommand) (application.CartView, error) {
	return application.CartView{}, value.NewError("InvalidTransition", "UnusedPort")
}

func (observer processObserver) EvaluateCart(context.Context, application.EvaluateCommand) (application.CartView, error) {
	return application.CartView{}, value.NewError("InvalidTransition", "UnusedPort")
}

func (observer processObserver) ConfirmEvaluation(context.Context, application.ConfirmCommand) (application.CartView, error) {
	return application.CartView{}, value.NewError("InvalidTransition", "UnusedPort")
}

func (observer processObserver) PlaceOrder(context.Context, application.PlaceCommand) (application.PurchaseResult, error) {
	return application.PurchaseResult{}, value.NewError("InvalidTransition", "UnusedPort")
}

func (observer processObserver) ReportDelivery(context.Context, application.DeliveryCommand) error {
	return value.NewError("InvalidTransition", "UnusedPort")
}
