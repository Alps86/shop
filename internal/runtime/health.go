package runtime

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/client"
	"net/http"
	"strings"
	"time"
)

func (app *App) healthHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.health)
	mux.HandleFunc("GET /ready", app.ready)

	return mux
}

func (app *App) health(writer http.ResponseWriter, request *http.Request) {
	status, err := app.Store.Diagnostics(request.Context())
	if err != nil {
		http.Error(writer, value.Code(err), http.StatusServiceUnavailable)
		return
	}

	http.ServeContent(writer, request, "", time.Time{}, strings.NewReader(status["PostgreSQL"]))
}

func (app *App) ready(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), time.Second)
	defer cancel()
	status, err := app.Snapshot(ctx)
	if err != nil {
		http.Error(writer, value.Code(err), http.StatusServiceUnavailable)
		return
	}

	if !app.allReady(status) {
		http.Error(writer, fmt.Sprint(status), http.StatusServiceUnavailable)
		return
	}

	http.ServeContent(writer, request, "", time.Time{}, strings.NewReader(fmt.Sprint(status)))
}

func (app *App) Snapshot(ctx context.Context) (map[string]string, error) {
	status, err := app.Store.Diagnostics(ctx)
	if err != nil {
		return nil, err
	}

	_, temporalError := app.Client.CheckHealth(ctx, &client.CheckHealthRequest{})
	status["Temporal"] = app.dependencyStatus(temporalError)
	_, searchError := app.Search.Search(ctx, application.SearchQuery{Page: 1, Sort: "relevance", Context: app.Config.Context})
	status["OpenSearch"] = app.dependencyStatus(searchError)
	app.processStatus(status)

	return status, nil
}

func (app *App) dependencyStatus(err error) string {
	if err != nil {
		return value.Code(err)
	}

	return "ready"
}

func (app *App) allReady(status map[string]string) bool {
	return status["PostgreSQL"] == "ready" && status["Temporal"] == "ready" && status["ProcessOutbox"] == "ready"
}

func (app *App) processStatus(status map[string]string) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	status["CDC"] = app.projectionStatus
	if status["CDC"] == "" {
		status["CDC"] = "notStarted"
	}

	status["ProcessOutbox"] = "notStarted"
	if app.dispatchStarted {
		status["ProcessOutbox"] = app.dependencyStatus(app.dispatchError)
	}
}
