package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
)

func (dispatcher ProcessDispatcher) Dispatch(ctx context.Context) error {
	return dispatcher.Store.transact(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT sequence,envelope FROM process_outbox WHERE NOT acknowledged ORDER BY sequence FOR UPDATE SKIP LOCKED")
		if err != nil {
			return err
		}

		pending, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (pendingProcess, error) {
			item := pendingProcess{}
			err := row.Scan(&item.sequence, &item.envelope)

			return item, err
		})
		if err != nil {
			return err
		}

		return dispatcher.deliver(ctx, tx, pending)
	})
}

func (dispatcher ProcessDispatcher) deliver(ctx context.Context, tx pgx.Tx, pending []pendingProcess) error {
	for _, item := range pending {
		if err := dispatcher.apply(ctx, item); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, "UPDATE process_outbox SET acknowledged=true WHERE sequence=$1", item.sequence); err != nil {
			return err
		}
	}

	return nil
}

func (dispatcher ProcessDispatcher) apply(ctx context.Context, item pendingProcess) error {
	if item.envelope.Kind == "EnsureOrder" {
		return dispatcher.Processes.EnsureOrder(ctx, item.envelope.OrderID, item.envelope.Binding)
	}

	return dispatcher.Processes.EnsureCart(ctx, item.envelope.CartID, item.envelope.Binding)
}
