package demo_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (current *world) selectClient(table *godog.Table) error {
	return current.selectName(read(table)[0]["Jar"])
}

func (current *world) selectName(name string) error {
	if existing, exists := current.clients[name]; exists {
		current.client = existing
		return nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	current.client = &http.Client{Jar: jar, Timeout: 15 * time.Second, CheckRedirect: noRedirect}
	current.clients[name] = current.client

	return nil
}

func noRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (current *world) request(table *godog.Table) error {
	data := read(table)[0]
	if data["Boundary"] == "public" {
		return current.getAt(current.deployment.public, data["Path"], data["HX"])
	}

	return current.get(data["Path"], data["HX"])
}

func (current *world) get(path, hx string) error {
	path = strings.ReplaceAll(path, "{orderID}", string(current.order.Purchase.ID))
	target := current.deployment.public
	if strings.HasPrefix(path, "/backend") {
		target = current.deployment.backend
	}

	return current.getAt(target, path, hx)
}

func (current *world) getAt(target, path, hx string) error {
	request, err := http.NewRequest(http.MethodGet, target+path, nil)
	if err != nil {
		return err
	}

	request.Header.Set("HX-Request", hx)

	return current.send(request)
}

func (current *world) send(request *http.Request) error {
	requested := time.Now()
	response, err := current.client.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()
	body, err := io.ReadAll(response.Body)
	current.response = acceptResponse(response, body)
	current.response.requested = requested
	current.forms = parseForms(current.response.body)

	return err
}

func acceptResponse(result *http.Response, body []byte) response {
	return response{status: result.StatusCode, header: result.Header.Clone(), body: string(body)}
}

func (current *world) observe(table *godog.Table) error {
	for _, data := range read(table) {
		if err := current.observeRow(data); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) observeRow(data map[string]string) error {
	if expected := data["Status"]; expected != "" {
		if strconv.Itoa(current.response.status) != expected {
			return fmt.Errorf("HTTP %d statt %s: %s", current.response.status, expected, current.response.diagnostic())
		}
	}

	if expected := data["Contains"]; expected != "" {
		if !strings.Contains(html.UnescapeString(current.response.body), expected) {
			return fmt.Errorf("HTML enthält %q nicht: %s", expected, current.response.diagnostic())
		}
	}

	return current.observeMetadata(data)
}

func (current *world) observeMetadata(data map[string]string) error {
	if excluded := data["Excludes"]; excluded != "" {
		if strings.Contains(html.UnescapeString(current.response.body), excluded) {
			return fmt.Errorf("HTML enthält unerlaubtes %q", excluded)
		}
	}

	if header := data["Header"]; header != "" {
		actual := current.response.header.Get(header)
		if !headerMatches(actual, data) {
			return fmt.Errorf("Header %s = %q statt %q", header, actual, data["Value"])
		}
	}

	return nil
}

func headerMatches(actual string, data map[string]string) bool {
	if prefix := data["Prefix"]; prefix != "" {
		return strings.HasPrefix(actual, prefix)
	}

	return actual == data["Value"]
}

func (current *world) await(table *godog.Table) error {
	data := read(table)[0]
	seconds, err := strconv.Atoi(data["TimeoutSeconds"])
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	for time.Now().Before(deadline) {
		if err := current.get(data["Path"], data["HX"]); err != nil {
			return err
		}

		if current.observe(table) == nil {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return current.awaitFailure(table)
}

func (current *world) awaitFailure(table *godog.Table) error {
	err := current.observe(table)
	if current.deployment.app == nil {
		return err
	}

	diagnostics, diagnosticError := current.deployment.app.Snapshot(context.Background())
	requestError := current.get(read(table)[0]["Path"], "true")

	return fmt.Errorf("%w; Diagnose %v %w; HXprobe %w: %s", err, diagnostics, diagnosticError, requestError, current.response.diagnostic())
}

func cloneValues(values url.Values) url.Values {
	result := url.Values{}
	for key, entries := range values {
		result[key] = append([]string{}, entries...)
	}

	return result
}
