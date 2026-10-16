package performance_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/adapter/postgres"
	shoptemporal "github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/client"
	"os"
	"strconv"
)

func (current *world) open(table *godog.Table) error {
	current.data = read(table)[0]
	current.public = os.Getenv("SHOP_PERFORMANCE_PUBLIC_URL")
	if current.public == "" {
		return fmt.Errorf("SHOP_PERFORMANCE_PUBLIC_URL muss die echte isolierte Benchmarklaufzeit benennen")
	}

	current.directory = os.Getenv("SHOP_PERFORMANCE_REPORT_DIRECTORY")
	if current.directory == "" {
		directory, err := os.MkdirTemp("", "shop-performance-")
		current.directory = directory
		if err != nil {
			return err
		}
	}

	if err := current.validateEnvironment(); err != nil {
		return err
	}

	return current.connect()
}

func (current *world) connect() error {
	config := postgres.Config{DSN: os.Getenv("SHOP_PERFORMANCE_POSTGRES_DSN"), Schema: os.Getenv("SHOP_PERFORMANCE_SCHEMA"), Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
	store, err := postgres.Open(context.Background(), config)
	current.store = store
	if err != nil {
		return err
	}

	connection, err := client.Dial(client.Options{HostPort: os.Getenv("SHOP_PERFORMANCE_TEMPORAL_ADDRESS"), Namespace: current.data["Namespace"]})
	current.temporal = connection
	if err != nil {
		return err
	}

	current.bindService()

	return current.prepareDataset()
}

func (current *world) bindService() {
	store := current.store
	service := application.Service{Guests: store.Guests(), Carts: store.Carts(), Orders: store.Orders(), Clock: opensearch.Clock{}, Starter: store.Carts(), Fulfillment: store.Fulfillment()}
	processes := shoptemporal.New(current.temporal, service, store.Carts())
	processes.Queue = os.Getenv("SHOP_PERFORMANCE_TASK_QUEUE")
	service.Processes = processes
	current.service = service
}

func (current *world) prepareDataset() error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	variants := 0
	for _, product := range snapshot.Products {
		variants += len(product.Variants)
	}

	observed := map[string]int{"Products": len(snapshot.Products), "Variants": variants, "Categories": len(snapshot.Categories)}
	for key, count := range observed {
		if strconv.Itoa(count) != current.data[key] {
			return fmt.Errorf("Messdatensatz %s=%d statt %s", key, count, current.data[key])
		}
	}

	return current.prepareOrders()
}
