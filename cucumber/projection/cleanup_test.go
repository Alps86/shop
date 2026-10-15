package projection_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"net/http"
)

func (current *world) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	rebuildErr := current.abortRebuild()
	stopErr := current.stopConsumer()
	indexErr := current.removeIndex(ctx, current.Alias+"-v1")
	if current.TargetIndex != "" {
		indexErr = errors.Join(indexErr, current.removeIndex(ctx, current.TargetIndex))
	}

	if current.Store != nil {
		current.Store.Close()
	}

	return ctx, errors.Join(rebuildErr, stopErr, indexErr, current.removeDatabase(ctx))
}

func (current *world) removeDatabase(ctx context.Context) error {
	if current.Admin == nil {
		return nil
	}

	_, slotErr := current.Admin.Exec(ctx, "SELECT pg_drop_replication_slot(slot_name) FROM pg_replication_slots WHERE slot_name=$1", current.Config.Slot)
	_, pubErr := current.Admin.Exec(ctx, "DROP PUBLICATION IF EXISTS "+pgx.Identifier{current.Config.Publication}.Sanitize())
	_, schemaErr := current.Admin.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{current.Config.Schema}.Sanitize()+" CASCADE")

	return errors.Join(slotErr, pubErr, schemaErr, current.Admin.Close(ctx))
}

func (current *world) removeIndex(ctx context.Context, index string) error {
	if current.Endpoint == "" {
		return nil
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, current.Endpoint+"/"+index, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusInternalServerError {
		return errors.Join(fmt.Errorf("Indexbereinigung %s", response.Status), response.Body.Close())
	}

	return response.Body.Close()
}
