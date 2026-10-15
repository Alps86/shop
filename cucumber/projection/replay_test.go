package projection_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"time"
)

func (current *world) registerReplay(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein früher bestätigter Projektionsstand aus PostgreSQL gesichert wird$`, current.capture)
	ctx.Step(`^der dauerhafte WAL-Verbraucher vorübergehend angehalten wird$`, current.stopConsumer)
	ctx.Step(`^derselbe dauerhafte WAL-Slot mit neuem Verbraucher fortgesetzt wird$`, current.startConsumer)
	ctx.Step(`^derselbe frühere Stand erneut an OpenSearch zugestellt wird$`, current.replay)
	ctx.Step(`^das Produkt mit einem dauerhaften Tombstone gelöscht wird$`, current.deleteProduct)
	ctx.Step(`^ein vollständiger echter Rebuild durchgeführt wird$`, current.fullRebuild)
	ctx.Step(`^dasselbe Produkt bewusst als neue Datensatzversion veröffentlicht wird$`, current.republish)
}

func (current *world) capture(table *godog.Table) error {
	snapshot, err := current.Source.Source.Snapshot(context.Background(), time.Now())
	if err != nil {
		return err
	}

	for _, event := range snapshot.Events {
		if string(event.ProductID) == current.rows(table)[0]["ProductID"] {
			current.Earlier = event
		}
	}

	if current.Earlier.ProductID == "" {
		return errors.Join(fmt.Errorf("Projektionsstand fehlt"), current.Source.Source.Abort(context.Background(), snapshot.Cursor))
	}

	return current.Source.Source.Abort(context.Background(), snapshot.Cursor)
}

func (current *world) replay(table *godog.Table) error {
	attempts, err := strconv.Atoi(current.rows(table)[0]["Attempts"])
	if err != nil {
		return err
	}

	for range attempts {
		if err := current.Search.Apply(context.Background(), current.Earlier); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) deleteProduct(table *godog.Table) error {
	return current.Store.DeleteProduct(context.Background(), value.ProductID(current.rows(table)[0]["ProductID"]))
}

func (current *world) fullRebuild(table *godog.Table) error {
	current.TargetIndex = current.Alias + "-" + current.rows(table)[0]["IndexSuffix"]

	return current.Search.Rebuild(context.Background(), current.TargetIndex, time.Now())
}

func (current *world) walObserved(table *godog.Table) error {
	var plugin string
	var persistent bool
	err := current.Admin.QueryRow(context.Background(), "SELECT plugin,NOT temporary FROM pg_replication_slots WHERE slot_name=$1", current.Config.Slot).Scan(&plugin, &persistent)
	if err != nil {
		return err
	}

	tables := []string{}
	err = current.Admin.QueryRow(context.Background(), "SELECT array_agg(tablename ORDER BY tablename) FROM pg_publication_tables WHERE pubname=$1", current.Config.Publication).Scan(&tables)
	if err != nil {
		return err
	}

	checkpoint, err := current.Store.Checkpoint(context.Background(), current.Config.Slot)
	if err != nil {
		return err
	}

	return current.matches(map[string]string{"Plugin": plugin, "Persistent": strconv.FormatBool(persistent), "PublishedTables": strings.Join(tables, ","), "CheckpointPresent": strconv.FormatBool(checkpoint > 0)}, current.rows(table)[0])
}

func (current *world) republish(table *godog.Table) error {
	row := current.rows(table)[0]
	current.Manifest.DatasetVersion = row["DatasetVersion"]
	current.Manifest.Products[0].Name["de-DE"] = row["Name"]
	_, err := current.Store.Import(context.Background(), current.Manifest)

	return err
}
