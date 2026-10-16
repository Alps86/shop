package postgres

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (repository GuestRepository) ForUser(ctx context.Context, id value.UserID) ([]sale.CategoryInterest, error) {
	user := sale.User{}
	err := repository.store.pool.QueryRow(ctx, "SELECT payload FROM users WHERE id=$1", id).Scan(&user)

	return user.CategoryInterests, databaseError(err)
}

func (repository GuestRepository) ForSession(ctx context.Context, session application.Session) (application.SessionInterests, error) {
	snapshot := application.SessionInterests{}
	var payload []byte
	err := repository.store.pool.QueryRow(ctx, "SELECT s.issued_at,u.payload FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.issued_at>now()-interval '30 days'", repository.hash(string(session))).Scan(&snapshot.IssuedAt, &payload)
	if err != nil {
		return application.SessionInterests{}, databaseError(err)
	}

	user := sale.User{}
	snapshot.PayloadError = databaseError(json.Unmarshal(payload, &user))
	if snapshot.PayloadError == nil {
		snapshot.Interests = user.CategoryInterests
	}

	return snapshot, nil
}
