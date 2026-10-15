package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (current *world) payment(view application.CartView, row map[string]string) (application.CartView, error) {
	if row["PaymentMethod"] == "" {
		return view, nil
	}

	current.paymentOutcome = row["Authorization"]

	return current.service.ApplyChange(context.Background(), application.ChangeCommand{UserID: current.guest.UserID, CartID: view.Cart.ID, ExpectedRevision: view.Cart.Revision, CommandID: "payment-a", Operation: "payment", PaymentMethod: sale.PaymentMethod(row["PaymentMethod"])})
}

func (current *world) Authorize(_ context.Context, cart sale.Cart, id value.CommandID) (application.PaymentAuthorization, error) {
	return application.PaymentAuthorization{Approved: current.paymentOutcome == "approved", CartID: cart.ID, CartRevision: cart.Revision, Total: cart.Checkout.Value.ConfirmedEvaluation.Value.Total, CommandID: id}, nil
}
