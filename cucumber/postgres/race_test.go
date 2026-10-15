package postgres_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"sync"
)

func (current *world) raceSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^zwei getrennte Gäste dieselbe letzte Einheit bestätigt haben$`, current.twoGuests)
	ctx.Step(`^beide Gäste gleichzeitig abschließen$`, current.purchaseRace)
	ctx.Step(`^lautet das echte Konkurrenzresultat$`, current.raceObserved)
}

func (current *world) twoGuests(input *godog.Table) error {
	current.commands = nil
	for index := range 2 {
		shopper := world{store: current.store}
		if err := shopper.newCart(input); err != nil {
			return err
		}

		if err := shopper.checkout(input); err != nil {
			return err
		}

		current.commands = append(current.commands, application.PlaceCommand{UserID: shopper.guest.UserID, CartID: shopper.cart.Cart.ID, EvaluationID: shopper.cart.Cart.Evaluation.Value.ID, CommandID: value.CommandID(fmt.Sprintf("race-%d", index))})
	}

	current.service = application.Service{Carts: current.store.Carts(), Purchase: current.store.Purchases(), Clock: current}

	return nil
}

func (current *world) purchaseRace() {
	current.results = make([]application.PurchaseResult, len(current.commands))
	current.errors = make([]error, len(current.commands))
	var group sync.WaitGroup
	for index, command := range current.commands {
		group.Go(func() {
			current.results[index], current.errors[index] = current.service.ApplyPlace(context.Background(), command)
		})
	}

	group.Wait()
}

func (current *world) raceObserved(input *godog.Table) error {
	counts := map[string]int{}
	for _, err := range current.errors {
		counts[value.Code(err)]++
	}

	page, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	return check(table(input), map[string]string{"OK": strconv.Itoa(counts["OK"]), "Unavailable": strconv.Itoa(counts["Unavailable"]), "Orders": strconv.FormatInt(page.Total, 10)})
}
