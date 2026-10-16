package opensearch_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
)

func (current *world) cleanup(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
	if current.pool != nil {
		current.pool.close()
		return ctx, nil
	}

	if current.msearch != nil {
		current.msearch.server.Close()
		return ctx, nil
	}

	return current.cleanupIndex(ctx, scenario, err)
}

func (current *world) cleanupIndex(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.faultServer != nil {
		defer current.faultServer.Close()
		current.config = current.cleanupConfig
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, current.config.Endpoint+"/"+current.config.Index+"*", nil)
	if err != nil {
		return ctx, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return ctx, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return ctx, errors.Join(fmt.Errorf("Indexbereinigung: %s", response.Status), response.Body.Close())
	}

	return ctx, response.Body.Close()
}
