package commerce_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
)

func (current *world) checkout(data row) error {
	err := current.completePurchase(data)
	current.codes[data["Operation"]] = value.Code(err)
	if current.codes[data["Operation"]] != "OK" {
		return nil
	}

	return current.purchaseObserved(data)
}

func (current *world) purchaseObserved(data row) error {
	order, err := current.app.Service.GetOrder(context.Background(), current.guest.Session, current.purchase.OrderID)
	if err != nil {
		return err
	}

	page, err := current.app.Store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	state, err := current.orderProcess()
	if err != nil {
		return err
	}

	current.answers[data["Operation"]] = map[string]string{"ProcessState": state, "OrderCount": strconv.FormatInt(page.Total, 10), "Quantity": strconv.FormatInt(order.Purchase.Items[0].Quantity, 10), "Total": strconv.FormatInt(order.Purchase.Total.MinorUnits, 10), "Pending": strconv.FormatBool(current.purchase.Pending), "SKU": string(order.Purchase.Items[0].SKU)}

	return nil
}

func (current *world) completePurchase(data row) error {
	if err := current.startCart(data); err != nil {
		return err
	}

	commands, err := data.purchaseChanges()
	if err != nil {
		return err
	}

	return current.applyPurchaseChanges(data, commands)
}

func (current *world) applyPurchaseChanges(data row, commands []application.ChangeCommand) error {
	for _, command := range commands {
		command.CartID = current.view.Cart.ID
		command.ExpectedRevision = current.view.Cart.Revision
		var err error
		current.view, err = current.app.Service.Change(context.Background(), current.guest.Session, command)
		if err != nil {
			return err
		}
	}

	return current.confirmPurchase(data)
}

func (data row) purchaseChanges() ([]application.ChangeCommand, error) {
	quantity, err := strconv.ParseInt(data["Quantity"], 10, 64)
	if err != nil {
		return nil, err
	}

	return []application.ChangeCommand{
		{CommandID: value.CommandID(data["ChangeCommandID"]), Operation: "add", SKU: value.SKU(data["SKU"]), Quantity: quantity},
		{CommandID: value.CommandID(data["AddressCommandID"]), Operation: "address", Address: sale.Address{Recipient: data["Recipient"], Street: data["Street"], PostalCode: data["PostalCode"], City: data["City"], Country: value.CountryCode(data["Country"])}},
		{CommandID: value.CommandID(data["PaymentCommandID"]), Operation: "payment", PaymentMethod: sale.PaymentMethod(data["PaymentMethod"])},
	}, nil
}

func (current *world) confirmPurchase(data row) error {
	ctx := context.Background()
	view, err := current.app.Service.Evaluate(ctx, current.guest.Session, current.view.Cart.ID)
	if err != nil {
		return err
	}

	current.view, err = current.app.Service.Confirm(ctx, current.guest.Session, application.ConfirmCommand{CartID: view.Cart.ID, CommandID: value.CommandID(data["ConfirmCommandID"]), ExpectedRevision: view.Cart.Revision, EvaluationID: view.Cart.Evaluation.Value.ID})
	if err != nil {
		return err
	}

	current.purchase, err = current.app.Service.Place(ctx, current.guest.Session, application.PlaceCommand{CartID: view.Cart.ID, CommandID: value.CommandID(data["PlaceCommandID"]), EvaluationID: view.Cart.Evaluation.Value.ID})

	return err
}
