package postgres

import (
	"context"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository InventoryRepository) Change(ctx context.Context, input application.InventoryCommand) (sale.Inventory, error) {
	result := sale.Inventory{}
	command := commandRecord{scope: "inventory/" + string(input.SKU), id: input.CommandID, payload: input, result: &result}
	err := repository.store.transact(ctx, func(tx pgx.Tx) error {
		return repository.change(ctx, tx, input, command)
	})

	return result, err
}

func (repository InventoryRepository) change(ctx context.Context, tx pgx.Tx, input application.InventoryCommand, command commandRecord) error {
	inventory := sale.Inventory{}
	if err := newRecord("inventory", string(input.SKU), &inventory).lock(ctx, tx); err != nil {
		return err
	}

	found, err := command.previous(ctx, tx)
	if found {
		return err
	}

	if err != nil {
		return err
	}

	return repository.apply(ctx, tx, inventory, command)
}

func (repository InventoryRepository) apply(ctx context.Context, tx pgx.Tx, inventory sale.Inventory, command commandRecord) error {
	updated, err := repository.mutate(inventory, command.payload.(application.InventoryCommand))
	if err != nil {
		return err
	}

	if err := newRecord("inventory", string(updated.SKU), updated).save(ctx, tx); err != nil {
		return err
	}

	if err := repository.store.projectSKU(ctx, tx, updated.SKU); err != nil {
		return err
	}

	*command.result.(*sale.Inventory) = updated

	return command.save(ctx, tx)
}

func (repository InventoryRepository) mutate(inventory sale.Inventory, input application.InventoryCommand) (sale.Inventory, error) {
	if inventory.Revision != input.ExpectedRevision {
		return sale.Inventory{}, value.NewError("Conflict", "Revision")
	}

	if input.Price.Present {
		updated, err := inventory.SetPrice(input.Price.Value, inventory.Revision)
		if err != nil {
			return sale.Inventory{}, err
		}

		inventory = updated
	}

	if input.Stock.Present {
		return inventory.SetStock(input.Stock.Value, inventory.Revision)
	}

	return inventory, nil
}
