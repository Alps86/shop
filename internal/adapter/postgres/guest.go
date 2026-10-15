package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository GuestRepository) Open(ctx context.Context) (application.GuestView, error) {
	token, err := repository.token()
	if err != nil {
		return application.GuestView{}, err
	}

	user := sale.User{ID: value.UserID("u-" + rand.Text())}
	view := application.GuestView{UserID: user.ID, Profile: "neutral", Session: application.Session(token)}
	err = repository.store.transact(ctx, func(tx pgx.Tx) error {
		return repository.insert(ctx, tx, user, token)
	})

	return view, err
}

func (repository GuestRepository) token() (string, error) {
	entropy := repository.store.config.Entropy
	if entropy == nil {
		entropy = rand.Reader
	}

	bytes := make([]byte, 32)
	if _, err := io.ReadFull(entropy, bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (repository GuestRepository) insert(ctx context.Context, tx pgx.Tx, user sale.User, token string) error {
	if _, err := tx.Exec(ctx, "INSERT INTO users(id,profile,payload) VALUES($1,'neutral',$2)", user.ID, user); err != nil {
		return databaseError(err)
	}

	_, err := tx.Exec(ctx, "INSERT INTO sessions(token_hash,user_id,issued_at) VALUES($1,$2,$3)", repository.hash(token), user.ID, time.Now().UTC())

	return databaseError(err)
}

func (repository GuestRepository) hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (repository GuestRepository) Resolve(ctx context.Context, session application.Session) (application.SessionIdentity, error) {
	var identity application.SessionIdentity
	err := repository.store.pool.QueryRow(ctx, "SELECT s.user_id,s.issued_at,u.profile FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.issued_at>now()-interval '30 days'", repository.hash(string(session))).Scan(&identity.UserID, &identity.IssuedAt, &identity.Profile)
	if err != nil {
		return application.SessionIdentity{}, databaseError(err)
	}

	cart, err := repository.activeCart(ctx, identity.UserID)
	identity.CartID = cart

	return identity, err
}

func (repository GuestRepository) activeCart(ctx context.Context, user value.UserID) (value.Optional[value.CartID], error) {
	var id value.CartID
	err := repository.store.pool.QueryRow(ctx, "SELECT id FROM carts WHERE user_id=$1 AND payload->>'State' IN ('open','checkingOut') ORDER BY payload->>'LastActivityAt' DESC LIMIT 1", user).Scan(&id)
	if value.Code(databaseError(err)) == "NotFound" {
		return value.Optional[value.CartID]{}, nil
	}

	return value.Optional[value.CartID]{Value: id, Present: err == nil}, databaseError(err)
}

func (repository GuestRepository) Profile(ctx context.Context, user value.UserID, profile application.ProfileCode) (application.GuestView, error) {
	interests := map[application.ProfileCode][]sale.CategoryInterest{"neutral": {}, "mobile": {{CategoryID: "mobile", Weight: "1"}}, "home": {{CategoryID: "home", Weight: "1"}}, "coffee": {{CategoryID: "coffee", Weight: "1"}}}
	selected, exists := interests[profile]
	if !exists {
		return application.GuestView{}, value.NewError("InvalidInput", "Profile")
	}

	validated, err := repository.profileUser(ctx, user, selected)
	if err != nil {
		return application.GuestView{}, err
	}

	tag, err := repository.store.pool.Exec(ctx, "UPDATE users SET profile=$1,payload=$2 WHERE id=$3", profile, validated, user)
	if err != nil {
		return application.GuestView{}, databaseError(err)
	}

	return repository.profileResult(user, profile, tag.RowsAffected())
}

func (repository GuestRepository) profileResult(user value.UserID, profile application.ProfileCode, changed int64) (application.GuestView, error) {
	if changed == 0 {
		return application.GuestView{}, value.NewError("NotFound", "UserID")
	}

	return application.GuestView{UserID: user, Profile: profile}, nil
}

func (repository GuestRepository) profileUser(ctx context.Context, id value.UserID, interests []sale.CategoryInterest) (sale.User, error) {
	categories := []catalog.Category{}
	if err := repository.store.documents(ctx, "category", &categories); err != nil {
		return sale.User{}, err
	}

	identifiers := []value.CategoryID{}
	for _, category := range categories {
		identifiers = append(identifiers, category.ID)
	}

	user := sale.User{ID: id, CategoryInterests: interests}

	return user, user.Validate(identifiers)
}
