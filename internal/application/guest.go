package application

import (
	"context"
	"time"

	"github.com/Alps86/shop/internal/domain/value"
)

func (service Service) identity(ctx context.Context, session Session) (SessionIdentity, error) {
	identity, err := service.Guests.Resolve(ctx, session)
	if err != nil {
		return SessionIdentity{}, err
	}

	if !service.Clock.Now().Before(identity.IssuedAt.Add(30 * 24 * time.Hour)) {
		return SessionIdentity{}, value.NewError("NotFound", "Session")
	}

	return identity, nil
}

func (service Service) Open(ctx context.Context, session Session) (GuestView, error) {
	if session == "" {
		return service.Guests.Open(ctx)
	}

	identity, err := service.identity(ctx, session)
	if err != nil {
		return GuestView{}, err
	}

	return GuestView{Session: session, UserID: identity.UserID, CartID: identity.CartID, Profile: identity.Profile}, nil
}

func (service Service) SelectProfile(ctx context.Context, session Session, profile ProfileCode) (GuestView, error) {
	identity, err := service.identity(ctx, session)
	if err != nil {
		return GuestView{}, err
	}

	if !profile.valid() {
		return GuestView{}, value.NewError("InvalidInput", "Profile")
	}

	return service.Guests.Profile(ctx, identity.UserID, profile)
}

func (profile ProfileCode) valid() bool {
	valid := map[ProfileCode]bool{"neutral": true, "mobile": true, "home": true, "coffee": true}
	return valid[profile]
}
