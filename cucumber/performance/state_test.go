package performance_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"os"
	"path/filepath"
	"strings"
)

func (current *world) statePath() string {
	if configured := os.Getenv("SHOP_PERFORMANCE_STATE_FILE"); configured != "" {
		return configured
	}

	return filepath.Join(os.TempDir(), "shop-benchmark-private-"+os.Getenv("SHOP_PERFORMANCE_SCHEMA")+".json")
}

func (current *world) reuseState() (bool, error) {
	data, err := os.ReadFile(current.statePath())
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	var state fixtureState
	if err := json.Unmarshal(data, &state); err != nil {
		return false, err
	}

	return true, current.restore(state)
}

func (current *world) restore(state fixtureState) error {
	if current.foreignState(state) {
		return fmt.Errorf("private Benchmarkfixture gehört zu anderer Laufzeit; expliziter Reset erforderlich")
	}

	current.guests = state.Guests
	current.order = state.Order
	current.cookie = state.OrderCookie
	current.data["CartCookie"] = state.CartCookie
	if err := current.validateGuests(); err != nil {
		return err
	}

	session := application.Session(strings.TrimPrefix(current.cookie, "shop_session="))
	_, err := current.service.GetOrder(context.Background(), session, current.order)

	return require(err, "persistierte Benchmarkorder über echten Besitzport")
}

func (current *world) validateGuests() error {
	profiles := strings.Split(current.data["ProfileCodes"], ",")
	if len(current.guests) != len(profiles) {
		return fmt.Errorf("persistierte Gastprofile unvollständig")
	}

	for index, guest := range current.guests {
		identity, err := current.service.Open(context.Background(), guest.Session)
		if err != nil {
			return err
		}

		if string(identity.Profile) != profiles[index] {
			return fmt.Errorf("persistiertes Gastprofil stimmt nicht")
		}
	}

	return current.validateCart()
}

func (current *world) validateCart() error {
	session := application.Session(strings.TrimPrefix(current.data["CartCookie"], "shop_session="))
	identity, err := current.service.Open(context.Background(), session)
	if err != nil {
		return err
	}

	if !identity.CartID.Present {
		return fmt.Errorf("persistierter Messwarenkorb fehlt")
	}

	_, err = current.service.GetCart(context.Background(), session, identity.CartID.Value)

	return err
}

func (current *world) saveState() error {
	state := fixtureState{Schema: os.Getenv("SHOP_PERFORMANCE_SCHEMA"), PublicURL: current.public, Guests: current.guests, Order: current.order, OrderCookie: current.cookie, CartCookie: current.data["CartCookie"]}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(current.statePath(), data, 0600)
}

func (current *world) foreignState(state fixtureState) bool {
	return state.Schema != os.Getenv("SHOP_PERFORMANCE_SCHEMA") || state.PublicURL != current.public
}
