package sale_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strings"
)

func (current *scenario) bindFulfillment(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Bestellung wird für die Lieferdemo verwendet$`, current.fulfillmentGiven)
	ctx.Step(`^fachliche Liefermeldungen nacheinander angewendet werden$`, current.fulfillmentSequence)
	ctx.Step(`^lautet der fachliche Lieferzustand$`, current.fulfillmentResult)
}

func (current *scenario) fulfillmentGiven(table *godog.Table) {
	row := rows(table)[0]
	current.order = sale.Order{ID: value.OrderID(row["OrderID"]), Items: []sale.OrderItem{{SKU: value.SKU(row["SKU"]), Quantity: number(row["Purchased"])}}}
	current.fulfillment = sale.NewFulfillment(current.order)
}

func (current *scenario) fulfillmentSequence(table *godog.Table) error {
	for _, row := range rows(table) {
		update := sale.FulfillmentUpdate{OrderID: value.OrderID(row["OrderID"]), ShipmentID: value.ShipmentID(row["ShipmentID"]), CommandID: value.CommandID(row["CommandID"]), Status: sale.DeliveryStatus(row["Status"])}
		if row["SKU"] != "none" {
			update.Lines = deliveryLines(row)
		}

		current.fulfillment, current.err = current.fulfillment.Apply(update)
		if value.Code(current.err) != row["Code"] {
			return fmt.Errorf("%s: erwartet %s, erhalten %w", row["CommandID"], row["Code"], current.err)
		}
	}

	return nil
}

func (current *scenario) fulfillmentResult(table *godog.Table) error {
	return check(table, map[string]string{"PackedQuantity": text(current.packedQuantity()), "State": current.fulfillment.State(), "Shipped": text(current.fulfillment.Shipped("PHONE")), "Delivered": text(current.fulfillment.Delivered("PHONE"))})
}

func deliveryLines(row map[string]string) []sale.FulfillmentLine {
	result := []sale.FulfillmentLine{}
	for _, sku := range strings.Split(row["SKU"], ",") {
		result = append(result, sale.FulfillmentLine{SKU: value.SKU(sku), Quantity: number(row["Quantity"])})
	}

	return result
}

func (current *scenario) packedQuantity() int64 {
	var quantity int64
	for _, update := range current.fulfillment.Updates {
		if update.Status == sale.Packed {
			quantity += update.Lines[0].Quantity
		}
	}

	return quantity
}
