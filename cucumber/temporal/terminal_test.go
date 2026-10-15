package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerTerminal(ctx *godog.ScenarioContext) {
	ctx.Step(`^der abgelaufene Cartworkflow terminal abgeschlossen ist$`, current.cartClosed)
	ctx.Step(`^die Gastfassade terminale Fachfehler ohne Workflowupdate liefert$`, current.terminalOperations)
}

func (current *world) startGuestCart(row map[string]string) error {
	if row["Age"] == "" {
		current.view, current.err = current.service.StartCart(context.Background(), current.guest.Session, value.CommandID(row["CommandID"]))
		return current.err
	}

	age, err := time.ParseDuration(row["Age"])
	if err != nil {
		return err
	}

	current.view, current.err = current.databaseStore.Carts().Start(context.Background(), current.guest.UserID, value.CommandID(row["CommandID"]), time.Now().Add(-age))

	return current.err
}

func (current *world) cartClosed() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return current.connection.GetWorkflow(ctx, "cart/"+current.id, "").Get(ctx, nil)
}

func (current *world) terminalOperations(table *godog.Table) error {
	for _, row := range current.rows(table) {
		current.terminalOperation(row["Operation"])
		if value.Code(current.err) != row["Code"] {
			return fmt.Errorf("%s: %s statt %s", row["Operation"], value.Code(current.err), row["Code"])
		}
	}

	return nil
}

func (current *world) terminalOperation(operation string) {
	ctx := context.Background()
	operations := map[string]func(){
		"change": func() {
			current.view, current.err = current.service.Change(ctx, current.guest.Session, application.ChangeCommand{CartID: value.CartID(current.id), CommandID: "terminal-change", ExpectedRevision: 1, Operation: "add", SKU: "PHONE", Quantity: 1})
		},
		"evaluate": func() {
			current.view, current.err = current.service.Evaluate(ctx, current.guest.Session, value.CartID(current.id))
		},
		"confirm": func() {
			current.view, current.err = current.service.Confirm(ctx, current.guest.Session, application.ConfirmCommand{CartID: value.CartID(current.id), CommandID: "terminal-confirm", EvaluationID: "missing", ExpectedRevision: 1})
		},
		"place": func() {
			current.purchase, current.err = current.service.Place(ctx, current.guest.Session, application.PlaceCommand{CartID: value.CartID(current.id), CommandID: "terminal-place", EvaluationID: "missing"})
		},
	}

	operations[operation]()
}
