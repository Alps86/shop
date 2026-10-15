package temporal_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerVersion(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein neuer Regelstand nur für neue Carts veröffentlicht wird$`, current.publishRules)
	ctx.Step(`^bestehende und neue Prozesse ihre jeweiligen Versionen behalten$`, current.versionsObserved)
}

func (current *world) publishRules(table *godog.Table) error {
	row := current.rows(table)[0]
	payload, err := current.readFixture(row["Fixture"])
	if err != nil {
		return err
	}

	var manifest postgres.ImportManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return err
	}

	manifest.DatasetVersion = row["PublishedRules"]
	if _, err := current.databaseStore.Import(context.Background(), manifest); err != nil {
		return err
	}

	return current.openVersionedCart(row)
}

func (current *world) openVersionedCart(row map[string]string) error {
	view, err := current.service.StartCart(context.Background(), current.guest.Session, value.CommandID(row["CommandID"]))
	if err != nil {
		return err
	}

	current.newCartID = view.Cart.ID
	binding := application.ProcessBinding{DefinitionVersion: view.Cart.DefinitionVersion, RuleSetVersion: view.Cart.RuleSetVersion}

	if err := current.adapter.EnsureCart(context.Background(), view.Cart.ID, binding); err != nil {
		return err
	}

	return current.initializeVersionedCart()
}

func (current *world) initializeVersionedCart() error {
	_, err := current.adapter.EvaluateCart(context.Background(), application.EvaluateCommand{UserID: current.guest.UserID, CartID: current.newCartID})

	return err
}

func (current *world) versionsObserved(table *godog.Table) error {
	row := current.rows(table)[0]
	expected := map[string]string{current.id: row["ExistingRules"], string(current.newCartID): row["NewRules"]}
	for id, rules := range expected {
		if err := current.boundVersion(id, rules); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) boundVersion(id, rules string) error {
	_, err := current.adapter.EvaluateCart(context.Background(), application.EvaluateCommand{UserID: current.guest.UserID, CartID: value.CartID(id)})
	if err != nil {
		return err
	}

	return current.queryBoundVersion(id, rules)
}

func (current *world) queryBoundVersion(id, rules string) error {
	cart, err := current.databaseStore.Carts().Get(context.Background(), value.CartID(id))
	if err != nil {
		return err
	}

	encoded, err := current.connection.QueryWorkflow(context.Background(), "cart/"+id, "", "ProcessStatus")
	if err != nil {
		return err
	}

	var status temporal.Status
	if err := encoded.Get(&status); err != nil {
		return err
	}

	if !current.matchesRules(status, cart, rules) {
		return fmt.Errorf("gebundene Version für %s: DB %s Workflow %s statt %s", id, cart.RuleSetVersion, status.Binding.RuleSetVersion, rules)
	}

	return nil
}

func (current *world) matchesRules(status temporal.Status, cart sale.Cart, rules string) bool {
	return status.Binding.RuleSetVersion == rules && cart.RuleSetVersion == rules && status.Binding.DefinitionVersion == cart.DefinitionVersion
}
