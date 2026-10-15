package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/cucumber/godog"
	enums "go.temporal.io/api/enums/v1"
	history "go.temporal.io/api/history/v1"
	"strconv"
)

func (current *world) registerBudget(ctx *godog.ScenarioContext) {
	ctx.Step(`^die echte History die begrenzten Activitybudgets enthält$`, current.budgetObserved)
	ctx.Step(`^wiederholtes EnsureOrder dieselbe Ausführung erhält$`, current.ensureOrderRepeated)
}

func (current *world) budgetObserved(table *godog.Table) error {
	events := current.connection.GetWorkflowHistory(context.Background(), "cart/"+current.id, "", false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for events.HasNext() {
		event, err := events.Next()
		if err != nil {
			return err
		}

		if activity := event.GetActivityTaskScheduledEventAttributes(); activity != nil {
			return current.activityBudget(table, activity)
		}
	}

	return fmt.Errorf("History enthält keine tatsächlich geplante Activity")
}

func (current *world) activityBudget(table *godog.Table, activity *history.ActivityTaskScheduledEventAttributes) error {
	retry := activity.GetRetryPolicy()
	actual := map[string]string{"StartToCloseSeconds": fmt.Sprint(activity.GetStartToCloseTimeout().AsDuration().Seconds()), "ScheduleToCloseSeconds": fmt.Sprint(activity.GetScheduleToCloseTimeout().AsDuration().Seconds()), "MaximumAttempts": fmt.Sprint(retry.GetMaximumAttempts()), "InitialRetrySeconds": fmt.Sprint(retry.GetInitialInterval().AsDuration().Seconds()), "MaximumRetrySeconds": fmt.Sprint(retry.GetMaximumInterval().AsDuration().Seconds())}
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}

func (current *world) ensureOrderRepeated(table *godog.Table) error {
	attempts, err := strconv.Atoi(current.rows(table)[0]["Attempts"])
	if err != nil {
		return err
	}

	before, err := current.orderRunID()
	if err != nil {
		return err
	}

	binding := application.ProcessBinding{DefinitionVersion: current.initialPurchase.DefinitionVersion, RuleSetVersion: current.initialPurchase.RuleSetVersion}
	for range attempts {
		if err := current.adapter.EnsureOrder(context.Background(), current.purchase.OrderID, binding); err != nil {
			return err
		}
	}

	return current.sameOrderRun(before)
}

func (current *world) orderRunID() (string, error) {
	description, err := current.connection.DescribeWorkflowExecution(context.Background(), "order/"+string(current.purchase.OrderID), "")
	if err != nil {
		return "", err
	}

	return description.WorkflowExecutionInfo.Execution.RunId, nil
}

func (current *world) sameOrderRun(before string) error {
	after, err := current.orderRunID()
	if err != nil {
		return err
	}

	if before != after {
		return fmt.Errorf("EnsureOrder wechselte RunID")
	}

	return nil
}
