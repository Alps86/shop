package runtime_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"os"
	"path/filepath"
	"strings"
)

func (current *world) configureFrontend() {
	directory, err := os.MkdirTemp("", "shop-runtime-content-")
	current.err = err
	current.contentDirectory = directory
	current.config.ContentSource = filepath.Join(directory, "generated")
	current.config.Frontend = runtime.FrontendConfig{PublicOrigin: "http://localhost", ContentDirectory: directory, CacheEndpoint: "http://127.0.0.1:1", CacheKey: "runtime-cache-key-01234567890123456789", CSRFKey: "runtime-csrf-key-01234567890123456789", BackendListen: "127.0.0.1:0"}
}

func (current *world) frontendEnvironment() []string {
	config := current.config
	frontend := config.Frontend

	return append(os.Environ(), "SHOP_PUBLIC_ORIGIN="+frontend.PublicOrigin, "SHOP_CONTENT_DIRECTORY="+frontend.ContentDirectory, "SHOP_CACHE_ENDPOINT="+frontend.CacheEndpoint, "SHOP_CACHE_KEY="+frontend.CacheKey, "SHOP_CSRF_KEY="+frontend.CSRFKey, "SHOP_BACKEND_LISTEN="+frontend.BackendListen, "SHOP_CONTENT_SOURCE="+config.ContentSource)
}

func (current *world) frontendFields(config *runtime.Config) map[string]*string {
	frontend := &config.Frontend
	return map[string]*string{"PublicOrigin": &frontend.PublicOrigin, "ContentDirectory": &frontend.ContentDirectory, "CacheEndpoint": &frontend.CacheEndpoint, "CacheKey": &frontend.CacheKey, "CSRFKey": &frontend.CSRFKey, "BackendListen": &frontend.BackendListen, "ContentSource": &config.ContentSource}
}

func (current *world) contentSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die erzeugten Hugo-Inhalte enthalten$`, current.contentObserved)
	ctx.Step(`^der Katalog in ein technisches Inhaltsziel exportiert wird$`, current.exportContent)
	ctx.Step(`^der Inhaltsexport ein Dateisystemergebnis liefert$`, current.exportObserved)
}

func (current *world) contentObserved(table *godog.Table) error {
	for _, data := range read(table) {
		content, err := os.ReadFile(filepath.Join(current.config.ContentSource, data["File"]))
		if err != nil {
			return err
		}

		if !strings.Contains(string(content), data["Contains"]) {
			return fmt.Errorf("Hugo-Inhalt %s enthält nicht %s: %s", data["File"], data["Contains"], content)
		}
	}

	return nil
}

func (current *world) exportContent(table *godog.Table) error {
	data := read(table)[0]
	target := filepath.Join(current.contentDirectory, data["Path"])
	if err := current.prepareContentTarget(target, data); err != nil {
		return err
	}

	current.err = current.app.ExportContent(context.Background(), target)

	return nil
}

func (current *world) prepareContentTarget(target string, data map[string]string) error {
	if data["Existing"] == "file" {
		return os.WriteFile(target, []byte(data["Body"]), 0600)
	}

	if data["Existing"] == "directory-entry" {
		return os.MkdirAll(filepath.Join(target, data["Body"]), 0750)
	}

	return nil
}

func (current *world) exportObserved(table *godog.Table) error {
	expected := read(table)[0]["Error"] == "true"
	if (current.err != nil) != expected {
		return fmt.Errorf("Inhaltsexport: %w, Fehler erwartet %t", current.err, expected)
	}

	return nil
}
