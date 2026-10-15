package cdc_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"strconv"
	"time"
)

func (current *world) validationSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^meldet die Snapshotgrenze ihren tatsächlichen Fehler$`, current.snapshotError)
	ctx.Step(`^ein Ereignis innerhalb derselben WAL-Transaktion wiederholt wird$`, current.batchEnvelope)
	ctx.Step(`^ein gespeichertes Ereignis erneut veröffentlicht wird$`, current.repeatEnvelope)
	ctx.Step(`^die Wiederholung ohne weiteren Targetaufruf quittiert wird$`, current.stopAfterCheckpoint)
	ctx.Step(`^ein ungültiges Projektionsereignis tatsächlich committed wird$`, current.invalidEnvelope)
	ctx.Step(`^bleibt der WAL-Checkpoint mit Diagnose unverändert$`, current.diagnosis)
	ctx.Step(`^der exportierte Snapshot wiederholt abgebrochen wird$`, current.abortSnapshot)
	ctx.Step(`^ist der temporäre Reindexslot freigegeben$`, current.slotReleased)
}

func (current *world) invalidEnvelope(input *godog.Table) error {
	row := current.fields(input)
	schema, err := strconv.Atoi(row["SchemaVersion"])
	if err != nil {
		return err
	}

	revision, err := strconv.ParseUint(row["Revision"], 10, 64)
	if err != nil {
		return err
	}

	envelope := event.ProjectionEnvelope{EventID: row["EventID"], SchemaVersion: schema, OccurredAt: time.Now().UTC(), Kind: "delete", Context: value.RequestContext{Locale: value.Locale(row["Locale"]), Currency: "EUR", Market: "DE"}, Projection: application.ProjectionEvent{ProductID: "p-phone", Revision: revision}}
	var payload any = envelope
	if row["RawJSON"] != "" {
		payload = row["RawJSON"]
	}

	_, err = current.admin.Exec(context.Background(), "INSERT INTO "+pgx.Identifier{current.config.Schema, "outbox"}.Sanitize()+"(envelope) VALUES($1::jsonb)", payload)

	return err
}

func (current *world) diagnosis(input *godog.Table) error {
	lsn, err := current.store.Checkpoint(context.Background(), current.config.Slot)
	if err != nil {
		return err
	}

	row := current.fields(input)
	if value.Code(current.err) != row["Code"] {
		return fmt.Errorf("Code: %s statt %s", value.Code(current.err), row["Code"])
	}

	expected := row["LSN"]
	if expected == "unchanged" {
		expected = pglogrepl.LSN(current.priorCheckpoint).String()
	}

	if pglogrepl.LSN(lsn).String() != expected {
		return fmt.Errorf("LSN: %s statt %s", pglogrepl.LSN(lsn).String(), expected)
	}

	return current.checkEventCount(row["EventCount"])
}

func (current *world) checkEventCount(expected string) error {
	if expected == "" {
		return nil
	}

	if strconv.Itoa(len(current.events)) != expected {
		return fmt.Errorf("EventCount: %d", len(current.events))
	}

	return nil
}

func (current *world) fields(input *godog.Table) map[string]string {
	row := map[string]string{}
	for index, cell := range input.Rows[1].Cells {
		row[input.Rows[0].Cells[index].Value] = cell.Value
	}

	return row
}

func (current *world) abortSnapshot() error {
	if err := current.source.Abort(context.Background(), current.snapshot.Cursor); err != nil {
		return err
	}

	return current.source.Abort(context.Background(), current.snapshot.Cursor)
}

func (current *world) slotReleased() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var count int
		if err := current.admin.QueryRow(ctx, "SELECT count(*) FROM pg_replication_slots WHERE slot_name=$1", current.snapshot.Cursor).Scan(&count); err != nil {
			return err
		}

		if count == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (current *world) snapshotError(input *godog.Table) error {
	expected := current.fields(input)["Code"]
	if actual := value.Code(current.err); actual != expected {
		return fmt.Errorf("Code: %s statt %s", actual, expected)
	}

	return nil
}
