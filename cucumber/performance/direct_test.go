package performance_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) direct(table *godog.Table) (err error) {
	if err := current.initializeInternal(); err != nil {
		return err
	}

	defer func() {
		err = current.finishReport(err)
	}()

	for _, data := range read(table) {
		if err := current.loadProfile(data); err != nil {
			return err
		}

		if err := current.directRow(data); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) directRow(data map[string]string) error {
	count := 0
	for _, route := range strings.Split(data["Routes"], ",") {
		if err := current.directRoute(route, data); err != nil {
			current.diagnoseFailure(route, data["Boundary"])
			return fmt.Errorf("direkte Messung %s/%s: %w", data["Boundary"], route, err)
		}

		count++
	}

	if strconv.Itoa(count) != data["ExpectedMeasurements"] {
		return fmt.Errorf("direkte Messungen %d statt %s", count, data["ExpectedMeasurements"])
	}

	return nil
}

func (current *world) directRoute(route string, data map[string]string) error {
	path, cookie, contains := current.route(route)
	prefix := "backend-"
	if data["Boundary"] == "requestStart-to-last-Go-fragment-byte" {
		prefix = "fragment-"
		path = "/fragments" + path
		cookie = ""
		contains = current.fragmentContains(route)
	}

	representation, err := current.probeDirect(path, cookie, contains)
	if err != nil {
		return err
	}

	return current.measureDirect(route, prefix+route, path, directRequest{cookie: cookie, data: data, representation: representation})
}

func (current *world) measureDirect(route, name, path string, input directRequest) error {
	if _, err := current.hey(current.internal, path, input.cookie, input.data["WarmupPerRoute"]); err != nil {
		return err
	}

	samples, err := current.heyResult(current.internal, name, path, input.cookie)
	if err != nil {
		return err
	}

	report := summarize(samples)
	report.measurementProfile = current.profile
	report.Route = route
	report.Cache = input.data["Cache"]
	report.Representation = input.representation

	return current.recordDirect(report)
}

func (current *world) fragmentContains(route string) string {
	if route == "product" {
		return current.data["SKU"]
	}

	return current.data["ListingContains"]
}

func (current *world) recordDirect(report measurement) error {
	report.TargetP95Milliseconds = 0
	report.Goal = "not-applicable; diagnostic origin boundary, no full-page SLA"
	report.NewConnections = -1
	report.CacheApplicability = "direct Go; shared proxy cache not applicable"
	current.appendDirect(report)
	if report.Requests != current.expectedRequests {
		return fmt.Errorf("direkte Samples %d statt %d", report.Requests, current.expectedRequests)
	}

	if report.ErrorRate != 0 {
		return fmt.Errorf("direkte Fehlerquote %s: %f", report.Route, report.ErrorRate)
	}

	return nil
}

func (current *world) appendDirect(report measurement) {
	if report.Boundary == "requestStart-to-last-Go-fragment-byte" {
		current.fragmentReports = append(current.fragmentReports, report)
		return
	}

	current.backendReports = append(current.backendReports, report)
}
