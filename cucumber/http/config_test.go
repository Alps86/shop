package http_test

import (
	"encoding/json"
	"fmt"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"testing/fstest"
	"time"
)

func (current *world) configure(table *godog.Table) error {
	current.options = read(table)[0]
	current.fragmentFault.Store(current.options["FragmentFault"] == "connectionClose")
	if source, exists := current.options["Routes"]; exists {
		return json.Unmarshal([]byte(source), &current.fixture.Routes)
	}

	return nil
}

func (current *world) optionsFor(config shophttp.Config) shophttp.Config {
	config = current.optionalPorts(config)
	config.ContentDirectory = current.options["ContentDirectory"]
	config = current.browserOptions(config)

	return current.templateOptions(config)
}

func (current *world) construct(table *godog.Table) error {
	factory := shophttp.New
	if read(table)[0]["Listener"] == "backend" {
		factory = shophttp.NewBackend
	}

	current.handler, current.err = factory(current.config())

	return nil
}

func (current *world) initializationResult(table *godog.Table) error {
	if current.err == nil {
		return fmt.Errorf("erwarteter Initialisierungsfehler fehlt")
	}

	if !strings.Contains(current.err.Error(), read(table)[0]["Error"]) {
		return fmt.Errorf("unerwarteter Fehler: %w", current.err)
	}

	return nil
}

func (current *world) Now() time.Time {
	return current.fixture.Now
}

func readCount(input string) int {
	count, _ := strconv.Atoi(input)
	return count
}

func (current *world) optionalPorts(config shophttp.Config) shophttp.Config {
	if current.options["Preview"] == "absent" {
		config.Preview = nil
	}

	if current.options["Variants"] == "absent" {
		config.Variants = nil
	}

	if current.options["Recommendations"] == "absent" {
		config.Recommendations = nil
	}

	if current.options["Interests"] == "absent" {
		config.Interests = nil
	}

	return config
}

func (current *world) browserOptions(config shophttp.Config) shophttp.Config {
	if current.options["Search"] == "absent" {
		config.Search = nil
	}

	if current.options["Diagnostics"] == "absent" {
		config.Diagnostics = nil
	}

	if current.options["SecureCookie"] == "true" {
		config.SecureCookie = true
	}

	if current.options["Clock"] == "fixture" {
		config.Clock = current
	}

	return config
}

func (current *world) templateOptions(config shophttp.Config) shophttp.Config {
	if key, exists := current.options["CSRFKey"]; exists {
		config.CSRFKey = []byte(key)
	}

	if source, exists := current.options["Template"]; exists {
		config.TemplateFS = fstest.MapFS{"template/page.html": &fstest.MapFile{Data: []byte(source)}}
	}

	return config
}
