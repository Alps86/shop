package postgres_test

import (
	"context"
	"reflect"
	"strconv"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) bundleSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine weitere Position tatsächlich hinzugefügt wird$`, current.addPosition)
	ctx.Step(`^der öffentliche Kaufsnapshot vor der Lieferung gesichert wird$`, current.captureOrder)
	ctx.Step(`^eine gemeinsame Lieferung für mehrere Positionen verarbeitet wird$`, current.bundleDelivery)
	ctx.Step(`^Einzelbestände und unveränderlicher Kauf übereinstimmen$`, current.bundleObserved)
}

func (current *world) addPosition(input *godog.Table) error {
	row := table(input)
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	current.cart, err = current.service.ApplyChange(context.Background(), application.ChangeCommand{UserID: current.guest.UserID, CartID: current.cart.Cart.ID, CommandID: value.CommandID(row["CommandID"]), ExpectedRevision: current.cart.Cart.Revision, Operation: "add", SKU: value.SKU(row["SKU"]), Quantity: quantity})

	return err
}

func (current *world) captureOrder() error {
	var err error
	current.originalOrder, err = current.store.Orders().Get(context.Background(), current.result.OrderID)

	return err
}

func (current *world) bundleDelivery(input *godog.Table) error {
	row := table(input)
	lines, err := current.fulfillmentLines(input)
	if err != nil {
		return err
	}

	update := sale.FulfillmentUpdate{OrderID: current.result.OrderID, CommandID: value.CommandID(row["CommandID"]), ShipmentID: value.ShipmentID(row["ShipmentID"]), Status: sale.DeliveryStatus(row["Status"]), Lines: lines}
	_, current.err = (application.FulfillmentService{Store: current.store.Fulfillment()}).Report(context.Background(), application.DeliveryCommand{OrderID: update.OrderID, CommandID: update.CommandID, Update: update})

	return nil
}

func (current *world) fulfillmentLines(input *godog.Table) ([]sale.FulfillmentLine, error) {
	lines := []sale.FulfillmentLine{}
	for _, row := range current.tableRows(input) {
		quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
		if err != nil {
			return nil, err
		}

		lines = append(lines, sale.FulfillmentLine{SKU: value.SKU(row["SKU"]), Quantity: quantity})
	}

	return lines, nil
}

func (current *world) bundleObserved(input *godog.Table) error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	order, err := current.store.Orders().Get(context.Background(), current.result.OrderID)
	if err != nil {
		return err
	}

	for _, row := range current.tableRows(input) {
		actual := current.inventoryFields(snapshot.Inventory, row["SKU"])
		actual["OrderUnchanged"] = strconv.FormatBool(reflect.DeepEqual(order, current.originalOrder))
		actual["Code"] = value.Code(current.err)
		if err := check(row, actual); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) inventoryFields(inventories []sale.Inventory, sku string) map[string]string {
	for _, inventory := range inventories {
		if string(inventory.SKU) == sku {
			return map[string]string{"SKU": sku, "Stock": strconv.FormatInt(inventory.StockQuantity, 10), "Reserved": strconv.FormatInt(inventory.ReservedQuantity(), 10)}
		}
	}

	return map[string]string{}
}

func (current *world) tableRows(input *godog.Table) []map[string]string {
	rows := []map[string]string{}
	for _, record := range input.Rows[1:] {
		row := map[string]string{}
		for index, cell := range record.Cells {
			row[input.Rows[0].Cells[index].Value] = cell.Value
		}

		rows = append(rows, row)
	}

	return rows
}
