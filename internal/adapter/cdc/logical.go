package cdc

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pglogrepl"
)

func (consumer *Consumer) logical(ctx context.Context, message pglogrepl.Message) error {
	if marker, ok := message.(*pglogrepl.LogicalDecodingMessage); ok {
		return consumer.marker(ctx, marker)
	}

	if relation, ok := message.(*pglogrepl.RelationMessage); ok {
		consumer.relations[relation.RelationID] = relation
		return nil
	}

	if insert, ok := message.(*pglogrepl.InsertMessage); ok {
		return consumer.insert(insert)
	}

	if commit, ok := message.(*pglogrepl.CommitMessage); ok {
		return consumer.commit(ctx, commit.TransactionEndLSN)
	}

	return nil
}

func (consumer *Consumer) insert(message *pglogrepl.InsertMessage) error {
	relation, exists := consumer.relations[message.RelationID]
	if !exists {
		return value.NewError("InvalidInput", "RelationID")
	}

	if relation.RelationName != "outbox" {
		return nil
	}

	index, err := consumer.envelopeColumn(relation)
	if err != nil {
		return err
	}

	if index >= len(message.Tuple.Columns) {
		return value.NewError("InvalidInput", "Tuple")
	}

	return consumer.appendEnvelope(message.Tuple.Columns[index].Data)
}

func (consumer *Consumer) envelopeColumn(relation *pglogrepl.RelationMessage) (int, error) {
	for index, column := range relation.Columns {
		if column.Name == "envelope" {
			return index, nil
		}
	}

	return 0, value.NewError("InvalidInput", "OutboxSchema")
}

func (consumer *Consumer) commit(ctx context.Context, lsn pglogrepl.LSN) error {
	unique, err := event.ProjectionEnvelopes(consumer.pending).Unique()
	if err != nil {
		return err
	}

	consumer.pending = unique
	if err := consumer.projectPending(ctx); err != nil {
		return err
	}

	if err := consumer.config.Store.Acknowledge(ctx, consumer.config.Slot, lsn.String(), consumer.pending); err != nil {
		return err
	}

	consumer.acknowledged = lsn
	consumer.retryDelay = 0
	consumer.pending = nil

	return consumer.feedback(ctx)
}

func (consumer *Consumer) marker(ctx context.Context, message *pglogrepl.LogicalDecodingMessage) error {
	if consumer.stop == 0 {
		return nil
	}

	if message.LSN != consumer.stop {
		return nil
	}

	consumer.acknowledged = message.LSN

	return consumer.feedback(ctx)
}

func (consumer *Consumer) appendEnvelope(data []byte) error {
	envelope := event.ProjectionEnvelope{}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return errors.Join(value.NewError("InvalidInput", "Envelope"), err)
	}

	if err := envelope.Validate(); err != nil {
		return err
	}

	consumer.pending = append(consumer.pending, envelope)

	return nil
}

func (consumer *Consumer) projectPending(ctx context.Context) error {
	for _, envelope := range consumer.pending {
		if err := consumer.config.Store.RegisterEnvelope(ctx, envelope); err != nil {
			return err
		}

		seen, err := consumer.config.Store.Seen(ctx, consumer.config.Slot, envelope)
		if err != nil {
			return err
		}

		if seen {
			continue
		}

		if err := consumer.config.Target.Apply(ctx, envelope.Projection); err != nil {
			return err
		}
	}

	return nil
}
