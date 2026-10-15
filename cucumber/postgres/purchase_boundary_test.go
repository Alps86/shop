package postgres_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) purchaseBoundarySteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Kartenzahlung im Cart neu bestätigt wird$`, current.confirmCard)
	ctx.Step(`^die direkte Kauftransaktionsgrenze aufgerufen wird$`, current.directPurchase)
}

func (current *world) confirmCard(input *godog.Table) error {
	view, err := current.payment(current.cart, table(input))
	if err != nil {
		return err
	}

	current.cart, err = current.service.ApplyConfirm(context.Background(), application.ConfirmCommand{UserID: current.guest.UserID, CartID: view.Cart.ID, CommandID: "confirm-card", EvaluationID: view.Cart.Evaluation.Value.ID, ExpectedRevision: view.Cart.Revision})

	return err
}

func (current *world) directPurchase(input *godog.Table) error {
	row := table(input)
	cart, err := current.preparePurchaseCart(row)
	if err != nil {
		return err
	}

	command := application.PurchaseCommand{UserID: cart.UserID, CartID: cart.ID, CommandID: value.CommandID(row["CommandID"]), EvaluationID: cart.Checkout.Value.ConfirmedEvaluation.Value.ID, Now: current.Now()}
	command = current.purchaseIdentity(command, row)

	command.Authorization = current.purchaseAuthorization(cart, command.CommandID, row["Authorization"])
	current.result, current.err = current.store.Purchases().Commit(context.Background(), command)

	return nil
}

func (current *world) purchaseIdentity(command application.PurchaseCommand, row map[string]string) application.PurchaseCommand {
	if row["User"] == "foreign" {
		command.UserID = "u-foreign"
	}

	if row["Cart"] == "missing" {
		command.CartID = "c-missing"
	}

	if row["Evaluation"] == "other" {
		command.EvaluationID = "e-other"
	}

	return command
}

func (current *world) preparePurchaseCart(row map[string]string) (sale.Cart, error) {
	cart, err := current.store.Carts().Get(context.Background(), current.cart.Cart.ID)
	if err != nil {
		return sale.Cart{}, err
	}

	cart, err = current.agePurchaseCart(cart, row)
	if err != nil {
		return sale.Cart{}, err
	}

	if row["EmptyCart"] == "true" {
		return current.emptyPurchaseCart(cart)
	}

	_, err = current.admin.Exec(context.Background(), "UPDATE "+current.schema+".carts SET payload=$1 WHERE id=$2", cart, cart.ID)

	return cart, err
}

func (current *world) emptyPurchaseCart(cart sale.Cart) (sale.Cart, error) {
	cart.Items = nil
	evaluation, err := current.store.Evaluator().Evaluate(context.Background(), cart)
	if err != nil {
		return sale.Cart{}, err
	}

	cart.Evaluation = value.Optional[sale.CartEvaluation]{Present: true, Value: evaluation}
	cart.Checkout.Value.ConfirmedEvaluation = value.Optional[sale.CartEvaluation]{Present: true, Value: evaluation}
	_, err = current.admin.Exec(context.Background(), "UPDATE "+current.schema+".carts SET payload=$1 WHERE id=$2", cart, cart.ID)

	return cart, err
}

func (current *world) purchaseAuthorization(cart sale.Cart, id value.CommandID, mode string) value.Optional[application.PaymentAuthorization] {
	if mode == "absent" {
		return value.Optional[application.PaymentAuthorization]{}
	}

	authorization := application.PaymentAuthorization{CartID: cart.ID, CartRevision: cart.Revision, CommandID: id, Total: cart.Checkout.Value.ConfirmedEvaluation.Value.Total, Approved: mode != "declined"}
	authorization = current.alterAuthorization(authorization, mode)

	return value.Optional[application.PaymentAuthorization]{Present: true, Value: authorization}
}

func (current *world) alterAuthorization(authorization application.PaymentAuthorization, mode string) application.PaymentAuthorization {
	if mode == "wrongCart" {
		authorization.CartID = "c-other"
	}

	if mode == "wrongRevision" {
		authorization.CartRevision++
	}

	if mode == "wrongCommand" {
		authorization.CommandID = "cmd-other"
	}

	if mode == "wrongTotal" {
		authorization.Total.MinorUnits++
	}

	return authorization
}

func (current *world) agePurchaseCart(cart sale.Cart, row map[string]string) (sale.Cart, error) {
	hours, hoursErr := strconv.ParseInt(row["AgeHours"], 10, 64)
	delta, deltaErr := strconv.ParseUint(row["ConfirmationRevisionDelta"], 10, 64)
	if err := errors.Join(hoursErr, deltaErr); err != nil {
		return cart, err
	}

	cart.LastActivityAt = current.Now().Add(-time.Duration(hours) * time.Hour)
	if row["Confirmed"] == "false" {
		cart.Checkout.Value.ConfirmedEvaluation = value.Optional[sale.CartEvaluation]{}
	}

	cart.Checkout.Value.ConfirmedEvaluation.Value.CartRevision += delta

	return cart, nil
}
