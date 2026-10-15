package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/client"
	"time"
)

func (current *world) registerBoundary(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine echte Temporalgrenze mit begrenztem Kontext aufgerufen wird$`, current.boundary)
	ctx.Step(`^die technische Grenze einen ehrlichen Fehler meldet$`, current.boundaryObserved)
}

func (current *world) boundary(table *godog.Table) error {
	row := current.rows(table)[0]
	duration, err := time.ParseDuration(row["Timeout"])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	current.boundaryOperation(ctx, row)

	return nil
}

func (current *world) boundaryOperation(ctx context.Context, row map[string]string) {
	operations := map[string]func(){
		"closed": func() {
			current.err = current.closedConnection(ctx)
		},
		"ensure": func() {
			current.err = current.adapter.EnsureCart(ctx, value.CartID(current.id), application.ProcessBinding{DefinitionVersion: row["DefinitionVersion"], RuleSetVersion: row["RuleSetVersion"]})
		},
		"update": func() {
			current.view, current.err = current.adapter.UpdateCart(ctx, application.ChangeCommand{CartID: value.CartID(current.id), UserID: value.UserID(row["UserID"])})
		},
		"report": func() {
			current.err = current.adapter.ReportDelivery(ctx, application.DeliveryCommand{CommandID: value.CommandID(row["OuterCommandID"]), Update: sale.FulfillmentUpdate{CommandID: value.CommandID(row["InnerCommandID"])}})
		},
	}

	operations[row["Operation"]]()
}

func (current *world) boundaryObserved(table *godog.Table) error {
	wanted := current.rows(table)[0]["Code"]
	if value.Code(current.err) != wanted {
		return fmt.Errorf("Code %s statt %s: %w", value.Code(current.err), wanted, current.err)
	}

	return nil
}

func (current *world) closedConnection(ctx context.Context) error {
	connection, err := client.NewLazyClient(client.Options{HostPort: "127.0.0.1:1"})
	if err != nil {
		return err
	}

	connection.Close()
	adapter := current.adapter
	adapter.Client = connection

	return adapter.EnsureCart(ctx, value.CartID(current.id), application.ProcessBinding{DefinitionVersion: "cart-v1", RuleSetVersion: "demo-v1"})
}
