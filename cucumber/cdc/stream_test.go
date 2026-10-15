package cdc_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
	"time"
)

func (current *world) snapshotReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	current.snapshot, current.err = current.source.Snapshot(ctx, time.Now())

	return current.err
}

func (current *world) commitEvent() error {
	envelope := event.ProjectionEnvelope{EventID: "e-delete", SchemaVersion: 1, OccurredAt: time.Now().UTC(), Kind: "delete", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, Projection: application.ProjectionEvent{ProductID: "p-phone", Revision: 2}}
	_, err := current.admin.Exec(context.Background(), "INSERT INTO "+pgx.Identifier{current.config.Schema, "outbox"}.Sanitize()+"(envelope) VALUES($1)", envelope)

	return err
}

func (current *world) catchup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	current.err = current.source.Catchup(ctx, current.snapshot.Cursor, current)
}

func (current *world) Apply(_ context.Context, projection application.ProjectionEvent) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	if !current.available {
		if current.attempted != nil {
			current.attempted <- struct{}{}
		}

		return value.NewError("DependencyUnavailable", "Projection")
	}

	current.events = append(current.events, projection)
	if current.projected != nil {
		current.projected <- struct{}{}
	}

	return nil
}

func (current *world) Rebuild(context.Context, string, time.Time) error {
	return value.NewError("InvalidTransition", "Rebuild")
}
