package postgres

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/jackc/pgx/v5"
)

func (repository CartRepository) Get(ctx context.Context, id value.CartID) (sale.Cart, error) {
	return repository.load(ctx, repository.store.pool, id, false)
}

func (repository CartRepository) load(ctx context.Context, db database, id value.CartID, lock bool) (sale.Cart, error) {
	query := "SELECT payload FROM carts WHERE id=$1"
	if lock {
		query += " FOR UPDATE"
	}

	var cart sale.Cart
	err := db.QueryRow(ctx, query, id).Scan(&cart)

	return cart, databaseError(err)
}

func (repository CartRepository) Start(ctx context.Context, user value.UserID, id value.CommandID, now time.Time) (application.CartView, error) {
	result := application.CartView{}
	command := commandRecord{scope: "start/" + string(user), id: id, payload: startCommand{UserID: user}, result: &result}
	err := repository.store.transact(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", user); err != nil {
			return err
		}

		return repository.start(ctx, tx, command, now)
	})

	return result, err
}

func (repository CartRepository) start(ctx context.Context, tx pgx.Tx, command commandRecord, now time.Time) error {
	found, err := command.previous(ctx, tx)
	if found {
		return err
	}

	if err != nil {
		return err
	}

	return repository.newCart(ctx, tx, command, now)
}

func (repository CartRepository) newCart(ctx context.Context, tx pgx.Tx, command commandRecord, now time.Time) error {
	cart, err := sale.NewCart(value.CartID("c-"+rand.Text()), command.payload.(startCommand).UserID, now)
	if err != nil {
		return err
	}

	var rules versionState
	if err := newRecord("active-ruleset", "active", &rules).read(ctx, tx); err != nil {
		return err
	}

	cart.RuleSetVersion = rules.Version

	return repository.insert(ctx, tx, cart, command)
}

func (repository CartRepository) insert(ctx context.Context, tx pgx.Tx, cart sale.Cart, command commandRecord) error {
	if _, err := tx.Exec(ctx, "INSERT INTO carts(id,user_id,revision,payload) VALUES($1,$2,$3,$4)", cart.ID, cart.UserID, cart.Revision, cart); err != nil {
		return databaseError(err)
	}

	*command.result.(*application.CartView) = application.NewCartView(cart)
	if err := repository.store.process(ctx, tx, "EnsureCart", cart); err != nil {
		return err
	}

	return command.save(ctx, tx)
}

func (repository CartRepository) Save(ctx context.Context, write application.CartWrite) (application.CartView, error) {
	result := application.NewCartView(write.Cart)
	command := commandRecord{scope: "cart/" + string(write.Cart.ID), id: write.CommandID, payload: repository.fingerprint(write), result: &result}
	err := repository.store.transact(ctx, func(tx pgx.Tx) error {
		return repository.write(ctx, tx, write, command)
	})

	return result, err
}

func (repository CartRepository) fingerprint(write application.CartWrite) any {
	if write.Change.Present {
		return write.Change.Value
	}

	if write.Confirmation.Present {
		return write.Confirmation.Value
	}

	return map[string]any{"CartID": write.Cart.ID, "EvaluationID": write.Cart.Evaluation.Value.ID}
}

func (repository CartRepository) write(ctx context.Context, tx pgx.Tx, write application.CartWrite, command commandRecord) error {
	previous, err := repository.load(ctx, tx, write.Cart.ID, true)
	if err != nil {
		return err
	}

	repeated, err := command.previous(ctx, tx)
	if repeated {
		return err
	}

	if err != nil {
		return err
	}

	if previous.Revision != write.ExpectedRevision {
		return value.NewError("Conflict", "Revision")
	}

	return repository.persistWrite(ctx, tx, write.Cart, command)
}

func (repository CartRepository) persistWrite(ctx context.Context, tx pgx.Tx, cart sale.Cart, command commandRecord) error {
	if err := repository.update(ctx, tx, cart); err != nil {
		return err
	}

	if err := repository.store.process(ctx, tx, "CartUpdated", cart); err != nil {
		return err
	}

	return command.save(ctx, tx)
}

func (repository CartRepository) update(ctx context.Context, tx pgx.Tx, cart sale.Cart) error {
	_, err := tx.Exec(ctx, "UPDATE carts SET revision=$1,payload=$2 WHERE id=$3", cart.Revision, cart, cart.ID)
	return databaseError(err)
}

func (repository CartRepository) Expire(ctx context.Context, id value.CartID, now time.Time) (sale.Cart, error) {
	result := sale.Cart{}
	err := repository.store.transact(ctx, func(tx pgx.Tx) error {
		cart, err := repository.load(ctx, tx, id, true)
		if err != nil {
			return err
		}

		result, err = cart.Expire(now)
		if err != nil {
			return err
		}

		return repository.update(ctx, tx, result)
	})

	return result, err
}
