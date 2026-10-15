package application_test

import (
	"context"
	messages "github.com/cucumber/messages/go/v21"
	"reflect"
	"strconv"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerDelivery(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Bestellung enthält unterschiedliche Positionen$`, current.deliveryPositions)
	ctx.Step(`^der Eigentümer die Bestellansicht liest$`, current.readDelivery)
	ctx.Step(`^enthält die separate Lieferansicht geordnete Positionen$`, current.deliveredPositions)
	ctx.Step(`^eine angenommene Bestellung wurde noch nicht versandt$`, current.deliveryOrder)
	ctx.Step(`^eine lokale simulierte Liefermeldung verarbeitet wird$`, current.reportDelivery)
	ctx.Step(`^lautet die getrennte Lieferansicht$`, current.delivered)
	ctx.Step(`^eine Teilmenge wurde bereits versandt$`, current.partiallyShipped)
	ctx.Step(`^eine weitere Liefermeldung verarbeitet wird$`, current.reportDelivery)
	ctx.Step(`^lautet die Lieferbeobachtung$`, current.delivered)
	ctx.Step(`^eine Bestellung mit reservierter Kaufmenge ist angenommen$`, current.deliveryOrder)
	ctx.Step(`^die lokale Liefersimulation nacheinander Meldungen verarbeitet$`, current.reportDeliveries)
	ctx.Step(`^lautet die endgültige Lieferansicht$`, current.delivered)
}

func (current *world) deliveryOrder(table *godog.Table) error {
	row := read(table)[0]
	quantity, err := strconv.ParseInt(row["Purchased"], 10, 64)
	if err != nil {
		return err
	}

	current.orders[value.OrderID(row["OrderID"])] = sale.Order{ID: value.OrderID(row["OrderID"]), Items: []sale.OrderItem{{SKU: value.SKU(row["SKU"]), Quantity: quantity}}}
	current.initialOrder = current.orders[value.OrderID(row["OrderID"])].Snapshot()
	current.fulfillment = sale.NewFulfillment(current.orders[value.OrderID(row["OrderID"])])

	return nil
}

func (current *world) reportDelivery(table *godog.Table) error {
	row := read(table)[0]
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	sku := value.SKU(row["SKU"])
	if sku == "" {
		sku = "PHONE"
	}

	command := application.DeliveryCommand{OrderID: "o-a", CommandID: value.CommandID(row["CommandID"]), Update: sale.FulfillmentUpdate{OrderID: "o-a", ShipmentID: value.ShipmentID(row["ShipmentID"]), CommandID: value.CommandID(row["CommandID"]), Status: sale.DeliveryStatus(row["Status"]), Lines: []sale.FulfillmentLine{{SKU: sku, Quantity: quantity}}}}
	current.fulfillmentView, current.err = (application.FulfillmentService{Store: deliveryStore{current}}).Report(context.Background(), command)

	return nil
}

func (current *world) partiallyShipped(table *godog.Table) error {
	if err := current.deliveryOrder(table); err != nil {
		return err
	}

	row := read(table)[0]
	update := sale.FulfillmentUpdate{OrderID: value.OrderID(row["OrderID"]), ShipmentID: value.ShipmentID(row["ShipmentID"]), CommandID: value.CommandID(row["CommandID"]), Status: sale.Shipped, Lines: []sale.FulfillmentLine{{SKU: value.SKU(row["SKU"]), Quantity: 1}}}
	fulfillment, err := current.fulfillment.Apply(update)
	current.fulfillment = fulfillment

	return err
}

func (current *world) reportDeliveries(table *godog.Table) error {
	for _, row := range table.Rows[1:] {
		next := &godog.Table{Rows: []*messages.PickleTableRow{table.Rows[0], row}}
		if err := current.reportDelivery(next); err != nil {
			return err
		}

		if current.err != nil {
			return current.err
		}
	}

	return nil
}

func (current *world) delivered(table *godog.Table) error {
	quantity := current.fulfillment.Purchased["PHONE"]
	shipped := current.fulfillment.Shipped("PHONE")
	delivered := current.fulfillment.Delivered("PHONE")
	actual := map[string]string{"Code": value.Code(current.err), "State": current.fulfillment.State(), "Shipped": strconv.FormatInt(shipped, 10), "Delivered": strconv.FormatInt(delivered, 10), "Reserved": strconv.FormatInt(quantity-shipped, 10), "PurchaseSnapshotChanged": strconv.FormatBool(!reflect.DeepEqual(current.initialOrder, current.orders["o-a"].Snapshot()))}

	return read(table).matches(actual)
}

func (port deliveryStore) Get(_ context.Context, _ value.OrderID) (sale.Fulfillment, error) {
	if port.current.fault == "fulfillmentRead" {
		return sale.Fulfillment{}, value.NewError("DependencyUnavailable", "Fulfillment")
	}

	return port.current.fulfillment, nil
}

func (port deliveryStore) Save(_ context.Context, fulfillment sale.Fulfillment, _ value.CommandID) error {
	if port.current.fault == "fulfillmentWrite" {
		return value.NewError("DependencyUnavailable", "Fulfillment")
	}

	port.current.fulfillment = fulfillment

	return nil
}

func (port cartStarter) Start(_ context.Context, user value.UserID, _ value.CommandID, now time.Time) (application.CartView, error) {
	if port.current.fault == "cartStart" {
		return application.CartView{}, value.NewError("DependencyUnavailable", "Cart")
	}

	cart, err := sale.NewCart("c-new", user, now)
	port.current.carts[cart.ID] = cart

	return application.CartView{Cart: cart}, err
}

func (current *world) deliveryPositions(table *godog.Table) error {
	order := sale.Order{ID: "o-a", UserID: "u-a"}
	for _, row := range read(table) {
		quantity, err := strconv.ParseInt(row["Purchased"], 10, 64)
		if err != nil {
			return err
		}

		order.Items = append(order.Items, sale.OrderItem{SKU: value.SKU(row["SKU"]), Quantity: quantity})
	}

	current.orders[order.ID] = order
	current.initialOrder = order.Snapshot()
	current.fulfillment = sale.NewFulfillment(order)
	current.sessions["a"] = application.SessionIdentity{UserID: order.UserID, IssuedAt: current.now}

	return nil
}

func (current *world) readDelivery(table *godog.Table) {
	row := read(table)[0]
	current.orderView, current.err = current.service.GetOrder(context.Background(), application.Session(row["Session"]), value.OrderID(row["OrderID"]))
}

func (current *world) deliveredPositions(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	for index, row := range read(table) {
		line := current.orderView.Fulfillment.Lines[index]
		if err := (rows{row}).matches(map[string]string{"SKU": string(line.SKU), "Purchased": strconv.FormatInt(line.Purchased, 10), "Shipped": strconv.FormatInt(line.Shipped, 10), "Delivered": strconv.FormatInt(line.Delivered, 10)}); err != nil {
			return err
		}
	}

	return nil
}
