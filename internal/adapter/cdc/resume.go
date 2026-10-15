package cdc

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pglogrepl"
)

func (consumer *Consumer) resume(ctx context.Context) error {
	checkpoint, err := consumer.config.Store.Checkpoint(ctx, consumer.config.Slot)
	if err != nil {
		return err
	}

	state, err := consumer.ensureSlot(ctx, checkpoint)
	if err != nil {
		return err
	}

	position, err := state.Resume(checkpoint)
	consumer.acknowledged = pglogrepl.LSN(position)

	return err
}

func (consumer *Consumer) ensureSlot(ctx context.Context, checkpoint uint64) (postgres.ReplicationState, error) {
	state, err := consumer.config.Store.ReplicationSlot(ctx, consumer.config.Slot)
	if err != nil {
		return state, err
	}

	if state.Exists {
		return state, nil
	}

	if checkpoint != 0 {
		return state, value.NewError("ReindexRequired", "ReplicationSlot")
	}

	_, err = consumer.createSlot(ctx, "NOEXPORT_SNAPSHOT")
	if err := consumer.slotResult(err); err != nil {
		return state, err
	}

	return consumer.config.Store.ReplicationSlot(ctx, consumer.config.Slot)
}
