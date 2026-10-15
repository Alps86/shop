package application_test

import (
	"context"
	"strconv"
	"strings"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerActivities(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Warenkorbsicht für den aktuellen Zustand erstellt wird$`, current.actionView)
	ctx.Step(`^nennt die Warenkorbsicht die erlaubten Aktionen$`, current.actionsObserved)
	ctx.Step(`^ein autorisierter Activityaufruf verarbeitet wird$`, current.activity)
	ctx.Step(`^lautet das Activityergebnis$`, current.code)
	ctx.Step(`^eine Liefermeldung an einer technischen Grenze scheitert$`, current.deliveryFailure)
}

func (current *world) activity(table *godog.Table) error {
	row := read(table)[0]
	current.fault = row["Fault"]
	current.authorization = row["Fault"]
	cart := current.carts["c-a"]
	cart.State = sale.CartState(row["State"])
	if cart.State == sale.Ordered {
		current.orders["o-a"] = sale.Order{ID: "o-a", CartID: cart.ID, UserID: cart.UserID}
	}

	cart.Checkout.Present = row["CheckoutPresent"] == "true"
	cart.Checkout.Value.ConfirmedEvaluation.Present = row["Confirmed"] == "true"
	cart.Checkout.Value.PaymentMethod = sale.PaymentMethod(row["Payment"])
	current.carts["c-a"] = cart
	dispatch := map[string]func(map[string]string){"evaluate": current.activityEvaluate, "confirm": current.activityConfirm, "place": current.activityPlace}
	if operation, exists := dispatch[row["Operation"]]; exists {
		operation(row)
		return nil
	}

	return current.activityChange(row)
}

func (current *world) activityEvaluate(row map[string]string) {
	current.cartView, current.err = current.service.ApplyEvaluate(context.Background(), application.EvaluateCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"])})
}

func (current *world) activityConfirm(row map[string]string) {
	command := application.ConfirmCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"]), EvaluationID: value.EvaluationID(row["EvaluationID"]), ExpectedRevision: 4, CommandID: "cmd-confirm"}
	current.cartView, current.err = current.service.ApplyConfirm(context.Background(), command)
}

func (current *world) activityPlace(row map[string]string) {
	command := application.PlaceCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"]), EvaluationID: value.EvaluationID(row["EvaluationID"]), CommandID: "cmd-place"}
	current.purchaseResult, current.err = current.service.ApplyPlace(context.Background(), command)
}

func (current *world) activityChange(row map[string]string) error {
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	command := application.ChangeCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"]), ExpectedRevision: 4, CommandID: "cmd-change", Operation: row["Operation"], SKU: "PHONE", Quantity: quantity, Address: sale.Address{Recipient: "Demo Gast", Street: "Testweg 1", PostalCode: "12345", City: "Teststadt", Country: "DE"}}
	current.cartView, current.err = current.service.ApplyChange(context.Background(), command)

	return nil
}

func (current *world) deliveryFailure(table *godog.Table) error {
	current.fault = read(table)[0]["Fault"]
	return current.reportDelivery(table)
}

func (current *world) actionView(table *godog.Table) {
	row := read(table)[0]
	cart := current.carts["c-a"]
	cart.State = sale.CartState(row["State"])
	if cart.State == sale.Ordered {
		current.orders["o-a"] = sale.Order{ID: "o-a", CartID: cart.ID, UserID: cart.UserID}
	}

	cart.Checkout.Value.ConfirmedEvaluation.Present = row["Confirmed"] == "true"
	current.carts["c-a"] = cart
	current.cartView, current.err = current.service.GetCart(context.Background(), "a", "c-a")
}

func (current *world) actionsObserved(table *godog.Table) error {
	return read(table).matches(map[string]string{"Actions": strings.Join(current.cartView.Actions, ",")})
}
