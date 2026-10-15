package projection_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) eventually(table *godog.Table) error {
	expected := current.rows(table)[0]
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		err := current.searchObserved(ctx, expected)
		if err == nil {
			checkpoint, checkpointErr := current.Store.Checkpoint(ctx, current.Config.Slot)
			current.PreviousCheckpoint = checkpoint

			return checkpointErr
		}

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
		}
	}
}

func (current *world) searchObserved(ctx context.Context, expected map[string]string) error {
	query := application.SearchQuery{Page: 1, Sort: "relevance", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
	result, err := current.Search.Search(ctx, query)
	if err != nil {
		return err
	}

	checkpoint, err := current.Store.Checkpoint(ctx, current.Config.Slot)
	if err != nil {
		return err
	}

	actual := current.searchFields(result, checkpoint)

	if err := current.matches(actual, expected); err != nil {
		return fmt.Errorf("Suchprojektion: %w", err)
	}

	return nil
}

func (current *world) searchFields(result application.SearchResult, checkpoint uint64) map[string]string {
	actual := map[string]string{"TotalProducts": strconv.FormatInt(result.TotalProducts, 10), "CheckpointAdvanced": strconv.FormatBool(checkpoint > current.PreviousCheckpoint)}
	if len(result.Cards) > 0 {
		actual["SKU"] = string(result.Cards[0].SKU)
		actual["Name"] = result.Cards[0].Name
		actual["Price"] = strconv.FormatInt(result.Cards[0].UnitPrice.MinorUnits, 10)
		actual["Available"] = strconv.FormatInt(result.Cards[0].AvailableQuantity, 10)
	}

	return actual
}
