package opensearch_test

import (
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (current *world) msearchSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein beobachtbarer Suchprotokollorigin vorbereitet ist$`, current.prepareMsearch)
	ctx.Step(`^die Suche über den Protokollorigin ausgeführt wird$`, current.protocolSearch)
	ctx.Step(`^der Suchdrahtvertrag entspricht$`, current.protocolObserved)
	ctx.Step(`^das vollständige Protokollsuchergebnis entspricht$`, current.protocolResult)
	ctx.Step(`^die NDJSON-Teilabfragen entsprechen$`, current.protocolParts)
	ctx.Step(`^die ursprüngliche Transportdiagnose ausschließlich sichere Felder enthält$`, current.transportDiagnostic)
}

func (current *world) prepareMsearch(table *godog.Table) error {
	data := read(table)[0]
	if data["CaptureDiagnostics"] == "true" {
		current.startTransportLog()
	}

	probe := &msearchProbe{status: 200}
	if err := probe.configure(data); err != nil {
		return err
	}

	current.msearch = probe
	probe.server = httptest.NewServer(http.HandlerFunc(probe.serve))
	current.config = opensearch.Config{Endpoint: probe.server.URL + data["EndpointSuffix"], Index: "protocol", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
	current.adapter = opensearch.New(current.config)

	return nil
}

func (probe *msearchProbe) configure(data row) error {
	probe.response = []byte(data["Body"])
	if data["ResponseFixture"] != "" {
		payload, err := os.ReadFile("../../" + data["ResponseFixture"])
		if err != nil {
			return err
		}

		probe.response = payload
	}

	if data["Status"] != "" {
		status, err := strconv.Atoi(data["Status"])
		if err != nil {
			return err
		}

		probe.status = status
	}

	return probe.configureDelay(data)
}

func (probe *msearchProbe) serve(writer http.ResponseWriter, request *http.Request) {
	payload, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}

	probe.mutex.Lock()
	probe.request = payload
	probe.path = request.URL.RequestURI()
	probe.mediaType = request.Header.Get("Content-Type")
	probe.calls++
	probe.mutex.Unlock()
	probe.respond(writer, request)
}

func (probe *msearchProbe) respond(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	if !probe.wait(request) {
		return
	}

	writer.WriteHeader(probe.status)
	if probe.phase == "body" {
		if !probe.sendFirstPart(writer) {
			return
		}

		if !probe.pause(request) {
			return
		}
	}

	_, _ = writer.Write(probe.response)
}

func (current *world) protocolSearch(table *godog.Table) error {
	ctx, cancel, err := protocolParent(read(table)[0])
	if err != nil {
		return err
	}

	defer cancel()
	if read(table)[0]["CloseOrigin"] == "true" {
		current.msearch.server.Close()
	}

	current.query = application.SearchQuery{Text: read(table)[0]["Text"], Page: 1, Sort: "relevance", Context: current.config.Context}
	start := time.Now()
	current.result, current.err = current.adapter.Search(ctx, current.query)
	current.elapsed = time.Since(start)

	return nil
}

func (current *world) protocolObserved(table *godog.Table) error {
	probe := current.msearch
	probe.mutex.Lock()
	defer probe.mutex.Unlock()
	actual := row{"Path": probe.path, "ContentType": probe.mediaType, "Calls": fmt.Sprint(probe.calls), "FirstPartSent": fmt.Sprint(probe.firstPartSent), "FinalNewline": fmt.Sprint(strings.HasSuffix(string(probe.request), "\n")), "Lines": fmt.Sprint(len(strings.Split(strings.TrimSuffix(string(probe.request), "\n"), "\n")))}

	return read(table)[0].poolMatches(actual)
}

func (current *world) protocolResult(table *godog.Table) error {
	data := read(table)[0]
	if code := value.Code(current.err); code != data["Code"] {
		return fmt.Errorf("Suchfehler %s statt %s: %w", code, data["Code"], current.err)
	}

	expected := application.SearchResult{}
	if data["ExpectedFixture"] != "" {
		payload, err := os.ReadFile("../../" + data["ExpectedFixture"])
		if err != nil {
			return err
		}

		if err := json.Unmarshal(payload, &expected); err != nil {
			return err
		}
	}

	return sameProtocolResult(current.result, expected)
}

func sameProtocolResult(actual, expected application.SearchResult) error {
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("vollständige Suchantwort weicht ab: %+v statt %+v", actual, expected)
	}

	return nil
}

func (probe *msearchProbe) configureDelay(data row) error {
	if data["DelayMilliseconds"] == "" {
		return nil
	}

	milliseconds, err := strconv.Atoi(data["DelayMilliseconds"])
	probe.delay = time.Duration(milliseconds) * time.Millisecond
	probe.phase = data["Phase"]

	if err != nil {
		return err
	}

	return probe.configureStream()
}

func (probe *msearchProbe) wait(request *http.Request) bool {
	if probe.phase == "headers" {
		return probe.pause(request)
	}

	return true
}

func (probe *msearchProbe) pause(request *http.Request) bool {
	timer := time.NewTimer(probe.delay)
	defer timer.Stop()
	select {
	case <-request.Context().Done():
		return false
	case <-timer.C:
		return true
	}
}

func (current *world) protocolParts(table *godog.Table) error {
	current.msearch.mutex.Lock()
	payload := append([]byte{}, current.msearch.request...)
	current.msearch.mutex.Unlock()
	lines := strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n")
	for _, expected := range read(table) {
		if err := expected.protocolPart(lines); err != nil {
			return err
		}
	}

	return nil
}

func (expected row) protocolPart(lines []string) error {
	index, err := strconv.Atoi(expected["Line"])
	if err != nil {
		return err
	}

	if !validProtocolLine(index, lines) {
		return fmt.Errorf("NDJSON-Zeile fehlt")
	}

	var actual any
	if err := json.Unmarshal([]byte(lines[index]), &actual); err != nil {
		return err
	}

	return expected.protocolPath(actual)
}

func (expected row) protocolPath(actual any) error {
	for _, key := range strings.Split(expected["Path"], ".") {
		object, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("JSON-Objekt für %s fehlt", key)
		}

		actual = object[key]
	}

	return expected.protocolValue(actual)
}

func (expected row) protocolValue(actual any) error {
	var wanted any
	if err := json.Unmarshal([]byte(expected["JSON"]), &wanted); err != nil {
		return err
	}

	if !reflect.DeepEqual(actual, wanted) {
		return fmt.Errorf("NDJSON %s: %+v statt %+v", expected["Path"], actual, wanted)
	}

	return nil
}

func validProtocolLine(index int, lines []string) bool {
	return index >= 0 && index < len(lines)
}

func (probe *msearchProbe) configureStream() error {
	if probe.phase != "body" {
		return nil
	}

	var envelope protocolEnvelope
	if err := json.Unmarshal(probe.response, &envelope); err != nil {
		return err
	}

	if len(envelope.Responses) != 2 {
		return fmt.Errorf("Zeitfixture benötigt zwei Teilantworten")
	}

	probe.prefix = append([]byte(`{"responses":[`), envelope.Responses[0]...)
	probe.prefix = append(probe.prefix, ',')
	probe.response = append([]byte{}, envelope.Responses[1]...)
	probe.response = append(probe.response, ']', '}')

	return nil
}

func (probe *msearchProbe) sendFirstPart(writer http.ResponseWriter) bool {
	if _, err := writer.Write(probe.prefix); err != nil {
		return false
	}

	if err := http.NewResponseController(writer).Flush(); err != nil {
		return false
	}

	probe.mutex.Lock()
	probe.firstPartSent = true
	probe.mutex.Unlock()

	return true
}
