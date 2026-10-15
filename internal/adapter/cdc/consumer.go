package cdc

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
)

func (consumer *Consumer) run(ctx context.Context) error {
	if err := consumer.Prepare(ctx); err != nil {
		return err
	}

	return consumer.Stream(ctx)
}

func (consumer *Consumer) Stream(ctx context.Context) error {
	if err := consumer.start(ctx, consumer.acknowledged); err != nil {
		return consumer.classify(err)
	}

	return consumer.receive(ctx)
}

func (consumer *Consumer) Prepare(ctx context.Context) error {
	consumer.pending = nil
	if err := consumer.Close(ctx); err != nil {
		return err
	}

	if err := consumer.publication(ctx); err != nil {
		return err
	}

	if err := consumer.connect(ctx); err != nil {
		return err
	}

	return consumer.resume(ctx)
}

func (consumer *Consumer) slotResult(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		if postgresError.Code == "42710" {
			return nil
		}
	}

	return err
}

func (consumer *Consumer) receive(ctx context.Context) error {
	for {
		message, err := consumer.connection.ReceiveMessage(ctx)
		if err != nil {
			return err
		}

		data, ok := message.(*pgproto3.CopyData)
		if !ok {
			continue
		}

		if err := consumer.DecodeFrame(ctx, data.Data); err != nil {
			return err
		}

		if consumer.finished() {
			return nil
		}
	}
}

func (consumer *Consumer) DecodeFrame(ctx context.Context, data []byte) error {
	if len(data) == 0 {
		return value.NewError("InvalidInput", "WALData")
	}

	if data[0] == pglogrepl.PrimaryKeepaliveMessageByteID {
		return consumer.keepalive(ctx, data[1:])
	}

	return consumer.decodeLog(ctx, data[1:])
}

func (consumer *Consumer) decodeLog(ctx context.Context, data []byte) error {
	record, err := pglogrepl.ParseXLogData(data)
	if err != nil {
		return err
	}

	if len(record.WALData) == 0 {
		return value.NewError("InvalidInput", "WALData")
	}

	message, err := pglogrepl.Parse(record.WALData)
	if err != nil {
		return err
	}

	return consumer.logical(ctx, message)
}

func (consumer *Consumer) feedback(ctx context.Context) error {
	return pglogrepl.SendStandbyStatusUpdate(ctx, consumer.connection, pglogrepl.StandbyStatusUpdate{WALWritePosition: consumer.acknowledged, WALFlushPosition: consumer.acknowledged, WALApplyPosition: consumer.acknowledged})
}

func (consumer *Consumer) finished() bool {
	return consumer.stop != 0 && consumer.acknowledged >= consumer.stop
}

func (consumer *Consumer) keepalive(ctx context.Context, data []byte) error {
	message, err := pglogrepl.ParsePrimaryKeepaliveMessage(data)
	if err != nil {
		return err
	}

	if message.ReplyRequested {
		return consumer.feedback(ctx)
	}

	return nil
}
