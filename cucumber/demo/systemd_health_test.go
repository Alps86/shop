package demo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/client"
)

func (deployment *deployment) waitSystemd(ctx context.Context, dependencies bool) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if deployment.systemdHealthy(ctx, dependencies) {
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("Referenzdienste nicht bereit: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (deployment *deployment) systemdHealthy(ctx context.Context, dependencies bool) bool {
	if dependencies {
		if _, err := deployment.store.Diagnostics(ctx); err != nil {
			return false
		}

		if _, err := deployment.connection.CheckHealth(ctx, &client.CheckHealthRequest{}); err != nil {
			return false
		}
	}

	return deployment.systemdSearchHealthy(ctx)
}

func (deployment *deployment) systemdSearchHealthy(ctx context.Context) bool {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, deployment.config.SearchEndpoint+"/_cluster/health?wait_for_status=yellow&timeout=1s", nil)
	if err != nil {
		return false
	}

	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	if err != nil {
		return false
	}

	healthy := systemdSearchResponse(response)
	if err := response.Body.Close(); err != nil {
		return false
	}

	return healthy
}

func systemdSearchResponse(response *http.Response) bool {
	if response.StatusCode != http.StatusOK {
		return false
	}

	var status systemdSearchHealth
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		return false
	}

	if status.TimedOut {
		return false
	}

	return map[string]bool{"green": true, "yellow": true}[status.Status]
}
