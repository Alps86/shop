package opensearch_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

func (current *world) budgetSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^der echte Suchdienst über eine verzögerte Verbindung angesprochen wird$`, current.delayedConnection)
	ctx.Step(`^bleibt die öffentliche Suchantwort innerhalb des Fehlerbudgets$`, current.budgetObserved)
}

func (current *world) delayedConnection(table *godog.Table) error {
	milliseconds, err := strconv.Atoi(read(table)[0]["Milliseconds"])
	if err != nil {
		return err
	}

	endpoint, err := url.Parse(current.config.Endpoint)
	if err != nil {
		return err
	}

	current.openDelayed(endpoint, time.Duration(milliseconds)*time.Millisecond, read(table)[0]["Client"])

	return nil
}

func (current *world) openDelayed(endpoint *url.URL, delay time.Duration, client string) {
	current.cleanupConfig = current.config
	proxy := &delayedProxy{next: httputil.NewSingleHostReverseProxy(endpoint), delay: delay}
	current.faultServer = httptest.NewServer(proxy)
	current.config.Endpoint = current.faultServer.URL
	current.config.HTTP = nil
	if client == "custom" {
		current.config.HTTP = &http.Client{Timeout: 5 * time.Second}
	}

	current.adapter = opensearch.New(current.config)
}

func (proxy *delayedProxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	timer := time.NewTimer(proxy.delay)
	defer timer.Stop()
	select {
	case <-request.Context().Done():
		return
	case <-timer.C:
		proxy.next.ServeHTTP(writer, request)
	}
}

func (current *world) budgetObserved(table *godog.Table) error {
	maximum, err := strconv.Atoi(read(table)[0]["MaximumMilliseconds"])
	if err != nil {
		return err
	}

	if current.elapsed > time.Duration(maximum)*time.Millisecond {
		return fmt.Errorf("Suchantwort nach %s überschreitet %dms", current.elapsed, maximum)
	}

	return nil
}

func (current *world) measuredSearch() {
	started := time.Now()
	current.result, current.err = current.adapter.Search(context.Background(), current.query)
	current.elapsed = time.Since(started)
}
