package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) Now() time.Time {
	if !current.clockNow.IsZero() {
		return current.clockNow
	}

	return time.Now().UTC()
}

func (current *world) configureCheckout(table *godog.Table) error {
	for _, row := range current.rows(table) {
		quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
		if err != nil {
			return err
		}

		address := sale.Address{Recipient: row["Recipient"], Street: row["Street"], PostalCode: row["PostalCode"], City: row["City"], Country: value.CountryCode(row["Country"])}
		command := application.ChangeCommand{CartID: current.view.Cart.ID, ExpectedRevision: current.view.Cart.Revision, CommandID: value.CommandID(row["CommandID"]), Operation: row["Operation"], SKU: value.SKU(row["SKU"]), Quantity: quantity, Address: address, ShippingMethod: sale.ShippingMethod(row["ShippingMethod"]), PaymentMethod: sale.PaymentMethod(row["PaymentMethod"])}
		current.view, current.err = current.service.Change(context.Background(), current.guest.Session, command)
		if current.err != nil {
			return current.err
		}
	}

	return nil
}

func (current *world) rows(table *godog.Table) []map[string]string {
	rows := []map[string]string{}
	for _, row := range table.Rows[1:] {
		values := map[string]string{}
		for index, cell := range row.Cells {
			values[table.Rows[0].Cells[index].Value] = cell.Value
		}

		rows = append(rows, values)
	}

	return rows
}

func (current *world) buy(table *godog.Table) error {
	row := current.rows(table)[0]
	current.view, current.err = current.service.Confirm(context.Background(), current.guest.Session, application.ConfirmCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["ConfirmCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID, ExpectedRevision: current.view.Cart.Revision})
	if current.err != nil {
		return current.err
	}

	current.purchase, current.err = current.service.Place(context.Background(), current.guest.Session, application.PlaceCommand{CartID: current.view.Cart.ID, CommandID: value.CommandID(row["PlaceCommandID"]), EvaluationID: current.view.Cart.Evaluation.Value.ID})
	if current.err != nil {
		return current.err
	}

	return current.capturePurchase()
}

func (current *world) capturePurchase() error {
	order, err := current.service.GetOrder(context.Background(), current.guest.Session, current.purchase.OrderID)
	if err != nil {
		return err
	}

	current.initialPurchase = order.Purchase

	return (postgres.ProcessDispatcher{Store: current.databaseStore, Processes: current.adapter}).Dispatch(context.Background())
}

func (current *world) purchaseObserved(table *godog.Table) error {
	page, err := current.databaseStore.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	snapshot, err := current.databaseStore.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	actual := map[string]string{"OrderCount": strconv.FormatInt(page.Total, 10), "Reserved": strconv.FormatInt(snapshot.Inventory[0].ReservedQuantity(), 10), "Total": strconv.FormatInt(page.Orders[0].Purchase.Total.MinorUnits, 10), "RuleSetVersion": page.Orders[0].Purchase.RuleSetVersion, "Code": value.Code(current.err)}
	for field, wanted := range current.rows(table)[0] {
		if actual[field] != wanted {
			return fmt.Errorf("%s: %s statt %s", field, actual[field], wanted)
		}
	}

	return current.orderExists()
}

func (current *world) orderExists() error {
	encoding, err := current.connection.QueryWorkflow(context.Background(), "order/"+string(current.purchase.OrderID), "", "ProcessStatus")
	if err != nil {
		return err
	}

	var status temporal.Status
	if err := encoding.Get(&status); err != nil {
		return err
	}

	expected := application.ProcessBinding{DefinitionVersion: current.initialPurchase.DefinitionVersion, RuleSetVersion: current.initialPurchase.RuleSetVersion}
	if status.Binding != expected {
		return fmt.Errorf("Orderprozess wechselte gebundenen Regelstand")
	}

	return nil
}
