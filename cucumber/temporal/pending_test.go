package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) registerPending(ctx *godog.ScenarioContext) {
	ctx.Step(`^der angenommene Kauf nach dem Anfragebudget weiterläuft$`, current.pendingPurchase)
}

func (current *world) pendingPurchase(table *godog.Table) error {
	row := current.rows(table)[0]
	if err := current.confirmPurchase(row); err != nil {
		return err
	}

	current.purchase, current.err = current.service.Place(context.Background(), current.guest.Session, application.PlaceCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["PlaceCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID})
	if current.err != nil {
		return current.err
	}

	if strconv.FormatBool(current.purchase.Pending) != row["Pending"] {
		return fmt.Errorf("Pending %v statt %s", current.purchase.Pending, row["Pending"])
	}

	if current.purchase.ProcessReference != "cart/"+current.id {
		return fmt.Errorf("quittierte Prozessreferenz fehlt")
	}

	return current.awaitPurchase()
}

func (current *world) awaitPurchase() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		id, err := current.databaseStore.Orders().ForCart(ctx, value.CartID(current.id))
		if err == nil {
			current.purchase.OrderID = id
			return current.capturePurchase()
		}

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
		}
	}
}
