package cdc

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pglogrepl"
	"strings"
	"time"
)

func NewReindexSource(config Config) *ReindexSource {
	return &ReindexSource{Config: config, snapshots: map[string]*Consumer{}}
}

func (source *ReindexSource) Snapshot(ctx context.Context, _ time.Time) (event.ReindexSnapshot, error) {
	config := source.Config
	config.Slot = config.Slot[:min(len(config.Slot), 24)] + "_reindex_" + strings.ToLower(rand.Text())
	consumer := New(config)
	if err := consumer.publication(ctx); err != nil {
		return event.ReindexSnapshot{}, err
	}

	if err := consumer.connect(ctx); err != nil {
		return event.ReindexSnapshot{}, err
	}

	snapshot, err := consumer.export(ctx)
	if err != nil {
		return event.ReindexSnapshot{}, errors.Join(err, consumer.Close(ctx))
	}

	source.register(config.Slot, consumer)

	return snapshot, nil
}

func (consumer *Consumer) export(ctx context.Context) (event.ReindexSnapshot, error) {
	slot, err := consumer.createSlot(ctx, "EXPORT_SNAPSHOT")
	if err != nil {
		return event.ReindexSnapshot{}, err
	}

	events, err := consumer.config.Store.ExportedSnapshot(ctx, slot.SnapshotName)
	if err != nil {
		return event.ReindexSnapshot{}, err
	}

	state, err := consumer.config.Store.ReplicationSlot(ctx, consumer.config.Slot)
	if err != nil {
		return event.ReindexSnapshot{}, err
	}

	consumer.acknowledged = pglogrepl.LSN(state.Confirmed)

	return event.ReindexSnapshot{Events: events, Cursor: consumer.config.Slot}, nil
}

func (source *ReindexSource) Catchup(ctx context.Context, cursor string, target application.ProjectionPort) error {
	consumer, exists := source.claim(cursor)
	if !exists {
		return value.NewError("NotFound", "ReindexCursor")
	}

	err := consumer.catchupTarget(ctx, target)

	return errors.Join(err, consumer.Close(ctx))
}

func (consumer *Consumer) catchupTarget(ctx context.Context, target application.ProjectionPort) error {
	consumer.config.Target = target
	stop, err := consumer.config.Store.WALBarrier(ctx)
	if err != nil {
		return err
	}

	consumer.stop = pglogrepl.LSN(stop)

	if err := consumer.start(ctx, consumer.acknowledged); err != nil {
		return err
	}

	return consumer.receive(ctx)
}

func (source *ReindexSource) register(cursor string, consumer *Consumer) {
	source.mutex.Lock()
	defer source.mutex.Unlock()
	source.snapshots[cursor] = consumer
}

func (source *ReindexSource) claim(cursor string) (*Consumer, bool) {
	source.mutex.Lock()
	defer source.mutex.Unlock()
	consumer, exists := source.snapshots[cursor]
	delete(source.snapshots, cursor)

	return consumer, exists
}

func (source *ReindexSource) Abort(ctx context.Context, cursor string) error {
	consumer, exists := source.claim(cursor)
	if !exists {
		return nil
	}

	return consumer.Close(ctx)
}
