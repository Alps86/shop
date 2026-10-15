package postgres_test

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/cucumber/godog"
)

func (current *world) blockedWrite(input *godog.Table) error {
	row := table(input)
	if row["Boundary"] == "StartCart" {
		return current.blockedStart()
	}

	if _, err := current.admin.Exec(context.Background(), "SELECT pg_advisory_lock(741902)"); err != nil {
		return err
	}

	err := current.timedWrite(current.cancelImport)
	_, unlockErr := current.admin.Exec(context.Background(), "SELECT pg_advisory_unlock(741902)")

	return errors.Join(err, unlockErr)
}

func (current *world) blockedStart() error {
	if _, err := current.admin.Exec(context.Background(), "SELECT pg_advisory_lock(hashtextextended($1,0))", current.guest.UserID); err != nil {
		return err
	}

	err := current.timedWrite(current.cancelStart)
	_, unlockErr := current.admin.Exec(context.Background(), "SELECT pg_advisory_unlock(hashtextextended($1,0))", current.guest.UserID)

	return errors.Join(err, unlockErr)
}

func (current *world) timedWrite(operation func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	current.err = operation(ctx)
	current.observed = map[string]string{"Deadline": strconv.FormatBool(errors.Is(current.err, context.DeadlineExceeded))}

	return nil
}
