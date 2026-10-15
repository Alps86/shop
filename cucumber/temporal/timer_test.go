package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/testsuite"
	"strconv"
	"time"
)

func (current *world) registerTimer(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein kontrollierter SDK-Zeitlauf vorbereitet ist$`, current.timerSetup)
	ctx.Step(`^der Zeitlauf ohne reale Wartezeit ausgeführt wird$`, current.timerRun)
	ctx.Step(`^die Ablaufactivity den letzten Datenbankstand respektiert$`, current.timerObserved)
}

func (current *world) timerSetup(table *godog.Table) error {
	row := table.Rows[1].Cells
	hours, err := strconv.Atoi(row[1].Value)
	if err != nil {
		return err
	}

	current.testEnvironment = (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	current.started = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	current.testEnvironment.SetStartTime(current.started)
	current.store = &cartStore{cart: sale.Cart{ID: "c-timer", UserID: "u-a", State: sale.CartState(row[0].Value), LastActivityAt: current.started}}
	current.adapter = temporal.New(nil, application.Service{}, current.store)
	current.testEnvironment.RegisterActivity(current.adapter.Expire)
	current.scheduleActivity(hours)

	return nil
}

func (current *world) timerRun() {
	current.testEnvironment.ExecuteWorkflow(current.adapter.CartWorkflow, temporal.Start{ID: "c-timer", Binding: application.ProcessBinding{DefinitionVersion: "cart-v1", RuleSetVersion: "demo-v1"}})
	current.err = current.testEnvironment.GetWorkflowError()
}

func (current *world) timerObserved(table *godog.Table) error {
	row := table.Rows[1].Cells
	expected, err := strconv.Atoi(row[1].Value)
	if err != nil {
		return err
	}

	if current.err != nil {
		return current.err
	}

	elapsed := current.testEnvironment.Now().Sub(current.started) / time.Hour
	if elapsed != time.Duration(expected) {
		return fmt.Errorf("Ablauf nach %d statt %d Stunden", elapsed, expected)
	}

	if string(current.store.cart.State) != row[0].Value {
		return fmt.Errorf("Zustand %s statt %s", current.store.cart.State, row[0].Value)
	}

	return current.timerCheckpointObserved(table)
}

func (current *world) watchStaleTimer(hours int) {
	if hours == 0 {
		return
	}

	current.testEnvironment.RegisterDelayedCallback(func() {
		current.timerCheckpoint, current.queryErr = current.store.Get(context.Background(), "c-timer")
	}, 168*time.Hour+time.Second)
}

func (current *world) timerCheckpointObserved(table *godog.Table) error {
	row := current.rows(table)[0]
	if err := current.activityCountObserved(row); err != nil {
		return err
	}

	if row["CheckpointState"] == "" {
		return nil
	}

	deadline := current.timerCheckpoint.LastActivityAt.Add(168*time.Hour).Sub(current.started) / time.Hour
	if strconv.FormatInt(int64(deadline), 10) != row["NextDeadlineHours"] {
		return fmt.Errorf("erneuerte Frist %d statt %s", deadline, row["NextDeadlineHours"])
	}

	if string(current.timerCheckpoint.State) != row["CheckpointState"] {
		return fmt.Errorf("alter Timer überschrieb neue Aktivität")
	}

	return current.queryErr
}

func (current *world) scheduleActivity(hours int) {
	current.testEnvironment.RegisterDelayedCallback(func() {
		current.store.mutex.Lock()
		defer current.store.mutex.Unlock()
		current.store.cart.LastActivityAt = current.started.Add(time.Duration(hours) * time.Hour)
	}, time.Duration(hours)*time.Hour)

	current.watchStaleTimer(hours)
}

func (current *world) activityCountObserved(row map[string]string) error {
	if row["ActivityCalls"] == "" {
		return nil
	}

	if strconv.Itoa(current.store.expirationCalls) != row["ActivityCalls"] {
		return fmt.Errorf("Activityversuche %d statt %s", current.store.expirationCalls, row["ActivityCalls"])
	}

	return nil
}
