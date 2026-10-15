package temporal_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"reflect"
	"strconv"
	"time"
)

func (current *world) registerDelivery(ctx *godog.ScenarioContext) {
	ctx.Step(`^Liefermeldungen durch den echten Orderworkflow laufen$`, current.deliveries)
	ctx.Step(`^der Orderworkflow terminal beendet ist$`, current.orderClosed)
	ctx.Step(`^der noch offene Orderworkflow administrativ beendet wird$`, current.stopOrder)
}

func (current *world) deliveries(table *godog.Table) error {
	for _, row := range current.rows(table) {
		if err := current.delivery(row); err != nil {
			return err
		}

		if err := current.waitDelivery(row); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) delivery(row map[string]string) error {
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	id := current.purchase.OrderID
	if row["OrderID"] != "" {
		id = value.OrderID(row["OrderID"])
	}

	update := sale.FulfillmentUpdate{OrderID: id, CommandID: value.CommandID(row["CommandID"]), ShipmentID: value.ShipmentID(row["ShipmentID"]), Status: sale.DeliveryStatus(row["Status"]), Lines: []sale.FulfillmentLine{{SKU: value.SKU(row["SKU"]), Quantity: quantity}}}
	err = current.adapter.ReportDelivery(context.Background(), application.DeliveryCommand{OrderID: id, CommandID: update.CommandID, Update: update})
	if value.Code(err) != row["Code"] {
		return fmt.Errorf("Liefercode %s statt %s: %w", value.Code(err), row["Code"], err)
	}

	return nil
}

func (current *world) waitDelivery(row map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		err := current.deliveryObserved(row)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
		}
	}
}

func (current *world) deliveryObserved(row map[string]string) error {
	order, err := current.service.GetOrder(context.Background(), current.guest.Session, current.purchase.OrderID)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(order.Purchase, current.initialPurchase) {
		return fmt.Errorf("Liefermeldung änderte Kaufsnapshot")
	}

	line := order.Fulfillment.Lines[0]
	actual := map[string]string{"State": order.Fulfillment.State, "Shipped": strconv.FormatInt(line.Shipped, 10), "Delivered": strconv.FormatInt(line.Delivered, 10)}
	for _, field := range []string{"State", "Shipped", "Delivered"} {
		if actual[field] != row[field] {
			return fmt.Errorf("%s: %s statt %s", field, actual[field], row[field])
		}
	}

	return nil
}

func (current *world) orderClosed() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return current.connection.GetWorkflow(ctx, "order/"+string(current.purchase.OrderID), "").Get(ctx, nil)
}

func (current *world) stopOrder() error {
	return current.connection.TerminateWorkflow(context.Background(), "order/"+string(current.purchase.OrderID), "", "Kontrollierter Prozessverlust im Integrationsvertrag")
}
