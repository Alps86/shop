package demo_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/Alps86/shop/internal/runtime"
	"github.com/cucumber/godog"
	"go.temporal.io/sdk/client"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func (current *world) open(table *godog.Table) error {
	data := read(table)[0]
	current.deployment = &deployment{}
	if os.Getenv("SHOP_DEMO_DEPLOYMENT") == "compose" {
		return current.deployment.connectCompose(data)
	}

	return current.deployment.startNative(data)
}

func (deployment *deployment) connectCompose(data map[string]string) error {
	deployment.external = true
	deployment.public = os.Getenv("SHOP_DEMO_PUBLIC_URL")
	deployment.backend = os.Getenv("SHOP_DEMO_BACKEND_URL")
	if deployment.missingURLs() {
		return fmt.Errorf("Compose benötigt öffentliche und getrennte Backend-URL")
	}

	deployment.config = connections(data)
	deployment.config.Schema = "demo"
	deployment.config.SearchIndex = "shop-products"
	deployment.config.TaskQueue = "shop"
	store, err := postgres.Open(context.Background(), postgres.Config{DSN: deployment.config.DSN, Schema: deployment.config.Schema, Context: deployment.config.Context})
	deployment.store = store
	if err != nil {
		return err
	}

	deployment.connection, err = client.Dial(client.Options{HostPort: deployment.config.TemporalAddress, Namespace: deployment.config.TemporalNamespace})

	return err
}

func (deployment *deployment) missingURLs() bool {
	return deployment.public == "" || deployment.backend == ""
}

func connections(data map[string]string) runtime.Config {
	return runtime.Config{DSN: os.Getenv(data["PostgresEnv"]), TemporalAddress: os.Getenv(data["TemporalEnv"]), TemporalNamespace: data["Namespace"], SearchEndpoint: os.Getenv(data["SearchEnv"]), Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
}

func (deployment *deployment) startNative(data map[string]string) error {
	directory, err := os.MkdirTemp("", "shop-demo-")
	if err != nil {
		return err
	}

	deployment.directory = directory
	deployment.config = connections(data)
	suffix := strings.ToLower(rand.Text()[:10])
	deployment.config.Schema = "demo_" + suffix
	deployment.config.SearchIndex = "shop-demo-" + suffix
	deployment.config.TaskQueue = "demo_" + suffix
	deployment.config.Publication = "demo_pub_" + suffix
	deployment.config.Slot = "demo_slot_" + suffix
	deployment.config.SeedFile, err = filepath.Abs("../../" + data["SeedFile"])
	if err != nil {
		return err
	}

	return deployment.prepareNative(data)
}

func (deployment *deployment) reserveAddress() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	if deployment.reserved == nil {
		deployment.reserved = make(map[string]net.Listener)
	}

	address := listener.Addr().String()
	deployment.reserved[address] = listener

	return address, nil
}

func binary(variable, name string) string {
	if configured := os.Getenv(variable); configured != "" {
		return configured
	}

	return name
}
