package cdc_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/cucumber/godog"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strconv"
	"strings"
	"time"
)

func (current *world) databaseFaultSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^derselbe Slotname zu einer anderen eigenen Testdatenbank gehört$`, current.foreignSlot)
	ctx.Step(`^folgende CDC-SQL-Fixture nur eigene Artefakte verändert$`, current.sqlFixture)
	ctx.Step(`^der eigene bestätigte Replikationsslot verloren geht$`, current.loseSlot)
	ctx.Step(`^derselbe Verbraucher erneut vorbereitet wird$`, current.prepareConsumer)
	ctx.Step(`^ein echter Reindexsnapshot versucht wird$`, current.attemptSnapshot)
	ctx.Step(`^meldet die WAL-Vorbereitung ihren tatsächlichen Fehler$`, current.snapshotError)
	ctx.Step(`^der eigene SQL-Verbindungspool geschlossen wird$`, current.closeStore)
	ctx.Step(`^die eigene exportierte Replikationsverbindung serverseitig beendet wird$`, current.terminateSnapshot)
	ctx.Step(`^ein schon gekündigter nativer Runnerstart versucht wird$`, current.canceledRunner)
	ctx.Step(`^der eigene laufende Replikationsbefehl serverseitig gekündigt wird$`, current.cancelReplication)
	ctx.Step(`^ein ungültiger eigener Slotname konfiguriert ist$`, current.invalidSlot)
	ctx.Step(`^ein langer gültiger eigener Slotname konfiguriert ist$`, current.longSlot)
}

func (current *world) longSlot(input *godog.Table) error {
	row := current.fields(input)
	length, err := strconv.Atoi(row["Length"])
	if err != nil {
		return err
	}

	current.config.Slot = strings.Repeat(row["Character"], length)
	current.source = cdc.NewReindexSource(current.config)

	return nil
}

func (current *world) invalidSlot(input *godog.Table) {
	current.config.Slot = current.fields(input)["Slot"]
	current.source = cdc.NewReindexSource(current.config)
}

func (current *world) cancelReplication() error {
	_, err := current.admin.Exec(context.Background(), "SELECT pg_cancel_backend(active_pid) FROM pg_replication_slots WHERE slot_name=$1", current.config.Slot)

	return err
}

func (current *world) canceledRunner() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	current.consumer = cdc.New(current.config)
	current.err = current.consumer.Run(ctx)
}

func (current *world) terminateSnapshot() error {
	_, err := current.admin.Exec(context.Background(), "SELECT pg_terminate_backend(active_pid) FROM pg_replication_slots WHERE slot_name=$1", current.snapshot.Cursor)
	if err != nil {
		return err
	}

	return current.slotReleased()
}

func (current *world) sqlFixture(input *godog.Table) error {
	replacements := strings.NewReplacer("$schema", pgx.Identifier{current.config.Schema}.Sanitize(), "$publication", pgx.Identifier{current.config.Publication}.Sanitize())
	for _, row := range input.Rows[1:] {
		if _, err := current.admin.Exec(context.Background(), replacements.Replace(row.Cells[0].Value)); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) loseSlot() error {
	var err error
	current.priorCheckpoint, err = current.store.Checkpoint(context.Background(), current.config.Slot)
	if err != nil {
		return err
	}

	_, err = current.admin.Exec(context.Background(), "SELECT pg_drop_replication_slot($1)", current.config.Slot)

	return err
}

func (current *world) prepareConsumer() {
	current.consumer = cdc.New(current.config)
	current.err = current.consumer.Prepare(context.Background())
}

func (current *world) foreignSlot() error {
	current.foreignDatabase = current.config.Schema + "_db"
	if _, err := current.admin.Exec(context.Background(), "CREATE DATABASE "+pgx.Identifier{current.foreignDatabase}.Sanitize()); err != nil {
		return err
	}

	config, err := pgconn.ParseConfig(current.config.DSN)
	if err != nil {
		return err
	}

	config.Database = current.foreignDatabase
	config.RuntimeParams["replication"] = "database"
	connection, err := pgconn.ConnectConfig(context.Background(), config)
	if err != nil {
		return err
	}

	return current.createForeignSlot(connection)
}

func (current *world) createForeignSlot(connection *pgconn.PgConn) error {
	_, slotErr := pglogrepl.CreateReplicationSlot(context.Background(), connection, pgx.Identifier{current.config.Slot}.Sanitize(), "pgoutput", pglogrepl.CreateReplicationSlotOptions{Mode: pglogrepl.LogicalReplication, SnapshotAction: "NOEXPORT_SNAPSHOT"})
	closeErr := connection.Close(context.Background())
	if slotErr != nil {
		return slotErr
	}

	return closeErr
}

func (current *world) attemptSnapshot() {
	current.snapshot, current.err = current.source.Snapshot(context.Background(), time.Now())
}

func (current *world) closeStore() {
	current.store.Close()
}
