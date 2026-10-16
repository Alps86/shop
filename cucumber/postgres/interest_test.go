package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strconv"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
)

func (current *world) sessionInterests(row map[string]string) error {
	session := current.guest.Session
	if row["Session"] == "unknown" {
		session = "unknown"
	}

	if row["Session"] == "previous" {
		session = current.previousGuest.Session
	}

	snapshot, err := current.store.Guests().ForSession(context.Background(), session)
	current.observeSessionInterests(snapshot, err)

	return nil
}

func (current *world) observeSessionInterests(snapshot application.SessionInterests, err error) {
	current.err = err
	current.observed["Code"] = value.Code(err)
	current.observed["PayloadCode"] = value.Code(snapshot.PayloadError)
	current.observed["Count"] = strconv.Itoa(len(snapshot.Interests))
	if len(snapshot.Interests) > 0 {
		current.observed["CategoryID"] = string(snapshot.Interests[0].CategoryID)
		current.observed["Weight"] = string(snapshot.Interests[0].Weight)
	}
}

func (current *world) interestsWithoutCart(input *godog.Table) (err error) {
	role := pgx.Identifier{current.schema + "_interests"}.Sanitize()
	if _, err := current.admin.Exec(context.Background(), "CREATE ROLE "+role+" NOLOGIN; GRANT USAGE ON SCHEMA "+pgx.Identifier{current.schema}.Sanitize()+" TO "+role+"; GRANT SELECT ON "+pgx.Identifier{current.schema}.Sanitize()+".sessions,"+pgx.Identifier{current.schema}.Sanitize()+".users TO "+role); err != nil {
		return err
	}

	defer func() {
		_, cleanup := current.admin.Exec(context.Background(), "DROP OWNED BY "+role+"; DROP ROLE "+role)
		err = errors.Join(err, cleanup)
	}()

	return current.restrictedInterests(table(input))
}

func (current *world) restrictedInterests(row map[string]string) error {
	store, err := current.interestStore()
	if err != nil {
		return err
	}

	defer store.Close()
	snapshot, err := store.Guests().ForSession(context.Background(), current.guest.Session)
	current.observed = map[string]string{}
	current.observeSessionInterests(snapshot, err)
	_, resolveErr := store.Guests().Resolve(context.Background(), current.guest.Session)
	current.observed["ResolveCode"] = value.Code(resolveErr)

	return check(row, current.observed)
}

func (current *world) interestStore() (*postgres.Store, error) {
	dsn, err := url.Parse(os.Getenv("SHOP_TEST_POSTGRES_DSN"))
	if err != nil {
		return nil, err
	}

	query := dsn.Query()
	query.Set("options", "-crole="+current.schema+"_interests")
	dsn.RawQuery = query.Encode()

	return postgres.Open(context.Background(), postgres.Config{DSN: dsn.String(), Schema: current.schema})
}

func (current *world) secondGuest() error {
	current.previousGuest = current.guest

	return current.openGuest()
}
