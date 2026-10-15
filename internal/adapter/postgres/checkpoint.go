package postgres

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (store *Store) Seen(ctx context.Context, consumer string, envelope event.ProjectionEnvelope) (bool, error) {
	var seen bool
	err := store.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM cdc_events WHERE consumer=$1 AND event_id=$2)", consumer, envelope.EventID).Scan(&seen)

	return seen, databaseError(err)
}

func (store *Store) Acknowledge(ctx context.Context, consumer, lsn string, envelopes []event.ProjectionEnvelope) error {
	return store.transact(ctx, func(tx pgx.Tx) error {
		for _, envelope := range envelopes {
			if _, err := tx.Exec(ctx, "INSERT INTO cdc_events(consumer,event_id) VALUES($1,$2) ON CONFLICT DO NOTHING", consumer, envelope.EventID); err != nil {
				return err
			}
		}

		_, err := tx.Exec(ctx, "INSERT INTO cdc_checkpoints(consumer,lsn) VALUES($1,$2::pg_lsn) ON CONFLICT(consumer) DO UPDATE SET lsn=GREATEST(cdc_checkpoints.lsn,excluded.lsn)", consumer, lsn)

		return err
	})
}

func (store *Store) Checkpoint(ctx context.Context, consumer string) (uint64, error) {
	var lsn uint64
	err := store.pool.QueryRow(ctx, "SELECT COALESCE((SELECT lsn-'0/0'::pg_lsn FROM cdc_checkpoints WHERE consumer=$1),0)", consumer).Scan(&lsn)

	return lsn, databaseError(err)
}

func (store *Store) WALBarrier(ctx context.Context) (uint64, error) {
	var lsn uint64
	err := store.pool.QueryRow(ctx, "SELECT pg_logical_emit_message(false,'shop_reindex','')-'0/0'::pg_lsn").Scan(&lsn)

	return lsn, databaseError(err)
}

func (store *Store) ReplicationSlot(ctx context.Context, name string) (ReplicationState, error) {
	state := ReplicationState{Exists: true}
	err := store.pool.QueryRow(ctx, "SELECT COALESCE(restart_lsn-'0/0'::pg_lsn,0),COALESCE(confirmed_flush_lsn-'0/0'::pg_lsn,0),COALESCE(wal_status,'reserved') FROM pg_replication_slots WHERE slot_name=$1", name).Scan(&state.Restart, &state.Confirmed, &state.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReplicationState{}, nil
	}

	return state, databaseError(err)
}

func (store *Store) RegisterEnvelope(ctx context.Context, envelope event.ProjectionEnvelope) error {
	var id string
	query := `INSERT INTO cdc_event_registry(event_id,envelope) VALUES($1,$2) ON CONFLICT(event_id) DO UPDATE SET event_id=excluded.event_id WHERE cdc_event_registry.envelope=excluded.envelope RETURNING event_id`
	err := store.pool.QueryRow(ctx, query, envelope.EventID, envelope).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return value.NewError("Conflict", "EventID")
	}

	return databaseError(err)
}
