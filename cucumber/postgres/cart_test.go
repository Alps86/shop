package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) cartSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^der persistente Ablauf unter erneuter Sperre geprüft wird$`, current.expireCart)
	ctx.Step(`^derselbe Gast ausdrücklich einen Einkauf startet$`, current.startAgain)
	ctx.Step(`^der Cartschreibport eine typisierte Originaländerung erhält$`, current.directCartWrite)
	ctx.Step(`^der dauerhafte Cartstarter eine explizite Userkennung erhält$`, current.directStart)
}

func (current *world) expireCart(input *godog.Table) error {
	row := table(input)
	hours, err := strconv.Atoi(row["Hours"])
	if err != nil {
		return err
	}

	id := current.cart.Cart.ID
	if row["ID"] == "missing" {
		id = "c-missing"
	}

	cart, err := current.store.Carts().Expire(context.Background(), id, current.cart.Cart.LastActivityAt.Add(time.Duration(hours)*time.Hour))
	current.err = err
	current.observed = map[string]string{"State": string(cart.State), "Confirmed": strconv.FormatBool(cart.Checkout.Value.ConfirmedEvaluation.Present)}

	return nil
}

func (current *world) startAgain(input *godog.Table) error {
	row := table(input)
	view, err := current.service.StartCart(context.Background(), current.guest.Session, value.CommandID(row["CommandID"]))
	current.err = err
	var carts, sessions int
	if err := current.admin.QueryRow(context.Background(), "SELECT (SELECT count(*) FROM "+current.schema+".carts),(SELECT count(*) FROM "+current.schema+".sessions)").Scan(&carts, &sessions); err != nil {
		return err
	}

	current.observed = map[string]string{"OwnUser": strconv.FormatBool(view.Cart.UserID == current.guest.UserID), "OriginalCart": strconv.FormatBool(view.Cart.ID == current.cart.Cart.ID), "CartCount": strconv.Itoa(carts), "SessionCount": strconv.Itoa(sessions)}

	return nil
}

func (current *world) directCartWrite(input *godog.Table) error {
	row := table(input)
	revision, err := strconv.ParseUint(row["ExpectedRevision"], 10, 64)
	if err != nil {
		return err
	}

	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	command := application.ChangeCommand{UserID: current.guest.UserID, CartID: current.cart.Cart.ID, CommandID: value.CommandID(row["CommandID"]), ExpectedRevision: 1, Operation: "add", SKU: "PHONE", Quantity: quantity}
	write := application.CartWrite{Cart: current.cart.Cart, ExpectedRevision: revision, CommandID: command.CommandID, Change: value.Optional[application.ChangeCommand]{Present: true, Value: command}}
	_, current.err = current.store.Carts().Save(context.Background(), write)

	return nil
}

func (current *world) directStart(input *godog.Table) {
	row := table(input)
	_, current.err = current.store.Carts().Start(context.Background(), value.UserID(row["UserID"]), value.CommandID(row["CommandID"]), current.Now())
}
