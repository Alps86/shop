package temporal_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/api/serviceerror"
)

func (current *world) registerOperation(ctx *godog.ScenarioContext) {
	ctx.Step(`^kein neuer Cartworkflow beobachtet wird$`, current.noWorkflow)
	ctx.Step(`^eine typisierte Temporaloperation verarbeitet wird$`, current.operation)
}

func (current *world) operation(table *godog.Table) error {
	row := table.Rows[1].Cells
	user := value.UserID(row[1].Value)
	id := value.CartID(current.id)
	operations := map[string]func(){
		"evaluate": func() {
			current.view, current.err = current.adapter.EvaluateCart(context.Background(), application.EvaluateCommand{UserID: user, CartID: id})
		},
		"confirm": func() {
			current.view, current.err = current.adapter.ConfirmEvaluation(context.Background(), application.ConfirmCommand{UserID: user, CartID: id, CommandID: "confirm-a", EvaluationID: "eval-a", ExpectedRevision: 1})
		},
		"place": func() {
			current.purchase, current.err = current.adapter.PlaceOrder(context.Background(), application.PlaceCommand{UserID: user, CartID: id, CommandID: "place-a", EvaluationID: "eval-a"})
		},
	}

	operations[row[0].Value]()

	return nil
}

func (current *world) noWorkflow(table *godog.Table) error {
	_, err := current.connection.DescribeWorkflowExecution(context.Background(), "cart/"+current.id, "")
	var missing *serviceerror.NotFound
	actual := map[string]string{"Code": value.Code(current.err), "Exists": "true"}
	if errors.As(err, &missing) {
		actual["Exists"] = "false"
	}

	for key, want := range current.rows(table)[0] {
		if actual[key] != want {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], want)
		}
	}

	return nil
}
