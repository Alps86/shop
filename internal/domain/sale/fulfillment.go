package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"reflect"
)

const (
	Packed    DeliveryStatus = "packed"
	Shipped   DeliveryStatus = "shipped"
	Delivered DeliveryStatus = "delivered"
)

func NewFulfillment(order Order) Fulfillment {
	quantities := map[value.SKU]int64{}
	for _, item := range order.Items {
		quantities[item.SKU] = item.Quantity
	}

	return Fulfillment{OrderID: order.ID, Purchased: quantities}
}

func (fulfillment Fulfillment) Apply(update FulfillmentUpdate) (Fulfillment, error) {
	for _, previous := range fulfillment.Updates {
		if previous.matchesIdentity(update) {
			return fulfillment.repeat(previous, update)
		}
	}

	if err := fulfillment.validate(update); err != nil {
		return fulfillment, err
	}

	update.Lines = append([]FulfillmentLine{}, update.Lines...)
	fulfillment.Updates = append(append([]FulfillmentUpdate{}, fulfillment.Updates...), update)

	return fulfillment, nil
}

func (update FulfillmentUpdate) matchesIdentity(other FulfillmentUpdate) bool {
	return update.CommandID == other.CommandID || update.ShipmentID == other.ShipmentID && update.Status == other.Status
}

func (fulfillment Fulfillment) repeat(previous, update FulfillmentUpdate) (Fulfillment, error) {
	if !reflect.DeepEqual(previous, update) {
		return fulfillment, value.NewError("Conflict", "Shipment")
	}

	return fulfillment, nil
}

func (fulfillment Fulfillment) validate(update FulfillmentUpdate) error {
	if err := fulfillment.validateProgressIdentity(update); err != nil {
		return err
	}

	if err := update.validateLines(); err != nil {
		return err
	}

	for _, line := range update.Lines {
		if err := fulfillment.validateLine(update, line); err != nil {
			return err
		}
	}

	return nil
}

func (update FulfillmentUpdate) validateLines() error {
	if len(update.Lines) == 0 {
		return value.NewError("InvalidQuantity", "Lines")
	}

	seen := map[value.SKU]bool{}
	for _, line := range update.Lines {
		if seen[line.SKU] {
			return value.NewError("InvalidQuantity", "SKU")
		}

		seen[line.SKU] = true
	}

	return nil
}

func (fulfillment Fulfillment) validateLine(update FulfillmentUpdate, line FulfillmentLine) error {
	if line.Quantity <= 0 {
		return value.NewError("InvalidQuantity", "Quantity")
	}

	if fulfillment.Purchased[line.SKU] == 0 {
		return value.NewError("InvalidQuantity", "SKU")
	}

	if update.stockProgress() {
		return fulfillment.validateStockProgress(update, line)
	}

	if update.Status == Delivered {
		return fulfillment.validateDelivery(update.ShipmentID, line)
	}

	return value.NewError("InvalidTransition", "Status")
}

func (update FulfillmentUpdate) stockProgress() bool {
	return update.Status == Packed || update.Status == Shipped
}

func (fulfillment Fulfillment) validateShipment(line FulfillmentLine) error {
	if fulfillment.Shipped(line.SKU)+line.Quantity > fulfillment.Purchased[line.SKU] {
		return value.NewError("InvalidQuantity", "Quantity")
	}

	return nil
}

func (fulfillment Fulfillment) validateDelivery(shipment value.ShipmentID, line FulfillmentLine) error {
	if line.Quantity > fulfillment.shipmentQuantity(shipment, line.SKU, Shipped) {
		return value.NewError("InvalidTransition", "Shipment")
	}

	return nil
}

func (fulfillment Fulfillment) shipmentQuantity(shipment value.ShipmentID, sku value.SKU, status DeliveryStatus) int64 {
	var quantity int64
	for _, update := range fulfillment.Updates {
		if update.matches(shipment, status) {
			quantity += update.quantity(sku)
		}
	}

	return quantity
}

func (update FulfillmentUpdate) matches(shipment value.ShipmentID, status DeliveryStatus) bool {
	return update.ShipmentID == shipment && update.Status == status
}

func (update FulfillmentUpdate) quantity(sku value.SKU) int64 {
	var quantity int64
	for _, line := range update.Lines {
		if line.SKU == sku {
			quantity += line.Quantity
		}
	}

	return quantity
}

func (fulfillment Fulfillment) quantity(sku value.SKU, status DeliveryStatus) int64 {
	var quantity int64
	for _, update := range fulfillment.Updates {
		if update.Status == status {
			quantity += update.quantity(sku)
		}
	}

	return quantity
}

func (fulfillment Fulfillment) Shipped(sku value.SKU) int64 {
	return fulfillment.quantity(sku, Shipped)
}

func (fulfillment Fulfillment) Delivered(sku value.SKU) int64 {
	return fulfillment.quantity(sku, Delivered)
}

func (fulfillment Fulfillment) State() string {
	var purchased, shipped, delivered int64
	for sku, quantity := range fulfillment.Purchased {
		purchased += quantity
		shipped += fulfillment.Shipped(sku)
		delivered += fulfillment.Delivered(sku)
	}

	if delivered == purchased {
		return "delivered"
	}

	if shipped == purchased {
		return "shipped"
	}

	if shipped > 0 {
		return "partiallyShipped"
	}

	return "accepted"
}

func (fulfillment Fulfillment) validateProgressIdentity(update FulfillmentUpdate) error {
	if update.OrderID != fulfillment.OrderID {
		return value.NewError("InvalidReference", "OrderID")
	}

	if fulfillment.State() == "delivered" {
		return value.NewError("InvalidTransition", "Status")
	}

	return nil
}

func (fulfillment Fulfillment) validateStockProgress(update FulfillmentUpdate, line FulfillmentLine) error {
	if update.Status == Packed {
		return fulfillment.validatePacking(line)
	}

	return fulfillment.validateShipment(line)
}

func (fulfillment Fulfillment) validatePacking(line FulfillmentLine) error {
	if fulfillment.quantity(line.SKU, Packed)+line.Quantity > fulfillment.Purchased[line.SKU] {
		return value.NewError("InvalidQuantity", "Quantity")
	}

	return nil
}
