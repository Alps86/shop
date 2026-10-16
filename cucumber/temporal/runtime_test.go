package temporal_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc/status"
	"time"
)

func (current *world) prepare(table *godog.Table) error {
	payload, err := current.readFixture(current.rows(table)[0]["Fixture"])
	if err != nil {
		return err
	}

	current.store = &cartStore{}
	if err := json.Unmarshal(payload, &current.store.cart); err != nil {
		return err
	}

	current.store.cart.ID = value.CartID(current.id)
	current.store.cart.LastActivityAt = time.Now()
	current.adapter.Service = application.Service{Carts: current.store, Clock: current.store, Writer: current.store}
	current.adapter.Expiration = current.store
	current.adapter.Queue = current.id
	current.runner = current.adapter.Worker()

	return current.runner.Start()
}

func (current *world) Get(_ context.Context, _ value.CartID) (sale.Cart, error) {
	return current.store.Get(context.Background(), value.CartID(current.id))
}

func (store *cartStore) Get(_ context.Context, _ value.CartID) (sale.Cart, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return store.cart, nil
}

func (store *cartStore) Now() time.Time {
	return time.Now()
}

func (store *cartStore) Save(ctx context.Context, write application.CartWrite) (application.CartView, error) {
	timer := time.NewTimer(store.writeDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return application.CartView{}, ctx.Err()
	case <-timer.C:
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.cart = write.Cart

	return application.NewCartView(store.cart), nil
}

func (store *cartStore) Expire(_ context.Context, _ value.CartID, now time.Time) (sale.Cart, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.expirationCalls++
	if store.expirationFailures > 0 {
		store.expirationFailures--
		return sale.Cart{}, store.expirationError
	}

	if store.cart.IsExpired(now) {
		store.cart.State = sale.Expired
		return store.cart, value.NewError("Expired", "Cart")
	}

	return store.cart, nil
}

func (current *world) restart(table *godog.Table) error {
	if err := current.resumeCart(); err != nil {
		return err
	}

	current.stopColdWorker()
	current.runner = current.adapter.Worker()
	if err := current.runner.Start(); err != nil {
		return err
	}

	if err := current.awaitCartWorker(table); err != nil {
		return err
	}

	return current.resumeCart()
}

func (current *world) stopColdWorker() {
	current.runner.Stop()
	// This serial fixture has no other live worker; a process restart loses this cache.
	worker.PurgeStickyWorkflowCache()
}

func (current *world) awaitCartWorker(table *godog.Table) error {
	timeout, err := time.ParseDuration(current.rows(table)[0]["ReadinessTimeout"])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// Start spawns pollers; only a successful query proves cold history replay.
	encoded, err := current.connection.QueryWorkflow(ctx, "cart/"+current.id, current.runID, "ProcessStatus")
	if err != nil {
		return fmt.Errorf("Restart.Readiness QueryWorkflow: type=%T grpc=%s canceled=%t caller=%s: %w", err, status.Code(err), errors.Is(err, context.Canceled), fmt.Sprint(ctx.Err()), err)
	}

	if err := encoded.Get(&current.processStatus); err != nil {
		return fmt.Errorf("Restart.Readiness Decode: %w", err)
	}

	return nil
}

func (current *world) resumeCart() error {
	user := value.UserID("u-a")
	if current.databaseStore != nil {
		user = current.guest.UserID
	}

	view, err := current.adapter.EvaluateCart(context.Background(), application.EvaluateCommand{UserID: user, CartID: value.CartID(current.id)})
	if err != nil {
		return err
	}

	current.view = view

	return current.sameRun()
}

func (current *world) sameRun() error {
	description, err := current.connection.DescribeWorkflowExecution(context.Background(), "cart/"+current.id, "")
	if err != nil {
		return err
	}

	if current.runID != "" {
		if description.WorkflowExecutionInfo.Execution.RunId != current.runID {
			return fmt.Errorf("Workflowausführung wechselte nach Neustart")
		}
	}

	current.runID = description.WorkflowExecutionInfo.Execution.RunId

	return nil
}

func (current *world) update(table *godog.Table) error {
	row := table.Rows[1].Cells
	command := application.ChangeCommand{UserID: value.UserID(row[0].Value), CartID: value.CartID(current.id), CommandID: value.CommandID(row[1].Value), ExpectedRevision: current.store.cart.Revision, Operation: row[2].Value, ShippingMethod: sale.Standard}
	current.view, current.err = current.adapter.UpdateCart(context.Background(), command)

	return nil
}

func (current *world) observed(table *godog.Table) error {
	row := table.Rows[1].Cells
	if value.Code(current.err) != row[0].Value {
		return fmt.Errorf("Code %s statt %s: %w", value.Code(current.err), row[0].Value, current.err)
	}

	encoding, err := current.connection.QueryWorkflow(context.Background(), "cart/"+current.id, "", "ProcessStatus")
	if err != nil {
		return err
	}

	var status temporal.Status
	if err := encoding.Get(&status); err != nil {
		return err
	}

	return current.statusObserved(table, status)
}

func (current *world) statusObserved(table *godog.Table, status temporal.Status) error {
	actual := map[string]string{"Code": value.Code(current.err), "State": status.State, "RuleSetVersion": status.Binding.RuleSetVersion}
	for key, wanted := range current.rows(table)[0] {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}

func (current *world) registerRuntime(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Worker den gespeicherten Cart übernimmt$`, current.prepare)
	ctx.Step(`^der Worker mit geleertem lokalen SDK-Cache neu startet und seine History wieder bereitsteht$`, current.restart)
	ctx.Step(`^ein echtes Workflowupdate gesendet wird$`, current.update)
	ctx.Step(`^antwortet der weitergeführte Prozess$`, current.observed)
}
