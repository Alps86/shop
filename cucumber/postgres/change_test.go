package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) changeSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^der Bestand atomar geändert wird$`, current.inventoryChange)
	ctx.Step(`^lautet die persistente Zustandsbeobachtung$`, current.stateObserved)
	ctx.Step(`^eine bestätigte Lieferung dauerhaft verarbeitet wird$`, current.delivery)
	ctx.Step(`^die dauerhafte Liefertransaktionsgrenze direkt aufgerufen wird$`, current.directDelivery)
	ctx.Step(`^derselbe Mengenbefehl nach dem Commit erneut ausgeführt wird$`, current.replayChange)
	ctx.Step(`^dieselbe Bestätigung nach dem Commit erneut ausgeführt wird$`, current.replayConfirm)
}

func (current *world) inventoryChange(input *godog.Table) error {
	row := table(input)
	price, err := strconv.ParseInt(row["Price"], 10, 64)
	if err != nil {
		return err
	}

	stock, err := strconv.ParseInt(row["Stock"], 10, 64)
	if err != nil {
		return err
	}

	revision, err := strconv.ParseUint(row["Revision"], 10, 64)
	if err != nil {
		return err
	}

	command := application.InventoryCommand{SKU: value.SKU(row["SKU"]), CommandID: value.CommandID(row["CommandID"]), ExpectedRevision: revision, Price: value.Optional[value.Money]{Present: row["SetPrice"] == "true", Value: value.Money{MinorUnits: price, Currency: "EUR"}}, Stock: value.Optional[int64]{Present: row["SetStock"] == "true", Value: stock}}
	_, current.err = current.store.Inventory().Change(context.Background(), command)

	return nil
}

func (current *world) stateObserved(input *godog.Table) error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	page, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	var projectionRevision uint64
	if err := current.admin.QueryRow(context.Background(), "SELECT revision FROM "+current.schema+".projection_versions WHERE product_id='p-phone'").Scan(&projectionRevision); err != nil {
		return err
	}

	actual := map[string]string{"Code": value.Code(current.err), "Reserved": strconv.FormatInt(snapshot.Inventory[0].ReservedQuantity(), 10), "Stock": strconv.FormatInt(snapshot.Inventory[0].StockQuantity, 10), "OrderCount": strconv.FormatInt(page.Total, 10), "ProjectionRevision": strconv.FormatUint(projectionRevision, 10), "ProductRevision": strconv.FormatUint(snapshot.Products[0].Revision, 10), "ReturnedCart": strconv.FormatBool(current.result.Cart.Present), "Confirmation": strconv.FormatBool(current.result.Cart.Value.Cart.Checkout.Value.ConfirmedEvaluation.Present), "UnavailableSKU": string(current.result.UnavailableSKU.Value), "Alternatives": strconv.Itoa(len(current.result.Alternatives)), "AlternativeSKU": current.alternativeSKU()}

	return check(table(input), actual)
}

func (current *world) delivery(input *godog.Table) error {
	row := table(input)
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	update := sale.FulfillmentUpdate{OrderID: current.result.OrderID, ShipmentID: value.ShipmentID(row["ShipmentID"]), CommandID: value.CommandID(row["CommandID"]), Status: sale.DeliveryStatus(row["Status"]), Lines: []sale.FulfillmentLine{{SKU: value.SKU(row["SKU"]), Quantity: quantity}}}
	service := application.FulfillmentService{Store: current.store.Fulfillment()}
	_, current.err = service.Report(context.Background(), application.DeliveryCommand{OrderID: update.OrderID, CommandID: update.CommandID, Update: update})

	return nil
}

func (current *world) replayChange(input *godog.Table) error {
	row := table(input)
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	command := application.ChangeCommand{UserID: current.guest.UserID, CartID: current.cart.Cart.ID, CommandID: "add-a", ExpectedRevision: 1, Operation: "add", SKU: value.SKU(row["SKU"]), Quantity: quantity}
	current.cart, current.err = current.service.ApplyChange(context.Background(), command)

	return nil
}

func (current *world) replayConfirm(_ *godog.Table) {
	cart := current.cart.Cart
	command := application.ConfirmCommand{UserID: current.guest.UserID, CartID: cart.ID, CommandID: "confirm-a", EvaluationID: cart.Evaluation.Value.ID, ExpectedRevision: cart.Revision}
	current.cart, current.err = current.service.ApplyConfirm(context.Background(), command)
}

func (current *world) directDelivery(input *godog.Table) error {
	row := table(input)
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	update := sale.FulfillmentUpdate{OrderID: current.result.OrderID, CommandID: value.CommandID(row["CommandID"]), ShipmentID: value.ShipmentID(row["ShipmentID"]), Status: sale.DeliveryStatus(row["Status"]), Lines: []sale.FulfillmentLine{{SKU: value.SKU(row["SKU"]), Quantity: quantity}}}
	fulfillment := sale.Fulfillment{Updates: []sale.FulfillmentUpdate{update}}
	current.err = current.store.Fulfillment().Save(context.Background(), fulfillment, value.CommandID(row["LookupCommandID"]))
	current.observed = map[string]string{}

	return nil
}

func (current *world) alternativeSKU() string {
	if len(current.result.Alternatives) == 0 {
		return ""
	}

	return string(current.result.Alternatives[0].SKU)
}
