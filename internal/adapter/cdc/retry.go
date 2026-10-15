package cdc

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

func (consumer *Consumer) Run(ctx context.Context) error {
	for {
		err := errors.Join(consumer.run(ctx), consumer.Close(context.WithoutCancel(ctx)))
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !consumer.recoverable(err) {
			return err
		}

		if err := consumer.pause(ctx); err != nil {
			return err
		}
	}
}

func (consumer *Consumer) recoverable(err error) bool {
	recoverable := map[string]bool{"DependencyUnavailable": true, "InternalError": true}
	return recoverable[value.Code(consumer.classify(err))]
}

func (consumer *Consumer) classify(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		if postgresError.Code == "55000" {
			return errors.Join(value.NewError("ReindexRequired", "ReplicationSlot"), err)
		}
	}

	return err
}

func (consumer *Consumer) pause(ctx context.Context) error {
	consumer.retryDelay = min(30*time.Second, max(time.Second, 2*consumer.retryDelay))
	timer := time.NewTimer(consumer.retryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
