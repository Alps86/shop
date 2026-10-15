package temporal_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"strconv"
)

func (current *world) registerRegistration(ctx *godog.ScenarioContext) {
	ctx.Step(`^das SDK eine Handlerregistrierung ablehnt$`, current.registrationFailure)
	ctx.Step(`^der Workflow ohne Activitywirkung fehlschlägt$`, current.registrationObserved)
}

func (hook *registrationInterceptor) InterceptWorkflow(_ workflow.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	return &registrationInbound{WorkflowInboundInterceptorBase: interceptor.WorkflowInboundInterceptorBase{Next: next}, Operation: hook.Operation}
}

func (hook *registrationInbound) Init(next interceptor.WorkflowOutboundInterceptor) error {
	return hook.Next.Init(&registrationOutbound{WorkflowOutboundInterceptorBase: interceptor.WorkflowOutboundInterceptorBase{Next: next}, Operation: hook.Operation})
}

func (hook *registrationOutbound) SetUpdateHandler(ctx workflow.Context, name string, handler any, options workflow.UpdateHandlerOptions) error {
	if hook.Operation == "update" {
		return value.NewError("DependencyUnavailable", "SDKRegistration")
	}

	return hook.Next.SetUpdateHandler(ctx, name, handler, options)
}

func (hook *registrationOutbound) SetQueryHandler(ctx workflow.Context, name string, handler any) error {
	if hook.Operation == "query" {
		return value.NewError("DependencyUnavailable", "SDKRegistration")
	}

	return hook.Next.SetQueryHandler(ctx, name, handler)
}

func (current *world) registrationFailure(table *godog.Table) error {
	row := current.rows(table)[0]
	current.testEnvironment.SetWorkerOptions(worker.Options{Interceptors: []interceptor.WorkerInterceptor{&registrationInterceptor{Operation: row["Registration"]}}})
	workflows := map[string]any{"cart": current.adapter.CartWorkflow, "order": current.adapter.OrderWorkflow}
	current.testEnvironment.ExecuteWorkflow(workflows[row["Workflow"]], temporal.Start{ID: "c-timer", Binding: application.ProcessBinding{DefinitionVersion: "cart-v1", RuleSetVersion: "demo-v1"}})
	current.err = current.testEnvironment.GetWorkflowError()

	return nil
}

func (current *world) registrationObserved(table *godog.Table) error {
	actual := map[string]string{"Failed": strconv.FormatBool(current.err != nil), "ActivityCalls": strconv.Itoa(current.store.expirationCalls)}
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}
