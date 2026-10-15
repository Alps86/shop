package application_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"reflect"
	"strconv"
)

func (current *world) registerReplay(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein bereits gespeicherter Fachbefehl erneut als Activity ausgeführt wird$`, current.retryCommand)
	ctx.Step(`^eine gespeicherte Bestätigung erneut als Activity ausgeführt wird$`, current.retryConfirmation)
	ctx.Step(`^lautet das wiederholte Activityergebnis$`, current.replayObserved)
}

func (current *world) retryCommand(input *godog.Table) error {
	row := read(input)[0]
	current.service.Commands = commandReader{current}
	command := application.ChangeCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"]), CommandID: value.CommandID(row["CommandID"]), Operation: "setQuantity", SKU: value.SKU(row["SKU"]), Quantity: 2, ExpectedRevision: 4}
	current.cartView, current.err = current.service.ApplyChange(context.Background(), command)
	if current.err != nil {
		return current.err
	}

	quantity, err := strconv.ParseInt(row["RetryQuantity"], 10, 64)
	if err != nil {
		return err
	}

	command.Quantity = quantity
	current.cartView, current.err = current.service.ApplyChange(context.Background(), command)

	return nil
}

func (current *world) replayObserved(input *godog.Table) error {
	return read(input).matches(map[string]string{"Code": value.Code(current.err), "Quantity": strconv.FormatInt(current.carts["c-a"].Items.Quantity("PHONE"), 10)})
}

func (port commandReader) ReplayChange(_ context.Context, command application.ChangeCommand) (application.CartView, bool, error) {
	if port.current.fault == "commandRead" {
		return application.CartView{}, false, value.NewError("DependencyUnavailable", "Command")
	}

	write, found := port.current.commandWrites[command.CommandID]
	if !found {
		return application.CartView{}, false, nil
	}

	if !reflect.DeepEqual(write.Change.Value, command) {
		return application.CartView{}, true, value.NewError("Conflict", "CommandID")
	}

	return application.NewCartView(write.Cart), true, nil
}

func (port commandReader) ReplayConfirm(_ context.Context, command application.ConfirmCommand) (application.CartView, bool, error) {
	if port.current.fault == "commandRead" {
		return application.CartView{}, false, value.NewError("DependencyUnavailable", "Command")
	}

	write, found := port.current.commandWrites[command.CommandID]
	if !found {
		return application.CartView{}, false, nil
	}

	if !reflect.DeepEqual(write.Confirmation.Value, command) {
		return application.CartView{}, true, value.NewError("Conflict", "CommandID")
	}

	return application.NewCartView(write.Cart), true, nil
}

func (current *world) retryConfirmation(input *godog.Table) error {
	row := read(input)[0]
	current.service.Commands = commandReader{current}
	command := application.ConfirmCommand{UserID: value.UserID(row["UserID"]), CartID: value.CartID(row["CartID"]), CommandID: value.CommandID(row["CommandID"]), EvaluationID: value.EvaluationID(row["EvaluationID"]), ExpectedRevision: 4}
	current.cartView, current.err = current.service.ApplyConfirm(context.Background(), command)
	if current.err != nil {
		return current.err
	}

	command.EvaluationID = value.EvaluationID(row["RetryEvaluationID"])
	current.fault = row["Fault"]
	current.cartView, current.err = current.service.ApplyConfirm(context.Background(), command)

	return nil
}
