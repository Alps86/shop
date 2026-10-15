package commerce_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"os"
	"strings"
	"time"
)

func (current *world) open(table *godog.Table) error {
	*current = world{codes: map[string]string{}, answers: map[string]map[string]string{}}
	data := read(table)[0]
	current.configure(data)
	var err error
	current.app, err = runtime.Open(context.Background(), current.config)
	if err != nil {
		return err
	}

	if err := current.importCatalog(data["Fixture"]); err != nil {
		return err
	}

	current.start()

	return current.awaitDispatch()
}

func (current *world) configure(data row) {
	suffix := strings.ToLower(rand.Text()[:10])
	current.config = runtime.Config{DSN: os.Getenv("SHOP_TEST_POSTGRES_DSN"), Schema: "commerce_" + suffix, TemporalAddress: os.Getenv("TEMPORAL_ADDRESS"), TemporalNamespace: data["Namespace"], TaskQueue: "commerce_" + suffix, SearchEndpoint: os.Getenv("SHOP_TEST_OPENSEARCH_URL"), SearchIndex: "shop-commerce-" + suffix, Publication: "commerce_pub_" + suffix, Slot: "commerce_slot_" + suffix, Listen: "127.0.0.1:0", Context: value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: value.Currency(data["Currency"]), Market: value.CountryCode(data["Market"])}}
	if data["OpenSearch"] == "down" {
		current.config.SearchEndpoint = data["SearchEndpoint"]
	}
}

func (current *world) importCatalog(path string) error {
	payload, err := current.fixture(path)
	if err != nil {
		return err
	}

	var manifest postgres.ImportManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return err
	}

	_, err = current.app.Store.Import(context.Background(), manifest)

	return err
}

func (current *world) start() {
	ctx, cancel := context.WithCancel(context.Background())
	current.cancel = cancel
	current.finished = make(chan error, 1)
	go func() {
		current.finished <- current.app.Run(ctx)
	}()
}

func (current *world) awaitDispatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		status, err := current.app.Snapshot(ctx)
		if err != nil {
			return err
		}

		if status["ProcessOutbox"] == "ready" {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (current *world) fixture(path string) ([]byte, error) {
	root, err := os.OpenRoot("../..")
	if err != nil {
		return nil, err
	}

	payload, err := root.ReadFile(path)

	return payload, errors.Join(err, root.Close())
}
