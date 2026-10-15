package projection_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"os"
	"strconv"
)

func (current *world) publish(table *godog.Table) error {
	path := current.rows(table)[0]["Fixture"]
	root, err := os.OpenRoot("../..")
	if err != nil {
		return err
	}

	payload, readErr := root.ReadFile(path)
	if err := errors.Join(readErr, root.Close()); err != nil {
		return err
	}

	if err := json.Unmarshal(payload, &current.Manifest); err != nil {
		return err
	}

	_, err = current.Store.Import(context.Background(), current.Manifest)

	return err
}

func (current *world) inventoryChange(table *godog.Table) error {
	for _, row := range current.rows(table) {
		if err := current.changeInventory(row); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) changeInventory(row map[string]string) error {
	price, err := strconv.ParseInt(row["Price"], 10, 64)
	if err != nil {
		return err
	}

	stock, err := strconv.ParseInt(row["Stock"], 10, 64)
	if err != nil {
		return err
	}

	revision, err := strconv.ParseUint(row["ExpectedRevision"], 10, 64)
	if err != nil {
		return err
	}

	_, err = current.Store.Inventory().Change(context.Background(), application.InventoryCommand{SKU: value.SKU(row["SKU"]), ExpectedRevision: revision, CommandID: value.CommandID(row["CommandID"]), Price: value.Optional[value.Money]{Present: true, Value: value.Money{MinorUnits: price, Currency: "EUR"}}, Stock: value.Optional[int64]{Present: true, Value: stock}})

	return err
}
