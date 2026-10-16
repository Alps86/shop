package application

import (
	"context"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

func (service Service) GetInterests(ctx context.Context, session Session) ([]sale.CategoryInterest, error) {
	snapshot, err := service.Interests.ForSession(ctx, session)
	if err != nil {
		return nil, err
	}

	if !service.Clock.Now().Before(snapshot.IssuedAt.Add(30 * 24 * time.Hour)) {
		return nil, value.NewError("NotFound", "Session")
	}

	if snapshot.PayloadError != nil {
		return nil, snapshot.PayloadError
	}

	return snapshot.Interests, nil
}
