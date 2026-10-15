package temporal_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) registerRetry(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Ablaufabhängigkeit kontrolliert scheitert$`, current.retryFailure)
	ctx.Step(`^das ausgeschöpfte Budget als technische Diagnose sichtbar bleibt$`, current.retryObserved)
}

func (current *world) retryFailure(table *godog.Table) error {
	row := current.rows(table)[0]
	attempts, err := strconv.Atoi(row["InitialFailures"])
	if err != nil {
		return err
	}

	current.store.expirationFailures = attempts
	current.store.expirationError = value.NewError(row["Code"], "Dependency")
	delay, err := time.ParseDuration(row["ObserveAfter"])
	if err != nil {
		return err
	}

	current.testEnvironment.RegisterDelayedCallback(current.queryAndCancel, delay)

	return nil
}

func (current *world) retryObserved(table *godog.Table) error {
	if current.queryErr != nil {
		return current.queryErr
	}

	actual := map[string]string{"Attempts": strconv.Itoa(current.store.expirationCalls), "State": current.processStatus.State}
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}
