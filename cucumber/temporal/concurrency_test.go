package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) registerConcurrency(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine neue Revision während der simulierten Freigabe entsteht$`, current.competingRevision)
	ctx.Step(`^der alte Kaufversuch keine Bestellung oder Reservierung hinterlässt$`, current.conflictObserved)
}

func (payment gatedPayment) Authorize(ctx context.Context, cart sale.Cart, id value.CommandID) (application.PaymentAuthorization, error) {
	close(payment.Started)
	select {
	case <-payment.Resume:
		return payment.Payment.Authorize(ctx, cart, id)
	case <-ctx.Done():
		return application.PaymentAuthorization{}, ctx.Err()
	}
}

func (current *world) competingRevision(table *godog.Table) error {
	row := current.rows(table)[0]
	if err := current.confirmPurchase(row); err != nil {
		return err
	}

	current.lastPlace = application.PlaceCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["PlaceCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID}
	done := make(chan struct{})
	go func() {
		current.purchase, current.err = current.service.Place(context.Background(), current.guest.Session, current.lastPlace)
		close(done)
	}()
	<-current.gate.Started
	err := current.competingChange(row)
	close(current.gate.Resume)
	<-done

	return err
}

func (current *world) confirmPurchase(row map[string]string) error {
	current.view, current.err = current.service.Confirm(context.Background(), current.guest.Session, application.ConfirmCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["ConfirmCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID, ExpectedRevision: current.view.Cart.Revision})

	return current.err
}

func (current *world) competingChange(row map[string]string) error {
	quantity, err := strconv.ParseInt(row["NewQuantity"], 10, 64)
	if err != nil {
		return err
	}

	_, err = current.service.Change(context.Background(), current.guest.Session, application.ChangeCommand{CartID: current.view.Cart.ID, ExpectedRevision: current.view.Cart.Revision, CommandID: value.CommandID(row["ChangeCommandID"]), Operation: "setQuantity", SKU: value.SKU(row["SKU"]), Quantity: quantity})

	return err
}

func (current *world) conflictObserved(table *godog.Table) error {
	page, err := current.databaseStore.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	snapshot, err := current.databaseStore.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	actual := map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.FormatInt(page.Total, 10), "Reserved": strconv.FormatInt(snapshot.Inventory[0].ReservedQuantity(), 10)}

	return current.conflictDetails(table, actual)
}

func (current *world) conflictDetails(table *godog.Table, actual map[string]string) error {
	cart, err := current.databaseStore.Carts().Get(context.Background(), value.CartID(current.id))
	if err != nil {
		return err
	}

	actual["RequiresConfirmation"] = strconv.FormatBool(!cart.Checkout.Value.ConfirmedEvaluation.Present)
	_, err = current.service.Place(context.Background(), current.guest.Session, current.lastPlace)
	actual["RetryCode"] = value.Code(err)
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}
