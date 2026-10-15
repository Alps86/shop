package http_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (double guestDouble) Open(_ context.Context, session application.Session) (application.GuestView, error) {
	result := double.current.fixture.Guest
	if session == "foreign" {
		return application.GuestView{}, value.NewError("NotFound", "Session")
	}

	return result, double.current.portError("Open")
}

func (double guestDouble) SelectProfile(_ context.Context, _ application.Session, profile application.ProfileCode) (application.GuestView, error) {
	result := double.current.fixture.Guest
	result.Profile = profile

	return result, double.current.portError("SelectProfile")
}

func (double guestDouble) GetCart(_ context.Context, session application.Session, id value.CartID) (application.CartView, error) {
	if session == "foreign" {
		return application.CartView{}, value.NewError("NotFound", "Cart")
	}

	return double.current.fixture.Cart, double.current.portError("GetCart")
}

func (double guestDouble) StartCart(context.Context, application.Session, value.CommandID) (application.CartView, error) {
	return double.current.fixture.Cart, double.current.portError("StartCart")
}

func (double guestDouble) Change(_ context.Context, _ application.Session, command application.ChangeCommand) (application.CartView, error) {
	double.current.lastChange = command
	return double.current.fixture.Cart, double.current.portError("Change")
}

func (double guestDouble) Evaluate(context.Context, application.Session, value.CartID) (application.CartView, error) {
	return double.current.fixture.Cart, double.current.portError("Evaluate")
}

func (double guestDouble) Confirm(_ context.Context, _ application.Session, command application.ConfirmCommand) (application.CartView, error) {
	double.current.lastConfirm = command
	return double.current.fixture.Cart, double.current.portError("Confirm")
}

func (double guestDouble) Place(_ context.Context, _ application.Session, command application.PlaceCommand) (application.PurchaseResult, error) {
	double.current.lastPlace = command
	return double.current.fixture.Purchase, double.current.portError("Place")
}

func (double guestDouble) GetOrder(_ context.Context, session application.Session, id value.OrderID) (application.OrderView, error) {
	if session == "foreign" {
		return application.OrderView{}, value.NewError("NotFound", "Order")
	}

	return double.current.fixture.Order, double.current.portError("GetOrder")
}

func (double guestDouble) GetInterests(context.Context, application.Session) ([]sale.CategoryInterest, error) {
	return double.current.fixture.Interests, double.current.portError("Interests")
}

func (double guestDouble) PreviewAccessory(context.Context, application.Session, value.CartID, value.SKU) (sale.CartEvaluation, error) {
	return double.current.fixture.Preview, double.current.portError("PreviewAccessory")
}
