package http_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) registerBudget(ctx *godog.ScenarioContext) {
	ctx.Step("^die Fragmentansicht mit ihrer Laufzeit angefordert wird$", current.timedFragment)
	ctx.Step("^die zusammengesetzte Ansicht mit ihrer Laufzeit angefordert wird$", current.timedComposition)
	ctx.Step("^die gemessene Antwort ihre technische Abbruchgrenze einhält$", current.fragmentBudget)
}

func (current *world) waitSearch(ctx context.Context) error {
	raw := current.searchDelayString()
	if raw == "" {
		return nil
	}

	milliseconds, err := strconv.Atoi(raw)
	if err != nil {
		return err
	}

	timer := time.NewTimer(time.Duration(milliseconds) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (current *world) timedFragment(table *godog.Table) error {
	started := time.Now()
	err := current.request(table)
	current.responseElapsed = time.Since(started)

	return err
}

func (current *world) timedComposition(table *godog.Table) error {
	started := time.Now()
	err := current.composedRequest(table)
	current.responseElapsed = time.Since(started)

	return err
}

func (current *world) fragmentBudget(table *godog.Table) error {
	maximum, err := strconv.Atoi(read(table)[0]["MaximumMilliseconds"])
	if err != nil {
		return err
	}

	if current.responseElapsed > time.Duration(maximum)*time.Millisecond {
		return fmt.Errorf("Fragmentantwort nach %s überschreitet %dms", current.responseElapsed, maximum)
	}

	return nil
}

func (current *world) searchDelayString() string {
	raw, ok := current.searchDelay.Load().(string)
	if !ok {
		return ""
	}

	return raw
}
