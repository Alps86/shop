package sale

import "github.com/Alps86/shop/internal/domain/value"

func NewInventory(sku value.SKU, price value.Money, stock int64) (Inventory, error) {
	item := Inventory{SKU: sku, Price: price, StockQuantity: stock, Revision: 1}
	if err := item.Validate(); err != nil {
		return Inventory{}, err
	}

	return item, nil
}

func (item Inventory) Validate() error {
	if err := item.Price.Validate(); err != nil {
		return err
	}

	if err := item.validateIdentity(); err != nil {
		return err
	}

	if item.StockQuantity < item.ReservedQuantity() {
		return value.NewError("BelowReserved", "StockQuantity")
	}

	return nil
}

func (item Inventory) validateIdentity() error {
	if value.IdentifierSyntax(string(item.SKU)).Validate() != nil {
		return value.NewError("InvalidInput", "SKU")
	}

	if item.StockQuantity < 0 {
		return value.NewError("InvalidInput", "StockQuantity")
	}

	if item.Revision == 0 {
		return value.NewError("InvalidInput", "Revision")
	}

	return nil
}

func (item Inventory) ReservedQuantity() int64 {
	var total int64
	for _, reservation := range item.Reservations {
		total += reservation.Quantity
	}

	return total
}

func (item Inventory) AvailableQuantity() int64 {
	return item.StockQuantity - item.ReservedQuantity()
}

func (item Inventory) checkRevision(revision uint64) error {
	if item.Revision != revision {
		return value.NewError("Conflict", "Revision")
	}

	return nil
}

func (item Inventory) Reserve(orderID value.OrderID, quantity int64, revision uint64) (Inventory, error) {
	if err := item.checkRevision(revision); err != nil {
		return item, err
	}

	if quantity <= 0 {
		return item, value.NewError("InvalidInput", "Quantity")
	}

	if quantity > item.AvailableQuantity() {
		return item, value.NewError("Unavailable", string(item.SKU))
	}

	item.Reservations = append(append([]Reservation{}, item.Reservations...), Reservation{ID: value.ReservationID(string(orderID) + "/" + string(item.SKU)), OrderID: orderID, Quantity: quantity})
	item.Revision++

	return item, nil
}

func (item Inventory) SetStock(stock int64, revision uint64) (Inventory, error) {
	if err := item.checkRevision(revision); err != nil {
		return item, err
	}

	if stock < 0 {
		return item, value.NewError("InvalidInput", "StockQuantity")
	}

	if stock < item.ReservedQuantity() {
		return item, value.NewError("BelowReserved", "StockQuantity")
	}

	item.StockQuantity = stock
	item.Revision++

	return item, nil
}

func (item Inventory) SetPrice(price value.Money, revision uint64) (Inventory, error) {
	if err := item.checkRevision(revision); err != nil {
		return item, err
	}

	if err := price.Validate(); err != nil {
		return item, err
	}

	if price.Currency != item.Price.Currency {
		return item, value.NewError("UnsupportedContext", "Currency")
	}

	item.Price = price
	item.Revision++

	return item, nil
}

func (item Inventory) Ship(orderID value.OrderID, quantity int64, revision uint64) (Inventory, error) {
	if err := item.checkRevision(revision); err != nil {
		return item, err
	}

	for index, reservation := range item.Reservations {
		if reservation.OrderID == orderID {
			return item.consume(index, quantity)
		}
	}

	return item, value.NewError("InvalidQuantity", "Reservation")
}

func (item Inventory) consume(index int, quantity int64) (Inventory, error) {
	if quantity <= 0 {
		return item, value.NewError("InvalidQuantity", "Quantity")
	}

	if quantity > item.Reservations[index].Quantity {
		return item, value.NewError("InvalidQuantity", "Quantity")
	}

	item.Reservations = append([]Reservation{}, item.Reservations...)
	item.Reservations[index].Quantity -= quantity
	item.StockQuantity -= quantity
	item.Revision++

	return item, nil
}
