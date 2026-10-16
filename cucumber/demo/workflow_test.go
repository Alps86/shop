package demo_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"go.temporal.io/api/serviceerror"
	"net/url"
)

func (current *world) stopWorkflows() error {
	deployment := current.deployment
	if deployment.keepExternalWorkflows() {
		return nil
	}

	err := current.terminate("order/" + string(current.order.Purchase.ID))
	for _, id := range current.orders {
		err = errors.Join(err, current.terminate("order/"+string(id)))
	}

	for _, client := range current.clients {
		endpoint, _ := url.Parse(deployment.public)
		for _, cookie := range client.Jar.Cookies(endpoint) {
			err = errors.Join(err, current.stopGuest(application.Session(cookie.Value), cookie.Name))
		}
	}

	return err
}

func (deployment *deployment) keepExternalWorkflows() bool {
	return deployment.external || deployment.app == nil
}

func (current *world) stopGuest(session application.Session, name string) error {
	if name != "shop_session" {
		return nil
	}

	identity, err := current.deployment.store.Guests().Resolve(context.Background(), session)
	if err != nil {
		return err
	}

	if !identity.CartID.Present {
		return nil
	}

	return current.terminate("cart/" + string(identity.CartID.Value))
}

func (current *world) terminate(id string) error {
	if id == "order/" {
		return nil
	}

	err := current.deployment.connection.TerminateWorkflow(context.Background(), id, "", "Eigene Demoabnahme beendet")
	var missing *serviceerror.NotFound
	if errors.As(err, &missing) {
		return nil
	}

	return err
}
