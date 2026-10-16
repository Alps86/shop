package demo_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (current *world) concurrentPurchases(table *godog.Table) error {
	seconds, err := strconv.Atoi(read(table)[0]["TimeoutSeconds"])
	if err != nil {
		return err
	}

	start := make(chan struct{})
	results := make(chan purchaseAttempt, len(current.prepared))
	for jar, form := range current.prepared {
		go func() {
			<-start
			result, err := current.purchaseForm(jar, form, time.Duration(seconds)*time.Second)
			results <- purchaseAttempt{jar: jar, response: result, err: err}
		}()
	}

	close(start)
	current.collectAttempts(results)

	return nil
}

func (current *world) collectAttempts(results <-chan purchaseAttempt) {
	for range len(current.prepared) {
		current.attempts = append(current.attempts, <-results)
	}
}

func (current *world) purchaseForm(jar string, form url.Values, timeout time.Duration) (response, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := current.sendPurchase(jar, form)
		if err != nil {
			return result, err
		}

		if result.status != http.StatusAccepted {
			return result, nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return response{}, fmt.Errorf("Gastabschluss %s bleibt ausstehend", jar)
}

func (current *world) sendPurchase(jar string, form url.Values) (response, error) {
	request, err := http.NewRequest(http.MethodPost, current.deployment.public+"/de-DE/warenkorb/aktion", strings.NewReader(form.Encode()))
	if err != nil {
		return response{}, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	result, err := current.clients[jar].Do(request)
	if err != nil {
		return response{}, err
	}

	body, err := io.ReadAll(result.Body)

	return acceptResponse(result, body), errors.Join(err, result.Body.Close())
}

func (current *world) concurrentObserved(table *godog.Table) error {
	data := read(table)[0]
	counts := map[string]int64{"Orders": 0, "Unavailable": 0}
	for _, attempt := range current.attempts {
		if err := current.countAttempt(attempt, counts); err != nil {
			return err
		}
	}

	if err := current.countInventory(data["SKU"], counts); err != nil {
		return err
	}

	for key, actual := range counts {
		if strconv.FormatInt(actual, 10) != data[key] {
			return fmt.Errorf("Konkurrenz %s=%d statt %s", key, actual, data[key])
		}
	}

	return nil
}

func (current *world) countAttempt(attempt purchaseAttempt, counts map[string]int64) error {
	if attempt.err != nil {
		return attempt.err
	}

	id, err := current.deployment.store.Orders().ForCart(context.Background(), value.CartID(current.prepared[attempt.jar].Get("cartID")))
	if attempt.response.status == http.StatusSeeOther {
		return current.countOrder(id, err, counts)
	}

	if attempt.response.status != http.StatusConflict {
		return fmt.Errorf("Konkurrenzantwort %d: %s", attempt.response.status, attempt.response.diagnostic())
	}

	return current.countUnavailable(err, counts)
}

func (current *world) countOrder(id value.OrderID, err error, counts map[string]int64) error {
	if err != nil {
		return err
	}

	counts["Orders"]++
	current.orders = append(current.orders, id)

	return nil
}

func (current *world) countUnavailable(err error, counts map[string]int64) error {
	if value.Code(err) != "NotFound" {
		return fmt.Errorf("abgewiesener Gast besitzt Bestellung: %w", err)
	}

	counts["Unavailable"]++

	return nil
}

func (current *world) countInventory(sku string, counts map[string]int64) error {
	snapshot, err := current.deployment.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	for _, inventory := range snapshot.Inventory {
		if string(inventory.SKU) != sku {
			continue
		}

		counts["Stock"] = inventory.StockQuantity
		for _, reservation := range inventory.Reservations {
			counts["Reserved"] += reservation.Quantity
		}
	}

	return nil
}
