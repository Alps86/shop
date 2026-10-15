package application

import (
	"context"
	"github.com/Alps86/shop/internal/domain/sale"
)

func (service Service) GetInterests(ctx context.Context, session Session) ([]sale.CategoryInterest, error) {
	identity, err := service.identity(ctx, session)
	if err != nil {
		return nil, err
	}

	return service.Interests.ForUser(ctx, identity.UserID)
}
