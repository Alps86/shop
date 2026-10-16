package http_test

import (
	"fmt"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/application"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
)

func (current *world) config() shophttp.Config {
	config := shophttp.Config{Routes: current.fixture.Routes, Diagnostics: diagnosticDouble{current}, Preview: guestDouble{current}, Variants: catalogDouble{current}, Interests: guestDouble{current}, Guests: guestDouble{current}, Catalog: catalogDouble{current}, Search: searchDouble{current}, Recommendations: recommendationDouble{current}, AdminCatalog: catalogDouble{current}, AdminOrders: adminOrderDouble{current}, Inventory: inventoryDouble{current}, Processes: processDouble{current}, CSRFKey: []byte("deterministic-test-key-32-bytes-long")}
	return current.optionsFor(config)
}

func (current *world) request(table *godog.Table) error {
	input := read(table)[0]
	factory := shophttp.New
	if input["Listener"] == "backend" {
		factory = shophttp.NewBackend
	}

	handler, err := factory(current.config())
	if err != nil {
		return err
	}

	current.handler = handler

	return current.execute(input)
}

func (current *world) execute(input map[string]string) error {
	request, err := current.makeRequest(input)
	if err != nil {
		return err
	}

	current.response = httptest.NewRecorder()
	current.handler.ServeHTTP(current.response, request)
	current.lastCSRF = csrfFromHTML(current.response.Body.String())

	return nil
}

func (current *world) addCookies(request *http.Request, input map[string]string) {
	if input["Session"] != "" {
		request.AddCookie(&http.Cookie{Name: "shop_session", Value: input["Session"], HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
	}

	for _, cookie := range current.cookies {
		request.AddCookie(cookie)
	}
}

func (current *world) getCSRF(input map[string]string, form url.Values) {
	request := httptest.NewRequest("GET", input["CSRFSource"], nil)
	current.addCookies(request, input)
	response := httptest.NewRecorder()
	current.handler.ServeHTTP(response, request)
	current.cookies = response.Result().Cookies()
	token := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindStringSubmatch(response.Body.String())
	if len(token) > 1 {
		form.Set("csrf", token[1])
	}
}

func (current *world) expect(table *godog.Table) error {
	for _, fields := range read(table) {
		if err := current.expectField(fields); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) expectField(fields map[string]string) error {
	actual := current.responseValue(fields["Field"])
	if fields["Field"] == "BodyAll" {
		return current.expectAll(actual, fields["Value"])
	}

	if fields["Field"] == "Status" {
		return current.expectStatus(fields["Value"])
	}

	if strings.HasPrefix(fields["Field"], "Absent") {
		return current.expectAbsent(actual, fields["Value"])
	}

	if !strings.Contains(actual, fields["Value"]) {
		return fmt.Errorf("%s enthält %q nicht: %s", fields["Field"], fields["Value"], actual)
	}

	return nil
}

func (current *world) expectPort(table *godog.Table) error {
	observed := current.snapshotSearch()
	actual := map[string]string{"Calls": strings.Join(observed.calls, ","), "Brands": strings.Join(observed.query.Filters.Brands, ","), "Page": fmt.Sprint(observed.query.Page), "SKU": string(current.lastChange.SKU), "Operation": current.lastChange.Operation, "Revision": fmt.Sprint(current.lastChange.ExpectedRevision), "CommandID": string(current.lastChange.CommandID), "CartID": string(current.lastChange.CartID)}
	current.commandValues(actual, observed.query)

	return read(table).matches(actual)
}

func (current *world) makeRequest(input map[string]string) (*http.Request, error) {
	form, err := current.requestForm(input)
	if err != nil {
		return nil, err
	}

	body := current.requestBody(input, form)
	target := input["URL"] + strings.Repeat(input["Append"], readCount(input["Count"]))
	request := httptest.NewRequest(input["Method"], target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", input["HTMX"])
	request.Header.Set("X-Shop-Composition", input["Composition"])
	current.addCookies(request, input)

	return request, nil
}

func (current *world) responseValue(field string) string {
	if field == "Status" {
		return fmt.Sprint(current.response.Code)
	}

	if strings.HasPrefix(field, "AbsentHeader:") {
		return strings.Join(current.response.Header().Values(strings.TrimPrefix(field, "AbsentHeader:")), "\n")
	}

	if strings.HasPrefix(field, "Header:") {
		return strings.Join(current.response.Header().Values(strings.TrimPrefix(field, "Header:")), "\n")
	}

	return current.response.Body.String()
}

func (current *world) expectAll(actual, values string) error {
	for _, value := range strings.Split(values, ",") {
		value = strings.ReplaceAll(value, "&#44;", ",")
		if !strings.Contains(actual, value) {
			return fmt.Errorf("Komponente %q fehlt: %s", value, actual)
		}
	}

	return nil
}

func (current *world) expectStatus(expected string) error {
	if fmt.Sprint(current.response.Code) != expected {
		return fmt.Errorf("Status %d statt %s", current.response.Code, expected)
	}

	return nil
}

func csrfFromHTML(body string) string {
	token := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindStringSubmatch(body)
	if len(token) > 1 {
		return token[1]
	}

	return ""
}

func (current *world) expectAbsent(actual, value string) error {
	if value == "" {
		if actual == "" {
			return nil
		}

		return fmt.Errorf("unerwünschter Header: %s", actual)
	}

	if strings.Contains(actual, value) {
		return fmt.Errorf("unerwünschter Inhalt: %s", value)
	}

	return nil
}

func (current *world) requestForm(input map[string]string) (url.Values, error) {
	form, err := url.ParseQuery(input["Form"])
	if err != nil {
		return nil, err
	}

	if input["CSRFSource"] != "" {
		current.getCSRF(input, form)
	}

	if input["CSRF"] == "previous" {
		form.Set("csrf", current.lastCSRF)
	}

	return form, nil
}

func (current *world) requestBody(input map[string]string, form url.Values) string {
	if raw, exists := input["RawBody"]; exists {
		return raw
	}

	return form.Encode()
}

func (current *world) commandValues(actual map[string]string, query application.SearchQuery) {
	actual["InventorySKU"] = string(current.lastInventory.SKU)
	actual["InventoryRevision"] = fmt.Sprint(current.lastInventory.ExpectedRevision)
	actual["InventoryCommandID"] = string(current.lastInventory.CommandID)
	actual["InventoryPrice"] = fmt.Sprint(current.lastInventory.Price.Value.MinorUnits)
	actual["InventoryStock"] = fmt.Sprint(current.lastInventory.Stock.Value)
	actual["DeliveryOrder"] = string(current.lastDelivery.OrderID)
	actual["DeliveryCommandID"] = string(current.lastDelivery.CommandID)
	actual["SearchInterests"] = fmt.Sprint(query.Interests)
	actual["RecommendationInterests"] = fmt.Sprint(current.lastRecommendation.Interests)
}
