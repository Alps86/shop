package cdc_test

import (
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) envelopeSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ergänzend die technischen Ereigniswerte validiert werden$`, current.envelopeValues)
}

func (current *world) envelopeValues(input *godog.Table) error {
	for _, line := range input.Rows[1:] {
		row := map[string]string{}
		for index, cell := range line.Cells {
			row[input.Rows[0].Cells[index].Value] = cell.Value
		}

		envelope, err := current.technicalEnvelope(row)
		if err != nil {
			return err
		}

		if code := value.Code(envelope.Validate()); code != row["Code"] {
			return fmt.Errorf("Envelope: %s statt %s", code, row["Code"])
		}
	}

	return nil
}

func (current *world) technicalEnvelope(row map[string]string) (event.ProjectionEnvelope, error) {
	schema, schemaErr := strconv.Atoi(row["SchemaVersion"])
	revision, revisionErr := strconv.ParseUint(row["Revision"], 10, 64)
	count, countErr := strconv.Atoi(row["InventoryCount"])
	if err := errors.Join(schemaErr, revisionErr, countErr); err != nil {
		return event.ProjectionEnvelope{}, err
	}

	at := time.Time{}
	if row["OccurredAt"] != "" {
		var err error
		at, err = time.Parse(time.RFC3339, row["OccurredAt"])
		if err != nil {
			return event.ProjectionEnvelope{}, err
		}
	}

	optional := value.Optional[catalog.Product]{Present: row["ProductPresent"] == "true", Value: catalog.Product{ID: value.ProductID(row["SnapshotID"])}}
	projection := application.ProjectionEvent{ProductID: value.ProductID(row["ProductID"]), Revision: revision, Product: optional, Inventory: make([]sale.Inventory, count)}

	return event.ProjectionEnvelope{EventID: row["EventID"], SchemaVersion: schema, OccurredAt: at, Kind: row["Kind"], Context: value.RequestContext{Locale: value.Locale(row["Locale"]), Currency: "EUR", Market: "DE"}, Projection: projection}, nil
}
