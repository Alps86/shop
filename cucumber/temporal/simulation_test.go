package temporal_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) registerSimulation(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine ausschließlich simulierte Kartenfreigabe ausgeführt wird$`, current.simulation)
	ctx.Step(`^bindet das Simulationsergebnis ausschließlich den Kaufversuch$`, current.simulationObserved)
}

func (current *world) simulation(table *godog.Table) error {
	row := current.rows(table)[0]
	approved, err := strconv.ParseBool(row["Approved"])
	if err != nil {
		return err
	}

	delay, err := time.ParseDuration(row["Delay"])
	if err != nil {
		return err
	}

	deadline, err := time.ParseDuration(row["Deadline"])
	if err != nil {
		return err
	}

	return current.authorize(row, approved, delay, deadline)
}

func (current *world) authorize(row map[string]string, approved bool, delay, deadline time.Duration) error {
	revision, err := strconv.ParseUint(row["CartRevision"], 10, 64)
	if err != nil {
		return err
	}

	total, err := strconv.ParseInt(row["Total"], 10, 64)
	if err != nil {
		return err
	}

	cart := sale.Cart{ID: value.CartID(row["CartID"]), Revision: revision, Checkout: value.Optional[sale.Checkout]{Present: true, Value: sale.Checkout{ConfirmedEvaluation: value.Optional[sale.CartEvaluation]{Present: true, Value: sale.CartEvaluation{Total: value.Money{MinorUnits: total, Currency: "EUR"}}}}}}
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	current.authorization, current.err = temporal.NewPaymentSimulation(approved, delay).Authorize(ctx, cart, value.CommandID(row["CommandID"]))

	return nil
}

func (current *world) simulationObserved(table *godog.Table) error {
	code := value.Code(current.err)
	if errors.Is(current.err, context.DeadlineExceeded) {
		code = "DeadlineExceeded"
	}

	authorization := current.authorization
	actual := map[string]string{"Code": code, "Approved": strconv.FormatBool(authorization.Approved), "CartID": string(authorization.CartID), "CartRevision": strconv.FormatUint(authorization.CartRevision, 10), "Total": strconv.FormatInt(authorization.Total.MinorUnits, 10), "CommandID": string(authorization.CommandID)}
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}
