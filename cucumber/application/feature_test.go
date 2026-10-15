package application_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"testing"
	"time"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @app", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "application", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Applicationverträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		current.reset()
		return ctx, nil
	})
	current.registerCheckout(ctx)
	current.registerPreview(ctx)
	current.registerInterests(ctx)
	current.registerReplay(ctx)
	current.registerActivities(ctx)
	current.registerAtomicity(ctx)
	current.registerDelivery(ctx)
	current.registerGuest(ctx)
	current.registerCoordination(ctx)
	current.registerRecommendations(ctx)
	current.registerSearch(ctx)
}

func (current *world) reset() {
	*current = world{now: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), sessions: map[application.Session]application.SessionIdentity{}, profiles: map[value.UserID]application.ProfileCode{}, carts: map[value.CartID]sale.Cart{}, orders: map[value.OrderID]sale.Order{}, inventories: map[value.SKU]sale.Inventory{}}
	current.commandWrites = map[value.CommandID]application.CartWrite{}
	current.purchaseRecords = map[value.CommandID]purchaseRecord{}
	current.inventoryProducts = map[value.SKU]value.ProductID{}
	current.service = application.Service{Guests: current, Carts: current, Orders: orderStore{current}, Clock: current, Purchase: current, Payment: current, Evaluator: evaluator{current}, Writer: cartWriter{current}, Fulfillment: deliveryStore{current}, Starter: cartStarter{current}}
}

func (current *world) Now() time.Time {
	return current.now
}

func (current *world) code(table *godog.Table) error {
	expected := read(table)[0]["Code"]
	if value.Code(current.err) != expected {
		return fmt.Errorf("Code: %s statt %s (%w)", value.Code(current.err), expected, current.err)
	}

	return nil
}

func read(table *godog.Table) rows {
	result := rows{}
	for _, row := range table.Rows[1:] {
		fields := map[string]string{}
		for index, cell := range row.Cells {
			fields[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, fields)
	}

	return result
}

func (expected rows) matches(actual map[string]string) error {
	for field, wanted := range expected[0] {
		if actual[field] != wanted {
			return fmt.Errorf("%s: %q statt %q", field, actual[field], wanted)
		}
	}

	return nil
}
