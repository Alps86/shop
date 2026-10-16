package opensearch_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"os"
	"reflect"
	"strconv"
	"time"
)

func (current *world) poolSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein frischer Suchorigin mit synchronisierten Antworten bereitsteht$`, current.preparePool)
	ctx.Step(`^zwei getrennte parallele Suchbatches ausgeführt werden$`, current.poolBatches)
	ctx.Step(`^die vollständigen Suchantworten und Verbindungswiederverwendung entsprechen$`, current.poolObserved)
}

func (current *world) preparePool(table *godog.Table) error {
	data := read(table)[0]
	probe := &poolProbe{original: http.DefaultTransport}
	current.pool = probe
	if err := probe.fixtures(data); err != nil {
		return err
	}

	probe.server = httptest.NewServer(http.HandlerFunc(probe.serve))
	current.config = opensearch.Config{Endpoint: probe.server.URL, Index: "pool", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
	if err := probe.configure(&current.config, data["Client"]); err != nil {
		return err
	}

	current.adapter = opensearch.New(current.config)
	current.query = application.SearchQuery{Text: "ph", Page: 1, Sort: "relevance", Context: current.config.Context}

	return nil
}

func (probe *poolProbe) fixtures(data row) error {
	parallel, err := strconv.Atoi(data["Parallel"])
	if err != nil {
		return err
	}

	probe.parallel = parallel
	payload, err := os.ReadFile("../../" + data["ResponseFixture"])
	if err != nil {
		return err
	}

	probe.payload = payload
	expected, err := os.ReadFile("../../" + data["ExpectedFixture"])
	if err != nil {
		return err
	}

	return json.Unmarshal(expected, &probe.expected)
}

func (probe *poolProbe) configure(config *opensearch.Config, mode string) error {
	if mode == "default" {
		return nil
	}

	transport, err := newPoolTransport()
	if err != nil {
		return err
	}

	probe.transport = transport
	probe.transport.MaxIdleConnsPerHost = 1
	if mode == "explicit" {
		config.HTTP = &http.Client{Transport: probe.transport, Timeout: 5 * time.Second}
		return nil
	}

	probe.foreignDefault()

	return nil
}

func newPoolTransport() (*http.Transport, error) {
	standard, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("Fixture benötigt ursprünglichen Standardtransport")
	}

	return standard.Clone(), nil
}

func (probe *poolProbe) foreignDefault() {
	probe.transport.MaxIdleConnsPerHost = 2
	probe.sentinel = &poolSentinel{next: probe.transport}
	http.DefaultTransport = probe.sentinel
}

func (probe *poolProbe) serve(writer http.ResponseWriter, request *http.Request) {
	barrier := probe.barrier
	if barrier.arrived.Add(1) == int64(barrier.parallel) {
		close(barrier.release)
	}

	select {
	case <-request.Context().Done():
		return
	case <-barrier.release:
	}

	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(probe.payload)
}

func (current *world) poolBatches() {
	batch := poolBatch{probe: current.pool, adapter: current.adapter, query: current.query}
	batch.run()
	batch.run()
}

func (batch *poolBatch) run() {
	probe := batch.probe
	probe.barrier = &poolBarrier{parallel: probe.parallel, release: make(chan struct{})}
	probe.reused.Store(0)
	offset := len(probe.results)
	probe.results = append(probe.results, make([]application.SearchResult, probe.parallel)...)
	probe.failures = append(probe.failures, make([]error, probe.parallel)...)
	for index := range probe.parallel {
		batch.wait.Add(1)
		go batch.search(offset + index)
	}

	batch.wait.Wait()
}

func (batch *poolBatch) search(index int) {
	defer batch.wait.Done()
	trace := &httptrace.ClientTrace{GotConn: batch.probe.connection}
	ctx := httptrace.WithClientTrace(context.Background(), trace)
	batch.probe.results[index], batch.probe.failures[index] = batch.adapter.Search(ctx, batch.query)
}

func (probe *poolProbe) connection(info httptrace.GotConnInfo) {
	if info.Reused {
		probe.reused.Add(1)
	}
}

func (current *world) poolObserved(table *godog.Table) error {
	probe := current.pool
	for index, result := range probe.results {
		if err := probe.matches(index, result); err != nil {
			return err
		}
	}

	actual := row{"Results": fmt.Sprint(len(probe.results)), "Reused": fmt.Sprint(probe.reused.Load()), "SentinelCalls": "0"}
	if probe.sentinel != nil {
		actual["SentinelCalls"] = fmt.Sprint(probe.sentinel.calls.Load())
	}

	return read(table)[0].poolMatches(actual)
}

func (probe *poolProbe) matches(index int, result application.SearchResult) error {
	if probe.failures[index] != nil {
		return probe.failures[index]
	}

	if !reflect.DeepEqual(result, probe.expected) {
		return fmt.Errorf("Suchantwort %d weicht vollständig von der Erwartungsfixture ab: %+v", index, result)
	}

	return nil
}

func (expected row) poolMatches(actual row) error {
	for key, wanted := range expected {
		if actual[key] != wanted {
			return fmt.Errorf("%s: %s statt %s", key, actual[key], wanted)
		}
	}

	return nil
}

func (sentinel *poolSentinel) RoundTrip(request *http.Request) (*http.Response, error) {
	sentinel.calls.Add(1)

	return sentinel.next.RoundTrip(request)
}

func (probe *poolProbe) close() {
	http.DefaultTransport = probe.original
	if probe.server != nil {
		probe.server.Close()
	}

	if probe.transport != nil {
		probe.transport.CloseIdleConnections()
	}
}
