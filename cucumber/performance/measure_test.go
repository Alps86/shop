package performance_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (current *world) measure(table *godog.Table) (err error) {
	if err := current.prepareMeasurements(); err != nil {
		return err
	}

	defer func() {
		err = current.finishReport(err)
	}()

	for _, data := range read(table) {
		if err := current.loadProfile(data); err != nil {
			return err
		}

		if err := current.measureRoutes(data); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) measureRoute(route string, data map[string]string) error {
	endpoint, cookie, contains := current.route(route)
	if endpoint == "" {
		return fmt.Errorf("unbekannte Messroute %s", route)
	}

	if err := current.probe(probeRequest{route: route, origin: current.public, path: endpoint, cookie: cookie, contains: contains}); err != nil {
		return fmt.Errorf("Inhaltsprobe %s: %w", route, err)
	}

	if data["Cache"] == "warm" {
		return current.warm(route, endpoint, cookie, data)
	}

	return current.cold(route, endpoint, cookie, data)
}

func (current *world) route(route string) (string, string, string) {
	paths := map[string]string{"home": "/de-DE/", "category": current.data["CategoryPath"], "search": current.data["SearchPath"], "personalized-search": current.data["SearchPath"], "product": current.data["ProductPath"], "cart": "/de-DE/warenkorb", "checkout": "/de-DE/kasse", "order": "/de-DE/bestellung/" + string(current.order)}
	cookies := map[string]string{"personalized-search": current.cookie, "cart": current.data["CartCookie"], "checkout": current.data["CartCookie"], "order": current.cookie}
	contains := map[string]string{"home": "Lichtblick Technik", "category": current.data["ListingContains"], "search": current.data["ListingContains"], "personalized-search": current.data["ListingContains"], "product": current.data["SKU"], "cart": current.data["SKU"], "checkout": current.data["SKU"], "order": "Unveränderlicher Kauf"}

	return paths[route], cookies[route], contains[route]
}

func (current *world) reportPath(route, cache, suffix string) string {
	return filepath.Join(current.directory, route+"-"+cache+suffix)
}

func (current *world) measureRoutes(data map[string]string) error {
	for _, route := range strings.Split(data["Routes"], ",") {
		if err := current.measureRoute(route, data); err != nil {
			current.diagnoseFailure(route, data["Boundary"])
			return fmt.Errorf("Messroute %s/%s: %w", route, data["Cache"], err)
		}
	}

	return nil
}

func (current *world) loadProfile(data map[string]string) error {
	if !map[string]bool{"requestStart-to-last-composed-HTML-byte": true, "requestStart-to-last-Go-response-byte": true, "requestStart-to-last-Go-fragment-byte": true}[data["Boundary"]] {
		return fmt.Errorf("unbekannte Messgrenze %s", data["Boundary"])
	}

	if !map[string]bool{"warm": true, "cold-cache-services-ready": true, "direct-services-ready": true}[data["Cache"]] {
		return fmt.Errorf("unbekanntes Cacheprofil %s", data["Cache"])
	}

	current.profile = measurementProfile{Boundary: data["Boundary"], MutatingCommandsIncluded: false}

	return current.parseProfile(data)
}

func (current *world) parseProfile(data map[string]string) error {
	inputs := map[string]string{"RequestsPerRoute": data["RequestsPerRoute"], "WarmupPerRoute": data["WarmupPerRoute"], "Concurrency": current.data["Concurrency"]}
	targets := map[string]*int{"RequestsPerRoute": &current.expectedRequests, "WarmupPerRoute": &current.profile.WarmupPerRoute, "Concurrency": &current.profile.Concurrency}
	for key, target := range targets {
		value, err := strconv.Atoi(inputs[key])
		if err != nil {
			return err
		}

		*target = value
	}

	return nil
}

func (current *world) prepareMeasurements() error {
	if err := current.initializeInternal(); err != nil {
		return err
	}

	return os.MkdirAll(current.directory, 0750)
}

func (current *world) initializeInternal() error {
	if current.internal == "" {
		current.internal = os.Getenv("SHOP_PERFORMANCE_INTERNAL_URL")
	}

	return validateInternalOrigin(current.internal)
}
