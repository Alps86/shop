package performance_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
	"strings"
)

func (current *world) prepareOrders() error {
	reused, err := current.reuseState()
	if err != nil {
		return err
	}

	if reused {
		return current.orderCount()
	}

	return current.newOrders()
}

func (current *world) newOrders() error {
	for _, profile := range strings.Split(current.data["ProfileCodes"], ",") {
		guest, err := current.service.Open(context.Background(), "")
		if err != nil {
			return err
		}

		selected, err := current.service.SelectProfile(context.Background(), guest.Session, application.ProfileCode(profile))
		if err != nil {
			return err
		}

		guest.Profile = selected.Profile
		current.guests = append(current.guests, guest)
	}

	if strconv.Itoa(len(current.guests)) != current.data["Profiles"] {
		return fmt.Errorf("Profilanzahl stimmt nicht")
	}

	return current.buyOrders()
}

func (current *world) buyOrders() error {
	count, err := strconv.Atoi(current.data["Orders"])
	if err != nil {
		return err
	}

	existing, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	for index := int(existing.Total); index < count; index++ {
		if err := current.buy(current.guests[index%len(current.guests)]); err != nil {
			return err
		}
	}

	return current.preparePrivateRoutes(count)
}

func (current *world) buy(guest application.GuestView) error {
	view, err := current.prepareCart(guest)
	if err != nil {
		return err
	}

	command := application.ConfirmCommand{CartID: view.Cart.ID, ExpectedRevision: view.Cart.Revision, EvaluationID: view.Cart.Evaluation.Value.ID, CommandID: commandID()}
	view, err = current.service.Confirm(context.Background(), guest.Session, command)
	if err != nil {
		return err
	}

	result, err := current.service.Place(context.Background(), guest.Session, application.PlaceCommand{CartID: view.Cart.ID, CommandID: commandID(), EvaluationID: view.Cart.Checkout.Value.ConfirmedEvaluation.Value.ID})
	if err != nil {
		return err
	}

	current.order = result.OrderID
	current.cookie = "shop_session=" + string(guest.Session)

	return nil
}

func (current *world) prepareCart(guest application.GuestView) (application.CartView, error) {
	ctx := context.Background()
	view, err := current.service.StartCart(ctx, guest.Session, commandID())
	if err != nil {
		return view, err
	}

	for _, change := range current.cartChanges() {
		change.CartID = view.Cart.ID
		change.ExpectedRevision = view.Cart.Revision
		change.CommandID = commandID()
		view, err = current.service.Change(ctx, guest.Session, change)
		if err != nil {
			return view, err
		}
	}

	return view, nil
}

func (current *world) cartChanges() []application.ChangeCommand {
	data := current.data
	address := sale.Address{Recipient: data["Recipient"], Street: data["Street"], PostalCode: data["PostalCode"], City: data["City"], Country: value.CountryCode(data["Country"])}

	return []application.ChangeCommand{{Operation: "add", SKU: value.SKU(data["SKU"]), Quantity: 1}, {Operation: "address", Address: address}, {Operation: "shipping", ShippingMethod: sale.ShippingMethod(data["Shipping"])}, {Operation: "payment", PaymentMethod: sale.PaymentMethod(data["Payment"])}}
}

func (current *world) preparePrivateRoutes(count int) error {
	observed, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	if observed.Total != int64(count) {
		return fmt.Errorf("Bestellungen %d statt %d", observed.Total, count)
	}

	if current.order == "" {
		return fmt.Errorf("Benchmarklaufzeit muss frisches Schema ohne frühere Bestellungen verwenden")
	}

	return current.prepareMeasuredCart()
}

func (current *world) prepareMeasuredCart() error {
	guest := current.guests[0]
	_, err := current.prepareCart(guest)
	current.data["CartCookie"] = "shop_session=" + string(guest.Session)
	if err != nil {
		return err
	}

	return current.saveState()
}

func commandID() value.CommandID {
	return value.CommandID("bench-" + strings.ToLower(rand.Text()))
}

func (current *world) orderCount() error {
	observed, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	if strconv.FormatInt(observed.Total, 10) != current.data["Orders"] {
		return fmt.Errorf("persistierte Bestellanzahl stimmt nicht; expliziter Reset erforderlich")
	}

	return nil
}
