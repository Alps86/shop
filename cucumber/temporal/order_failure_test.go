package temporal_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerOrderFailure(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Lieferabhängigkeit im kontrollierten SDK-Lauf ausfällt$`, current.orderFailure)
}

func (store failedFulfillment) Get(_ context.Context, _ value.OrderID) (sale.Fulfillment, error) {
	return sale.Fulfillment{}, store.Failure
}

func (store failedFulfillment) Save(_ context.Context, _ sale.Fulfillment, _ value.CommandID) error {
	return store.Failure
}

func (current *world) orderFailure(table *godog.Table) error {
	row := current.rows(table)[0]
	current.adapter.Service.Fulfillment = failedFulfillment{Failure: value.NewError(row["Code"], "Fulfillment")}
	current.testEnvironment.RegisterActivity(current.adapter.Delivery)
	current.testEnvironment.RegisterDelayedCallback(func() {
		current.testEnvironment.SignalWorkflow("ReportDelivery", application.DeliveryCommand{OrderID: "o-a", CommandID: "ship-a"})
	}, time.Second)
	current.testEnvironment.RegisterDelayedCallback(current.queryAndCancel, 20*time.Second)
	current.testEnvironment.ExecuteWorkflow(current.adapter.OrderWorkflow, temporal.Start{ID: "o-a", Binding: application.ProcessBinding{DefinitionVersion: "order-v1", RuleSetVersion: "demo-v1"}})
	current.err = current.testEnvironment.GetWorkflowError()

	return nil
}

func (current *world) queryAndCancel() {
	encoding, err := current.testEnvironment.QueryWorkflow("ProcessStatus")
	current.queryErr = err
	if err == nil {
		current.queryErr = encoding.Get(&current.processStatus)
	}

	current.testEnvironment.CancelWorkflow()
}
