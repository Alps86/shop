package application_test

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) processPlace(_ *godog.Table) {
	current.service.Processes = processDouble{current}
	current.purchaseResult, current.err = current.service.Place(context.Background(), "a", current.placeCommand)
}

func (current *world) processObserved(table *godog.Table) error {
	return read(table).matches(map[string]string{"ProcessCalls": strconv.Itoa(current.processCalls)})
}

func (port processDouble) EnsureCart(context.Context, value.CartID, application.ProcessBinding) error {
	return nil
}

func (port processDouble) UpdateCart(ctx context.Context, command application.ChangeCommand) (application.CartView, error) {
	if err := port.observe(command.UserID, command); err != nil {
		return application.CartView{}, err
	}

	return port.current.service.ApplyChange(ctx, command)
}

func (port processDouble) EvaluateCart(ctx context.Context, command application.EvaluateCommand) (application.CartView, error) {
	if err := port.observe(command.UserID, command); err != nil {
		return application.CartView{}, err
	}

	return port.current.service.ApplyEvaluate(ctx, command)
}

func (port processDouble) ConfirmEvaluation(ctx context.Context, command application.ConfirmCommand) (application.CartView, error) {
	if err := port.observe(command.UserID, command); err != nil {
		return application.CartView{}, err
	}

	return port.current.service.ApplyConfirm(ctx, command)
}

func (port processDouble) PlaceOrder(ctx context.Context, command application.PlaceCommand) (application.PurchaseResult, error) {
	if err := port.observe(command.UserID, command); err != nil {
		return application.PurchaseResult{}, err
	}

	return port.current.service.ApplyPlace(ctx, command)
}

func (port processDouble) observe(user value.UserID, command any) error {
	port.current.processCalls++
	port.current.processUser = user
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	port.current.processPayload = string(payload)
	if strings.Contains(string(payload), "Session") {
		return value.NewError("InvalidInput", "PersistedSession")
	}

	return nil
}

func (port processDouble) EnsureOrder(context.Context, value.OrderID, application.ProcessBinding) error {
	return nil
}

func (port processDouble) ReportDelivery(context.Context, application.DeliveryCommand) error {
	return nil
}
