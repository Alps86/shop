package postgres

import (
	"context"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (repository GuestRepository) ForUser(ctx context.Context, id value.UserID) ([]sale.CategoryInterest, error) {
	user := sale.User{}
	err := repository.store.pool.QueryRow(ctx, "SELECT payload FROM users WHERE id=$1", id).Scan(&user)

	return user.CategoryInterests, databaseError(err)
}
