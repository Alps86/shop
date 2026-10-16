package demo_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (current *world) orderObserved(table *godog.Table) error {
	data := read(table)[0]
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		location := current.orderLocation()
		if strings.HasPrefix(location, "/de-DE/bestellung/") {
			current.orderPath = location
			return current.checkOrder(data)
		}

		if err := current.get("/de-DE/kasse", ""); err != nil {
			return err
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("keine Bestellung: HTTP %d %s", current.response.status, current.response.diagnostic())
}

func (current *world) orderLocation() string {
	if target := current.response.header.Get("HX-Redirect"); target != "" {
		return target
	}

	return current.response.header.Get("Location")
}

func (current *world) session() application.Session {
	endpoint, _ := url.Parse(current.deployment.public)
	for _, cookie := range current.client.Jar.Cookies(endpoint) {
		if cookie.Name == "shop_session" {
			return application.Session(cookie.Value)
		}
	}

	return ""
}

func (current *world) service() application.Service {
	store := current.deployment.store
	return application.Service{Guests: store.Guests(), Carts: store.Carts(), Orders: store.Orders(), Fulfillment: store.Fulfillment(), Clock: opensearch.Clock{}}
}

func (current *world) checkOrder(data map[string]string) error {
	id := value.OrderID(strings.TrimPrefix(current.orderPath, "/de-DE/bestellung/"))
	if current.differentOrder(id) {
		return fmt.Errorf("Wiederholung erzeugte andere Bestellung")
	}

	order, err := current.service().GetOrder(context.Background(), current.session(), id)
	if err != nil {
		return err
	}

	current.order = order
	actual := order.Purchase
	observed := map[string]string{"PaymentMethod": string(actual.PaymentMethod), "ShippingMethod": string(actual.ShippingMethod), "Items": strconv.Itoa(len(actual.Items)), "TotalMinor": strconv.FormatInt(actual.Total.MinorUnits, 10)}
	for key, expected := range data {
		if observed[key] != expected {
			return fmt.Errorf("Kaufsnapshot %s=%q statt %q", key, observed[key], expected)
		}
	}

	return current.get(current.orderPath, "")
}

func (current *world) differentOrder(id value.OrderID) bool {
	return current.order.Purchase.ID != "" && current.order.Purchase.ID != id
}
