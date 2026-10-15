package temporal_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerActivityTime(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine gespeicherte Aktivitätsbasis mit Commandbeleg besteht$`, current.seedActivity)
	ctx.Step(`^eine zeitbezogene Prozessoperation auf PostgreSQL ausgeführt wird$`, current.activityOperation)
	ctx.Step(`^nur erfolgreiche neue Nutzeränderungen die Aktivität verschieben$`, current.activityObserved)
}

func (current *world) seedActivity(table *godog.Table) error {
	row := current.rows(table)[0]
	current.activityBase = current.view.Cart.LastActivityAt
	current.clockNow = current.activityBase
	current.activityCommand = application.ChangeCommand{CartID: current.view.Cart.ID, ExpectedRevision: current.view.Cart.Revision, CommandID: value.CommandID(row["CommandID"]), Operation: "shipping", ShippingMethod: "standard"}
	current.view, current.err = current.service.Change(context.Background(), current.guest.Session, current.activityCommand)

	return current.err
}

func (current *world) activityOperation(table *godog.Table) error {
	row := current.rows(table)[0]
	offset, err := time.ParseDuration(row["AtOffset"])
	if err != nil {
		return err
	}

	current.clockNow = current.activityBase.Add(offset)
	operations := map[string]func(map[string]string) error{"read": current.activityRead, "change": current.activityChange, "place": current.activityPlace, "timer": current.activityExpire}
	current.err = operations[row["Operation"]](row)

	return nil
}

func (current *world) activityRead(_ map[string]string) error {
	var err error
	current.view, err = current.service.GetCart(context.Background(), current.guest.Session, value.CartID(current.id))

	return err
}

func (current *world) activityChange(row map[string]string) error {
	command := current.activityCommand
	if row["Duplicate"] != "true" {
		command.CommandID = value.CommandID(row["CommandID"])
		command.ExpectedRevision = current.view.Cart.Revision
	}

	if row["Valid"] != "true" {
		command.Operation = "add"
	}

	var err error
	current.view, err = current.service.Change(context.Background(), current.guest.Session, command)

	return err
}

func (current *world) activityPlace(row map[string]string) error {
	_, err := current.service.Place(context.Background(), current.guest.Session, application.PlaceCommand{CartID: value.CartID(current.id), CommandID: value.CommandID(row["CommandID"]), EvaluationID: "not-confirmed"})
	_, expireErr := current.adapter.Expire(context.Background(), value.CartID(current.id), current.clockNow)

	return errors.Join(err, expireErr)
}

func (current *world) activityExpire(row map[string]string) error {
	_, err := current.adapter.Expire(context.Background(), value.CartID(current.id), current.clockNow)
	if err != nil {
		return err
	}

	return current.activityRead(row)
}

func (current *world) activityObserved(table *godog.Table) error {
	row := current.rows(table)[0]
	cart, err := current.databaseStore.Carts().Get(context.Background(), value.CartID(current.id))
	if err != nil {
		return err
	}

	expected, err := time.ParseDuration(row["LastActivityOffset"])
	if err != nil {
		return err
	}

	if !cart.LastActivityAt.Equal(current.activityBase.Add(expected)) {
		return fmt.Errorf("LastActivityAt %s statt Basis plus %s", cart.LastActivityAt, expected)
	}

	return current.activityResult(row, string(cart.State))
}

func (current *world) activityResult(row map[string]string, state string) error {
	actual := map[string]string{"Code": value.Code(current.err), "State": state}
	for key, wanted := range actual {
		if row[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, wanted, row[key])
		}
	}

	return nil
}
