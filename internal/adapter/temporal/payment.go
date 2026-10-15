package temporal

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

func NewPaymentSimulation(approved bool, delay time.Duration) PaymentSimulation {
	return PaymentSimulation{Approved: approved, Delay: delay}
}

func (simulation PaymentSimulation) Authorize(ctx context.Context, cart sale.Cart, id value.CommandID) (application.PaymentAuthorization, error) {
	timer := time.NewTimer(simulation.Delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return application.PaymentAuthorization{}, ctx.Err()
	case <-timer.C:
		return application.PaymentAuthorization{CartID: cart.ID, CartRevision: cart.Revision, Total: cart.Checkout.Value.ConfirmedEvaluation.Value.Total, CommandID: id, Approved: simulation.Approved}, nil
	}
}
