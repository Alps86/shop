package postgres

import (
	"context"
	"slices"
	"strings"

	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository FulfillmentRepository) Get(ctx context.Context, id value.OrderID) (sale.Fulfillment, error) {
	var result sale.Fulfillment
	err := repository.store.pool.QueryRow(ctx, "SELECT fulfillment FROM orders WHERE id=$1", id).Scan(&result)

	return result, databaseError(err)
}

func (repository FulfillmentRepository) Save(ctx context.Context, fulfillment sale.Fulfillment, id value.CommandID) error {
	update, err := repository.updateFor(fulfillment, id)
	if err != nil {
		return err
	}

	return repository.store.transact(ctx, func(tx pgx.Tx) error {
		return repository.save(ctx, tx, update)
	})
}

func (repository FulfillmentRepository) updateFor(fulfillment sale.Fulfillment, id value.CommandID) (sale.FulfillmentUpdate, error) {
	for _, update := range fulfillment.Updates {
		if update.CommandID == id {
			return update, nil
		}
	}

	return sale.FulfillmentUpdate{}, value.NewError("InvalidInput", "CommandID")
}

func (repository FulfillmentRepository) save(ctx context.Context, tx pgx.Tx, update sale.FulfillmentUpdate) error {
	var current sale.Fulfillment
	if err := tx.QueryRow(ctx, "SELECT fulfillment FROM orders WHERE id=$1 FOR UPDATE", update.OrderID).Scan(&current); err != nil {
		return databaseError(err)
	}

	result := sale.Fulfillment{}
	command := commandRecord{scope: "delivery/" + string(update.OrderID), id: update.CommandID, payload: update, result: &result}
	found, err := command.previous(ctx, tx)
	if found {
		return err
	}

	if err != nil {
		return err
	}

	return repository.apply(ctx, tx, current, command)
}

func (repository FulfillmentRepository) apply(ctx context.Context, tx pgx.Tx, current sale.Fulfillment, command commandRecord) error {
	update := command.payload.(sale.FulfillmentUpdate)
	updated, err := current.Apply(update)
	if err != nil {
		return err
	}

	if update.Status == sale.Shipped {
		if err := repository.consume(ctx, tx, update); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(ctx, "UPDATE orders SET fulfillment=$1 WHERE id=$2", updated, updated.OrderID); err != nil {
		return databaseError(err)
	}

	*command.result.(*sale.Fulfillment) = updated

	return command.save(ctx, tx)
}

func (repository FulfillmentRepository) consume(ctx context.Context, tx pgx.Tx, update sale.FulfillmentUpdate) error {
	lines := append([]sale.FulfillmentLine{}, update.Lines...)
	slices.SortFunc(lines, func(left, right sale.FulfillmentLine) int {
		return strings.Compare(string(left.SKU), string(right.SKU))
	})

	for _, line := range lines {
		if err := repository.consumeLine(ctx, tx, update.OrderID, line); err != nil {
			return err
		}
	}

	return nil
}

func (repository FulfillmentRepository) consumeLine(ctx context.Context, tx pgx.Tx, order value.OrderID, line sale.FulfillmentLine) error {
	inventory := sale.Inventory{}
	record := newRecord("inventory", string(line.SKU), &inventory)
	if err := record.lock(ctx, tx); err != nil {
		return err
	}

	updated, err := inventory.Ship(order, line.Quantity, inventory.Revision)
	if err != nil {
		return err
	}

	if err := newRecord("inventory", string(line.SKU), updated).save(ctx, tx); err != nil {
		return err
	}

	return repository.store.projectSKU(ctx, tx, line.SKU)
}
