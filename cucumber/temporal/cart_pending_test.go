package temporal_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerCartPending(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Application-Schreibantwort kontrolliert verzögert wird$`, current.delayCart)
	ctx.Step(`^der angenommene Cartbefehl später abgeschlossen wird$`, current.cartEventually)
}

func (current *world) delayCart(table *godog.Table) error {
	delay, err := time.ParseDuration(current.rows(table)[0]["Delay"])
	current.store.writeDelay = delay

	return err
}

func (current *world) cartEventually(table *godog.Table) error {
	row := current.rows(table)[0]
	var failure value.Error
	if !errors.As(current.err, &failure) {
		return fmt.Errorf("strukturierte Pendingantwort fehlt: %w", current.err)
	}

	if failure.Field != "cart/"+current.id {
		return fmt.Errorf("Prozessreferenz fehlt")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return current.awaitCartRevision(ctx, row["Revision"])
}

func (current *world) awaitCartRevision(ctx context.Context, revision string) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		cart, err := current.store.Get(ctx, value.CartID(current.id))
		if fmt.Sprint(cart.Revision) == revision {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
