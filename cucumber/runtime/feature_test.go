package runtime_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @composition", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "runtime", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Laufzeitkomposition fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.After(current.cleanup)
	ctx.Step(`^die Laufzeit mit echten Dienstverbindungen geöffnet wird$`, current.open)
	ctx.Step(`^die Laufzeitdiagnose per HTTP abgefragt wird$`, current.request)
	ctx.Step(`^liefert die Laufzeitdiagnose$`, current.observed)
	ctx.Step(`^Worker Projektor und HTTP-Dienst gestartet werden$`, current.start)
	ctx.Step(`^die vollständige Bereitschaft begrenzt abgewartet wird$`, current.awaitReady)
	current.commandSteps(ctx)
	current.configurationSteps(ctx)
	current.failureSteps(ctx)
	current.permissionSteps(ctx)
	ctx.Step(`^die veröffentlichte Laufzeitquelle enthält$`, current.catalogObserved)
}

func (current *world) open(table *godog.Table) {
	*current = world{}
	data := read(table)[0]
	suffix := strings.ToLower(rand.Text()[:10])
	current.config = runtime.Config{DSN: os.Getenv(data["PostgresEnv"]), Schema: "runtime_" + suffix, TemporalAddress: os.Getenv(data["TemporalEnv"]), TemporalNamespace: data["Namespace"], TaskQueue: "runtime_" + suffix, SearchEndpoint: os.Getenv(data["SearchEnv"]), SearchIndex: "shop-runtime-" + suffix, Publication: "runtime_pub_" + suffix, Slot: "runtime_slot_" + suffix, Listen: data["Listen"], Context: value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: value.Currency(data["Currency"]), Market: value.CountryCode(data["Market"])}}
	current.config.SeedFile = data["SeedFile"]
	current.openApplication()
}

func (current *world) openApplication() {
	if current.config.SeedFile != "" {
		absolute, err := filepath.Abs(current.config.SeedFile)
		if err != nil {
			current.err = err
			return
		}

		current.config.SeedFile = absolute
	}

	current.app, current.err = runtime.Open(context.Background(), current.config)
	if current.err == nil {
		current.server = httptest.NewServer(current.app.Handler)
		current.baseURL = current.server.URL
	}
}

func (current *world) observed(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	for _, data := range read(table) {
		if !current.matches(data) {
			return fmt.Errorf("HTTP %d: %s, erwartet %v", current.status, current.body, data)
		}
	}

	return nil
}
