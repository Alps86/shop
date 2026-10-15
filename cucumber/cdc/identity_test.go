package cdc_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"time"
)

func (current *world) batchEnvelope(input *godog.Table) error {
	envelope := event.ProjectionEnvelope{EventID: "e-identity", SchemaVersion: 1, OccurredAt: time.Now().UTC(), Kind: "delete", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, Projection: application.ProjectionEvent{ProductID: "p-phone", Revision: 2}}
	repeated := envelope
	if current.fields(input)["Payload"] == "different" {
		repeated.Projection.Revision++
	}

	_, err := current.admin.Exec(context.Background(), "INSERT INTO "+pgx.Identifier{current.config.Schema, "outbox"}.Sanitize()+"(envelope) VALUES($1),($2)", envelope, repeated)

	return err
}

func (current *world) repeatEnvelope(input *godog.Table) error {
	var err error
	current.priorCheckpoint, err = current.store.Checkpoint(context.Background(), current.config.Slot)
	if err != nil {
		return err
	}

	table := pgx.Identifier{current.config.Schema, "outbox"}.Sanitize()
	expression := "envelope"
	if current.fields(input)["Payload"] == "different" {
		expression = "jsonb_set(envelope,'{Projection,Revision}','3')"
	}

	_, err = current.admin.Exec(context.Background(), "INSERT INTO "+table+"(envelope) SELECT "+expression+" FROM "+table+" ORDER BY sequence LIMIT 1")

	return err
}

func (current *world) stopAfterCheckpoint() error {
	if err := current.waitPosition(current.priorCheckpoint); err != nil {
		current.cancel()
		return err
	}

	current.cancel()
	current.err = <-current.stopped
	if errors.Is(current.err, context.Canceled) {
		current.err = nil
	}

	return errors.Join(current.err, current.closeStream())
}
