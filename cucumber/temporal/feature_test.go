package temporal_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @temporal", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "temporal", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Temporalverträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	current.registerComponent(ctx)
	current.registerPurchase(ctx)
	ctx.Before(current.before)
	ctx.Step(`^ein echter Temporalclient ist verbunden$`, current.connect)
	ctx.Step(`^der Cartprozess wird wiederholt gestartet$`, current.ensure)
	ctx.Step(`^genau dieselbe Workflowausführung besteht$`, current.exists)
	ctx.After(current.after)
}

func (current *world) connect() error {
	endpoint := os.Getenv("TEMPORAL_ADDRESS")
	if endpoint == "" {
		endpoint = "127.0.0.1:17233"
	}

	connection, err := client.Dial(client.Options{HostPort: endpoint})
	current.connection = connection
	current.adapter = temporal.New(connection, application.Service{}, nil)
	current.id = fmt.Sprintf("temporal-contract-%d", time.Now().UnixNano())

	return err
}

func (current *world) ensure(table *godog.Table) error {
	binding := application.ProcessBinding{DefinitionVersion: table.Rows[1].Cells[0].Value, RuleSetVersion: table.Rows[1].Cells[1].Value}
	attempts, err := strconv.Atoi(current.rows(table)[0]["Attempts"])
	if err != nil {
		return err
	}

	for range attempts {
		if err := current.adapter.EnsureCart(context.Background(), value.CartID(current.id), binding); err != nil {
			return err
		}

		if err := current.sameRun(); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) exists(table *godog.Table) error {
	response, err := current.connection.DescribeWorkflowExecution(context.Background(), "cart/"+current.id, "")
	if err != nil {
		return err
	}

	row := current.rows(table)[0]
	if strconv.FormatBool(response.WorkflowExecutionInfo.Execution.RunId == current.runID) != row["SameRun"] {
		return fmt.Errorf("Workflowausführung wechselte")
	}

	if response.ExecutionConfig.GetTaskQueue().GetName() != row["TaskQueue"] {
		return fmt.Errorf("Defaulttaskqueue änderte sich")
	}

	return current.connection.TerminateWorkflow(context.Background(), "cart/"+current.id, "", "Vertrag geprüft")
}

func (current *world) cleanup() error {
	if current.connection == nil {
		return nil
	}

	defer current.connection.Close()
	cartError := current.stopProcess("cart/" + current.id)
	if current.newCartID != "" {
		cartError = errors.Join(cartError, current.stopProcess("cart/"+string(current.newCartID)))
	}

	if current.purchase.OrderID == "" {
		return cartError
	}

	return errors.Join(cartError, current.stopProcess("order/"+string(current.purchase.OrderID)))
}

func (current *world) stopProcess(id string) error {
	err := current.connection.TerminateWorkflow(context.Background(), id, "", "Vertrag beendet")
	var missing *serviceerror.NotFound
	if errors.As(err, &missing) {
		return nil
	}

	return err
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	*current = world{}

	return ctx, nil
}

func (current *world) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.runner != nil {
		current.runner.Stop()
	}

	return ctx, errors.Join(current.cleanup(), current.cleanDatabase())
}

func (current *world) registerComponent(ctx *godog.ScenarioContext) {
	current.registerRuntime(ctx)
	current.registerOperation(ctx)
	current.registerTimer(ctx)
	current.registerSimulation(ctx)
	current.registerRetry(ctx)
	current.registerRecovery(ctx)
	current.registerRegistration(ctx)
	current.registerBoundary(ctx)
	current.registerOrderFailure(ctx)
	current.registerCartPending(ctx)
	current.registerBudget(ctx)
}

func (current *world) registerPurchase(ctx *godog.ScenarioContext) {
	current.registerDatabase(ctx)
	current.registerDelivery(ctx)
	current.registerConcurrency(ctx)
	current.registerPending(ctx)
	current.registerVersion(ctx)
	current.registerTerminal(ctx)
	current.registerActivityTime(ctx)
	current.registerCommitRestart(ctx)
}
