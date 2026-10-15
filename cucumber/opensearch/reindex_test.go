package opensearch_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) reindexSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein vollständiger Quellsnapshot in einen neuen physischen Suchindex aufgebaut wird$`, current.rebuild)
	ctx.Step(`^die Rebuildquelle einen expliziten Fehler liefert$`, current.reindexFault)
	ctx.Step(`^der Rebuildabschluss beobachtet wird$`, current.reindexObserved)
	ctx.Step(`^nach dem Snapshot eine laufende Projektion das Produkt löscht$`, current.lateDeletion)
}

func (current *world) lateDeletion(table *godog.Table) error {
	data := read(table)[0]
	revision, err := strconv.ParseUint(data["Revision"], 10, 64)
	if err != nil {
		return err
	}

	current.lateEvents = []application.ProjectionEvent{{ProductID: value.ProductID(data["ProductID"]), Revision: revision}}

	return nil
}

func (current *world) rebuild(table *godog.Table) {
	data := read(table)[0]
	current.snapshot = event.ReindexSnapshot{Cursor: data["Cursor"]}
	for id, product := range current.products {
		current.snapshot.Events = append(current.snapshot.Events, application.ProjectionEvent{ProductID: id, Revision: 1, Product: value.Optional[catalog.Product]{Present: true, Value: product}, Inventory: current.inventory[id]})
	}

	current.config.Source = current
	if data["Source"] == "absent" {
		current.config.Source = nil
	}

	current.adapter = opensearch.New(current.config)
	current.err = current.adapter.Rebuild(context.Background(), current.config.Index+data["IndexSuffix"], current.now)
}

func (current *world) Snapshot(context.Context, time.Time) (event.ReindexSnapshot, error) {
	if current.sourceFault == "snapshot" {
		return event.ReindexSnapshot{}, value.NewError(current.sourceCode, "Snapshot")
	}

	if current.sourceFault == "replay" {
		current.snapshot.Events = []application.ProjectionEvent{{}}
	}

	return current.snapshot, nil
}

func (current *world) Catchup(ctx context.Context, _ string, _ application.ProjectionPort) error {
	if current.sourceFault == "catchup" {
		return value.NewError(current.sourceCode, "Catchup")
	}

	if current.sourceFault == "concurrent" {
		return current.adapter.Rebuild(ctx, current.config.Index+"-parallel", current.now)
	}

	if current.sourceFault == "dualwrite" {
		current.writerErr = current.adapter.Apply(ctx, application.ProjectionEvent{ProductID: "p-late", Revision: 2})
	}

	for _, event := range current.lateEvents {
		if err := current.adapter.Apply(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) Abort(context.Context, string) error {
	current.abortCount++
	if current.sourceFault == "abort" {
		return value.NewError(current.sourceCode, "Abort")
	}

	return nil
}

func (current *world) reindexFault(table *godog.Table) {
	data := read(table)[0]
	current.sourceFault, current.sourceCode = data["Phase"], data["Code"]
}

func (current *world) reindexObserved(table *godog.Table) error {
	return read(table)[0].matches(map[string]string{"Code": value.Code(current.err), "AbortCount": strconv.Itoa(current.abortCount), "WriterCode": value.Code(current.writerErr)})
}
