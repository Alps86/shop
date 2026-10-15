package temporal_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerCommitRestart(ctx *godog.ScenarioContext) {
	ctx.Step(`^der Worker nach dem Kaufcommit vor der Activityantwort neu startet$`, current.commitRestart)
}

func (fault *purchaseFault) lostResponse(ctx context.Context) (application.PurchaseResult, error) {
	if fault.Committed != nil {
		close(fault.Committed)
		<-ctx.Done()
	}

	return application.PurchaseResult{}, value.NewError("DependencyUnavailable", "ResponseAfterCommit")
}

func (current *world) commitRestart(table *godog.Table) error {
	row := current.rows(table)[0]
	if err := current.confirmPurchase(row); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		current.purchase, current.err = current.service.Place(context.Background(), current.guest.Session, application.PlaceCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["PlaceCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID})
		close(done)
	}()
	<-current.commitFault.Committed
	current.runner.Stop()
	current.runner = current.adapter.Worker()
	if err := current.runner.Start(); err != nil {
		return err
	}

	<-done

	return current.awaitRestartedPurchase(row)
}

func (current *world) awaitRestartedPurchase(row map[string]string) error {
	if current.err != nil {
		return current.err
	}

	duration, err := time.ParseDuration(row["AwaitTimeout"])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	if err := current.connection.GetWorkflow(ctx, "cart/"+current.id, "").Get(ctx, nil); err != nil {
		return err
	}

	return current.restorePurchaseResult()
}

func (current *world) restorePurchaseResult() error {
	id, err := current.databaseStore.Orders().ForCart(context.Background(), value.CartID(current.id))
	if err != nil {
		return err
	}

	current.purchase.OrderID = id
	if err := current.sameRun(); err != nil {
		return err
	}

	return current.capturePurchase()
}
