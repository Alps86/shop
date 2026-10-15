package application_test

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerCoordination(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Gastzugriff mit folgenden Grenzen besteht$`, current.accessState)
	ctx.Step(`^der öffentliche Gastvertrag ausgeführt wird$`, current.accessCall)
	ctx.Step(`^ist das beobachtete Zugriffsergebnis$`, current.accessObserved)
}

func (current *world) accessState(table *godog.Table) error {
	row := read(table)[0]
	issued, err := time.Parse(time.RFC3339, row["IssuedAt"])
	if err != nil {
		return err
	}

	current.sessions["a"] = application.SessionIdentity{UserID: "u-a", IssuedAt: issued, Profile: "neutral"}
	current.profiles["u-a"] = "neutral"
	current.carts["c-a"] = sale.Cart{ID: "c-a", UserID: value.UserID(row["Owner"]), State: sale.CartState(row["State"]), LastActivityAt: issued}
	current.orders["o-a"] = sale.Order{ID: "o-a", CartID: "c-a", UserID: value.UserID(row["Owner"])}
	current.fault = row["Fault"]
	current.service.Processes = processDouble{current}

	return nil
}

func (current *world) accessCall(table *godog.Table) {
	row := read(table)[0]
	current.coordinationOperation = row["Operation"]
	current.service.Processes = processDouble{current}
	operations := map[string]func(map[string]string){"open": current.accessOpen, "profile": current.accessProfile, "cart": current.accessCart, "order": current.accessOrder, "start": current.accessStart, "evaluate": current.accessEvaluate, "confirm": current.accessConfirm, "change": current.accessChange, "place": current.accessPlace}
	operations[row["Operation"]](row)
}

func (current *world) accessOpen(row map[string]string) {
	current.result, current.err = current.service.Open(context.Background(), application.Session(row["Session"]))
}

func (current *world) accessProfile(row map[string]string) {
	current.result, current.err = current.service.SelectProfile(context.Background(), application.Session(row["Session"]), application.ProfileCode(row["Argument"]))
}

func (current *world) accessCart(row map[string]string) {
	current.cartView, current.err = current.service.GetCart(context.Background(), application.Session(row["Session"]), value.CartID(row["Argument"]))
}

func (current *world) accessOrder(row map[string]string) {
	current.orderView, current.err = current.service.GetOrder(context.Background(), application.Session(row["Session"]), value.OrderID(row["Argument"]))
}

func (current *world) accessStart(row map[string]string) {
	current.cartView, current.err = current.service.StartCart(context.Background(), application.Session(row["Session"]), value.CommandID(row["Argument"]))
}

func (current *world) accessEvaluate(row map[string]string) {
	current.cartView, current.err = current.service.Evaluate(context.Background(), application.Session(row["Session"]), value.CartID(row["Argument"]))
}

func (current *world) accessConfirm(row map[string]string) {
	revision, err := strconv.ParseUint(row["ExpectedRevision"], 10, 64)
	if err != nil {
		current.err = err
		return
	}

	current.cartView, current.err = current.service.Confirm(context.Background(), application.Session(row["Session"]), application.ConfirmCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["Argument"]), EvaluationID: value.EvaluationID(row["EvaluationID"]), ExpectedRevision: revision, CommandID: value.CommandID(row["CommandID"])})
}

func (current *world) accessChange(row map[string]string) {
	revision, err := strconv.ParseUint(row["ExpectedRevision"], 10, 64)
	if err != nil {
		current.err = err
		return
	}

	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		current.err = err
		return
	}

	current.cartView, current.err = current.service.Change(context.Background(), application.Session(row["Session"]), application.ChangeCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["Argument"]), CommandID: value.CommandID(row["CommandID"]), Operation: row["Change"], SKU: value.SKU(row["SKU"]), Quantity: quantity, ExpectedRevision: revision})
}

func (current *world) accessPlace(row map[string]string) {
	current.purchaseResult, current.err = current.service.Place(context.Background(), application.Session(row["Session"]), application.PlaceCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["Argument"]), CommandID: value.CommandID(row["CommandID"]), EvaluationID: value.EvaluationID(row["EvaluationID"])})
}

func (current *world) accessObserved(table *godog.Table) error {
	return read(table).matches(map[string]string{
		"CartQuantity": strconv.FormatInt(current.cartView.Cart.Items.Quantity("PHONE"), 10), "Confirmed": strconv.FormatBool(current.cartView.Cart.Checkout.Value.ConfirmedEvaluation.Present),
		"Code": value.Code(current.err), "GuestUser": string(current.result.UserID), "GuestSession": string(current.result.Session), "Profile": string(current.result.Profile),
		"StoredProfile": string(current.profiles["u-a"]), "CartID": string(current.cartView.Cart.ID), "CartUser": string(current.cartView.Cart.UserID), "OrderID": string(current.orderView.Purchase.ID),
		"SessionCount": strconv.Itoa(len(current.sessions)), "ProcessCalls": strconv.Itoa(current.processCalls), "ProcessUser": string(current.processUser), "ContainsSession": strconv.FormatBool(strings.Contains(current.processPayload, "Session")),
	})
}
