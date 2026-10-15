package runtime_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) catalogObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	snapshot, err := current.app.Store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	actual := map[string]string{"Products": strconv.Itoa(len(snapshot.Products)), "Families": strconv.Itoa(len(snapshot.Families)), "Categories": strconv.Itoa(len(snapshot.Categories))}
	for field, expected := range read(table)[0] {
		if actual[field] != expected {
			return fmt.Errorf("Laufzeitquelle %s: %s statt %s", field, actual[field], expected)
		}
	}

	return nil
}
