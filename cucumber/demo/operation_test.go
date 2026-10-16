package demo_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"os/exec"
	"strconv"
	"time"
)

func (current *world) registerOperations(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Kaufdaten und Workflowausführungen werden festgehalten$`, current.rememberOrder)
	ctx.Step(`^die eigene Demolaufzeit wird persistent neu gestartet$`, current.restart)
	ctx.Step(`^Kaufdaten und Workflowausführungen sind unverändert$`, current.retained)
	ctx.Step(`^die echte Suchverbindung wird umgeschaltet$`, current.searchState)
	ctx.Step(`^eigene HTTPgastkäufe sind bis einschließlich Preisbestätigung vorbereitet$`, current.preparePurchases)
	ctx.Step(`^die vorbereiteten Abschlussformulare werden gleichzeitig gesendet$`, current.concurrentPurchases)
	ctx.Step(`^die echten konkurrierenden Käufe ergeben$`, current.concurrentObserved)
	ctx.Step(`^ein vorbereitetes Abschlussformular wird gesendet$`, current.submitPrepared)
	ctx.Step(`^der öffentliche Produktcache wird tatsächlich wiederverwendet$`, current.cacheHit)
}

func (current *world) rememberOrder(table *godog.Table) error {
	data, err := json.Marshal(current.order.Purchase)
	if err != nil {
		return err
	}

	current.snapshot = string(data)
	seconds, err := strconv.Atoi(read(table)[0]["TimeoutSeconds"])
	if err != nil {
		return err
	}

	for _, id := range []string{"cart/" + current.cartID, "order/" + string(current.order.Purchase.ID)} {
		if err := current.rememberRun(id, time.Duration(seconds)*time.Second); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) rememberRun(id string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		response, err := current.deployment.connection.DescribeWorkflowExecution(context.Background(), id, "")
		if err == nil {
			current.runs[id] = response.WorkflowExecutionInfo.Execution.RunId
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Workflow %s nicht innerhalb %s gestartet", id, timeout)
}

func (current *world) restart() error {
	deployment := current.deployment
	if deployment.external {
		return deployment.restartCompose()
	}

	if err := errors.Join(deployment.stopApp(), deployment.app.Close(context.Background())); err != nil {
		return err
	}

	app, err := runtime.Open(context.Background(), deployment.config)
	if err != nil {
		return err
	}

	deployment.app, deployment.store, deployment.connection = app, app.Store, app.Client

	return deployment.startApp()
}

func (deployment *deployment) restartCompose() error {
	commands := []*exec.Cmd{exec.Command("docker", "compose", "stop", "shop"), exec.Command("docker", "compose", "restart", "postgres", "temporal", "opensearch"), exec.Command("docker", "compose", "up", "--detach", "--wait", "--wait-timeout", "180")}
	for _, command := range commands {
		command.Dir = "../.."
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Compose-Dienstrestart: %w: %s", err, output)
		}
	}

	return deployment.ready(30 * time.Second)
}

func (current *world) retained() error {
	order, err := current.service().GetOrder(context.Background(), current.session(), current.order.Purchase.ID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(order.Purchase)
	if err != nil {
		return err
	}

	if string(data) != current.snapshot {
		return fmt.Errorf("unveränderlicher Kauf änderte sich")
	}

	return current.sameRuns()
}

func (current *world) sameRuns() error {
	for id, run := range current.runs {
		response, err := current.deployment.connection.DescribeWorkflowExecution(context.Background(), id, "")
		if err != nil {
			return err
		}

		if response.WorkflowExecutionInfo.Execution.RunId != run {
			return fmt.Errorf("neue Workflowausführung für %s", id)
		}
	}

	return current.get(current.orderPath, "")
}
