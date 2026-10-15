package postgres

import (
	"context"

	"github.com/Alps86/shop/internal/application"
)

func (repository CartRepository) ReplayChange(ctx context.Context, input application.ChangeCommand) (application.CartView, bool, error) {
	result := application.CartView{}
	command := commandRecord{scope: "cart/" + string(input.CartID), id: input.CommandID, payload: input, result: &result}
	found, err := command.previous(ctx, repository.store.pool)

	return result, found, err
}

func (repository CartRepository) ReplayConfirm(ctx context.Context, input application.ConfirmCommand) (application.CartView, bool, error) {
	result := application.CartView{}
	command := commandRecord{scope: "cart/" + string(input.CartID), id: input.CommandID, payload: input, result: &result}
	found, err := command.previous(ctx, repository.store.pool)

	return result, found, err
}
