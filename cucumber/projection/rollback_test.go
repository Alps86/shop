package projection_test

import (
	"context"
	"strconv"

	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
)

func (current *world) rollbackInventory(table *godog.Table) error {
	if err := current.Admin.QueryRow(context.Background(), "SELECT count(*) FROM "+pgx.Identifier{current.Config.Schema, "outbox"}.Sanitize()).Scan(&current.PreviousOutbox); err != nil {
		return err
	}

	if err := current.failProjectionWrite(); err != nil {
		return err
	}

	current.IntegrationError = current.changeInventory(current.rows(table)[0])

	return nil
}

func (current *world) failProjectionWrite() error {
	function := pgx.Identifier{current.Config.Schema, "contract_rollback"}.Sanitize()
	statement := "CREATE FUNCTION " + function + "() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'gezielter Outboxfehler'; END $$"
	if _, err := current.Admin.Exec(context.Background(), statement); err != nil {
		return err
	}

	_, err := current.Admin.Exec(context.Background(), "CREATE TRIGGER contract_rollback BEFORE INSERT ON "+pgx.Identifier{current.Config.Schema, "outbox"}.Sanitize()+" FOR EACH ROW EXECUTE FUNCTION "+function+"()")

	return err
}

func (current *world) rollbackObserved(table *godog.Table) error {
	snapshot, err := current.Store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	var events int64
	if err := current.Admin.QueryRow(context.Background(), "SELECT count(*) FROM "+pgx.Identifier{current.Config.Schema, "outbox"}.Sanitize()).Scan(&events); err != nil {
		return err
	}

	inventory := snapshot.Inventory[0]

	return current.matches(map[string]string{"Code": value.Code(current.IntegrationError), "Price": strconv.FormatInt(inventory.Price.MinorUnits, 10), "Stock": strconv.FormatInt(inventory.StockQuantity, 10), "OutboxDelta": strconv.FormatInt(events-current.PreviousOutbox, 10)}, current.rows(table)[0])
}
