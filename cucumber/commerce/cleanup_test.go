package commerce_test

import (
	"context"
	"errors"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"go.temporal.io/api/serviceerror"
	"net/http"
)

func (current *world) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.app == nil {
		return ctx, nil
	}

	runErr := current.stop()
	workflowErr := errors.Join(current.terminate("cart/"+current.cartID), current.terminate("order/"+string(current.purchase.OrderID)))

	return ctx, errors.Join(runErr, workflowErr, current.app.Close(ctx), current.release(ctx))
}

func (current *world) release(ctx context.Context) error {
	connection, err := pgx.Connect(ctx, current.config.DSN)
	if err != nil {
		return err
	}

	_, slotErr := connection.Exec(ctx, "SELECT pg_drop_replication_slot(slot_name) FROM pg_replication_slots WHERE slot_name=$1", current.config.Slot)
	_, publicationErr := connection.Exec(ctx, "DROP PUBLICATION IF EXISTS "+pgx.Identifier{current.config.Publication}.Sanitize())
	_, schemaErr := connection.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{current.config.Schema}.Sanitize()+" CASCADE")

	return errors.Join(slotErr, publicationErr, schemaErr, connection.Close(ctx), current.removeIndex(ctx))
}

func (current *world) removeIndex(ctx context.Context) error {
	if current.config.SearchEndpoint == "http://127.0.0.1:1" {
		return nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, current.config.SearchEndpoint+"/"+current.config.SearchIndex+"*", nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	return response.Body.Close()
}

func (current *world) stop() error {
	if current.cancel == nil {
		return nil
	}

	current.cancel()
	err := <-current.finished
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func (current *world) terminate(id string) error {
	if id == "cart/" {
		return nil
	}

	if id == "order/" {
		return nil
	}

	err := current.app.Client.TerminateWorkflow(context.Background(), id, "", "Commercevertrag beendet")
	var missing *serviceerror.NotFound
	if errors.As(err, &missing) {
		return nil
	}

	return err
}
