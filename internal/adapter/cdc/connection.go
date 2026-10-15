package cdc

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
)

func New(config Config) *Consumer {
	return &Consumer{config: config, relations: map[uint32]*pglogrepl.RelationMessage{}}
}

func (consumer *Consumer) connect(ctx context.Context) error {
	config := consumer.connectionConfig.Copy()
	config.RuntimeParams["replication"] = "database"
	var err error
	consumer.connection, err = pgconn.ConnectConfig(ctx, config)

	return err
}

func (consumer *Consumer) publication(ctx context.Context) error {
	connection, err := pgx.Connect(ctx, consumer.config.DSN)
	if err != nil {
		return err
	}

	consumer.connectionConfig = connection.Config().Config.Copy()
	err = consumer.configurePublication(ctx, connection)

	return errors.Join(err, connection.Close(ctx))
}

func (consumer *Consumer) configurePublication(ctx context.Context, connection *pgx.Conn) error {
	name := pgx.Identifier{consumer.config.Publication}.Sanitize()
	table := pgx.Identifier{consumer.config.Schema, "outbox"}.Sanitize()
	_, err := connection.Exec(ctx, "CREATE PUBLICATION "+name+" FOR TABLE "+table+" WITH (publish='insert')")
	if err := consumer.slotResult(err); err != nil {
		return err
	}

	var total, expected int
	err = connection.QueryRow(ctx, "SELECT count(*),count(*) FILTER(WHERE schemaname=$2 AND tablename='outbox') FROM pg_publication_tables WHERE pubname=$1", consumer.config.Publication, consumer.config.Schema).Scan(&total, &expected)
	if err != nil {
		return err
	}

	if !consumer.validPublication(total, expected) {
		return value.NewError("InvalidInput", "Publication")
	}

	return nil
}

func (consumer *Consumer) createSlot(ctx context.Context, action string) (pglogrepl.CreateReplicationSlotResult, error) {
	return pglogrepl.CreateReplicationSlot(ctx, consumer.connection, pgx.Identifier{consumer.config.Slot}.Sanitize(), "pgoutput", pglogrepl.CreateReplicationSlotOptions{Temporary: action == "EXPORT_SNAPSHOT", Mode: pglogrepl.LogicalReplication, SnapshotAction: action})
}

func (consumer *Consumer) start(ctx context.Context, lsn pglogrepl.LSN) error {
	publication := strings.ReplaceAll(consumer.config.Publication, "'", "''")
	options := pglogrepl.StartReplicationOptions{PluginArgs: []string{"proto_version '1'", "messages 'true'", "publication_names '" + publication + "'"}}

	return pglogrepl.StartReplication(ctx, consumer.connection, consumer.config.Slot, lsn, options)
}

func (consumer *Consumer) Close(ctx context.Context) error {
	if consumer.connection == nil {
		return nil
	}

	return consumer.connection.Close(ctx)
}

func (consumer *Consumer) validPublication(total, expected int) bool {
	return total == 1 && expected == 1
}
