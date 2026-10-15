package temporal_test

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/testsuite"
	"strconv"
	"time"
)

func (current *world) registerRecovery(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein erfolgreicher Auftrag nach Reparatur den Prozess ausdrücklich fortsetzt$`, current.recovery)
}

func (clock fixedClock) Now() time.Time {
	return clock.At
}

func (current *world) recovery(table *godog.Table) error {
	row := current.rows(table)[0]
	attempts, err := strconv.Atoi(row["InitialFailures"])
	if err != nil {
		return err
	}

	delay, err := time.ParseDuration(row["RetryAfter"])
	if err != nil {
		return err
	}

	current.store.expirationFailures = attempts
	current.store.expirationError = value.NewError(row["Code"], "Dependency")
	current.adapter.Service = application.Service{Carts: current.store, Clock: fixedClock{At: current.started}}
	current.testEnvironment.RegisterActivity(current.adapter.Evaluate)
	current.scheduleRecovery(delay)

	return nil
}

func (current *world) scheduleRecovery(delay time.Duration) {
	current.testEnvironment.RegisterDelayedCallback(func() {
		current.testEnvironment.UpdateWorkflow("EvaluateCart", "recovery-a", &testsuite.TestUpdateCallback{}, application.EvaluateCommand{UserID: "u-a", CartID: "c-timer"})
	}, delay)
}
