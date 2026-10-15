package application_test

import (
	"context"
	"strconv"
	"sync"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerAtomicity(ctx *godog.ScenarioContext) {
	ctx.Step(`^die synthetischen Kaufvertragsgrunddaten gelten$`, current.atomicDefaultsGiven)
	ctx.Step(`^sich die Positionspreise vor dem Abschluss ändern$`, current.atomicPriceShift)
	ctx.Step(`^ein gültiger Abschlussversuch ist vorbereitet$`, current.atomicPrepared)
	ctx.Step(`^der Abschluss mit einer Störung und Wiederholung ausgeführt wird$`, current.atomicRetry)
	ctx.Step(`^lautet das dauerhafte Kaufergebnis$`, current.atomicResult)
	ctx.Step(`^zwei bestätigte Carts benötigen dieselbe letzte Einheit$`, current.atomicCompetitors)
	ctx.Step(`^beide Kauftransaktionen gleichzeitig freigegeben werden$`, current.atomicConcurrent)
	ctx.Step(`^lautet das gemeinsame Ergebnis unabhängig vom Gewinner$`, current.atomicConcurrentResult)
	ctx.Step(`^ein bestätigter Cart benötigt mehrere SKUs$`, current.atomicMultiple)
	ctx.Step(`^der Commit nach dem ersten Reservierungsschritt fehlschlägt$`, current.atomicRollback)
	ctx.Step(`^lautet der persistierte Zustand$`, current.atomicRolledBack)
	ctx.Step(`^eine bestätigte Bewertung ist weiterhin verfügbar$`, current.atomicStockRevision)
	ctx.Step(`^derselbe bestätigte Kauf abgeschlossen wird$`, current.place)
}

func (current *world) atomicDefaultsGiven(table *godog.Table) {
	current.atomicDefaults = read(table)[0]
}

func (current *world) atomicPrepared(table *godog.Table) error {
	return current.prepareAtomic(read(table)[0])
}

func (current *world) prepareAtomic(row map[string]string) error {
	data := map[string]string{}
	for field, content := range current.atomicDefaults {
		data[field] = content
	}

	for field, content := range row {
		data[field] = content
	}

	if err := current.prepareCart(data, atomicInteger(data["Quantity"]), atomicInteger(data["UnitPrice"]), atomicInteger(data["StockQuantity"]), atomicUnsigned(data["CartRevision"])); err != nil {
		return err
	}

	current.placeCommand.UserID = value.UserID(data["UserID"])
	current.placeCommand.CommandID = value.CommandID(data["CommandID"])

	return current.prepareAtomicCheckout(data)
}

func atomicInteger(raw string) int64 {
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return parsed
}

func atomicUnsigned(raw string) uint64 {
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		panic(err)
	}

	return parsed
}

func (current *world) atomicRetry(table *godog.Table) error {
	row := read(table)[0]
	current.fault = row["Fault"]
	current.place()
	if err := current.firstAttempt(row); err != nil {
		return err
	}

	current.fault = ""
	current.placeCommand.CommandID = value.CommandID(row["RetryCommandID"])
	if row["RetryPayload"] == "different" {
		current.placeCommand.EvaluationID = "e-different"
	}

	current.place()

	return nil
}

func (current *world) atomicResult(table *godog.Table) error {
	reserved := current.reservedQuantity()

	return read(table).matches(map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.Itoa(len(current.orders)), "Reservations": strconv.FormatInt(reserved, 10), "CartState": string(current.carts[current.placeCommand.CartID].State), "OrderID": current.durableOrderID(), "OrderOutboxCount": strconv.Itoa(len(current.orderOutbox)), "AuthorizationCalls": strconv.Itoa(current.authorizations)})
}

func (current *world) reservedQuantity() int64 {
	var reserved int64
	for _, inventory := range current.inventories {
		reserved += inventory.ReservedQuantity()
	}

	return reserved
}

func (current *world) atomicCompetitors(table *godog.Table) error {
	for _, row := range read(table) {
		data := map[string]string{}
		for field, content := range row {
			data[field] = content
		}

		data["StockQuantity"] = row["AvailableQuantity"]
		if err := current.prepareAtomic(data); err != nil {
			return err
		}

		current.placeCommand.CommandID = value.CommandID("cmd-" + row["CartID"])
		current.atomicCommands = append(current.atomicCommands, current.placeCommand)
	}

	return nil
}

func (current *world) atomicConcurrent() {
	var pending sync.WaitGroup
	results := make(chan application.PurchaseResult, len(current.atomicCommands))
	errors := make(chan error, len(current.atomicCommands))
	start := make(chan struct{})
	for _, command := range current.atomicCommands {
		pending.Add(1)
		go func(command application.PlaceCommand) {
			defer pending.Done()
			<-start
			result, err := current.service.Purchase.Commit(context.Background(), application.PurchaseCommand{UserID: command.UserID, CartID: command.CartID, CommandID: command.CommandID, EvaluationID: command.EvaluationID, Now: current.now})
			results <- result
			errors <- err
		}(command)
	}

	close(start)
	pending.Wait()
	current.collectAtomic(results, errors)
}

func (current *world) collectAtomic(results chan application.PurchaseResult, errors chan error) {
	close(results)
	close(errors)
	for result := range results {
		current.atomicResults = append(current.atomicResults, result)
	}

	for err := range errors {
		current.atomicErrors = append(current.atomicErrors, err)
	}
}

func (current *world) atomicConcurrentResult(table *godog.Table) error {
	var unavailable, successful int
	for _, err := range current.atomicErrors {
		if value.Code(err) == "Unavailable" {
			unavailable++
		}

		if err == nil {
			successful++
		}
	}

	return read(table).matches(map[string]string{"SuccessfulOrders": strconv.Itoa(successful), "OrderCount": strconv.Itoa(len(current.orders)), "UnavailableResults": strconv.Itoa(unavailable), "Reserved": strconv.FormatInt(current.reservedQuantity(), 10), "AvailableQuantity": strconv.FormatInt(current.inventories["LAST"].AvailableQuantity(), 10), "PartialOrders": strconv.Itoa(current.partialOrders())})
}

func (current *world) partialOrders() int {
	var partial int
	for _, order := range current.orders {
		if current.isPartial(order) {
			partial++
		}
	}

	return partial
}

func (current *world) isPartial(order sale.Order) bool {
	for _, item := range order.Items {
		if current.inventories[item.SKU].ReservedQuantity() != item.Quantity {
			return true
		}
	}

	return false
}

func (current *world) atomicMultiple(table *godog.Table) error {
	first := read(table)[0]
	if err := current.prepareAtomic(first); err != nil {
		return err
	}

	cart := current.carts[current.placeCommand.CartID]
	cart.Items = nil
	for _, row := range read(table) {
		sku := value.SKU(row["SKU"])
		cart.Items = append(cart.Items, sale.CartItem{SKU: sku, Quantity: atomicInteger(row["Quantity"])})
		current.inventories[sku] = sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: atomicInteger(current.atomicDefaults["UnitPrice"]), Currency: "EUR"}, StockQuantity: atomicInteger(current.atomicDefaults["StockQuantity"]), Revision: 1}
	}

	current.carts[cart.ID] = cart

	return current.evaluateAndConfirm()
}

func (current *world) atomicRollback(table *godog.Table) {
	current.fault = read(table)[0]["Fault"]
	current.place()
}

func (current *world) atomicRolledBack(table *godog.Table) error {
	return read(table).matches(map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.Itoa(len(current.orders)), "ReservationCount": strconv.FormatInt(current.reservedQuantity(), 10), "OrderOutboxCount": strconv.Itoa(len(current.orderOutbox)), "CartState": string(current.carts[current.placeCommand.CartID].State)})
}

func (current *world) atomicStockRevision(table *godog.Table) error {
	row := read(table)[0]
	if err := current.prepareAtomic(map[string]string{"UnitPrice": row["UnitPrice"], "StockQuantity": row["AvailableBefore"]}); err != nil {
		return err
	}

	inventory := current.inventories[current.carts[current.placeCommand.CartID].Items[0].SKU]
	inventory.Revision = atomicUnsigned(row["InventoryRevisionBefore"])
	current.inventories[inventory.SKU] = inventory
	if err := current.evaluateAndConfirm(); err != nil {
		return err
	}

	inventory.StockQuantity = atomicInteger(row["AvailableNow"])
	inventory.Revision = atomicUnsigned(row["InventoryRevisionNow"])
	current.inventories[inventory.SKU] = inventory

	return nil
}

func (current *world) atomicPriceShift(table *godog.Table) error {
	for _, row := range read(table) {
		sku := value.SKU(row["SKU"])
		item := current.inventories[sku]
		changed, err := item.SetPrice(value.Money{MinorUnits: atomicInteger(row["UnitPrice"]), Currency: item.Price.Currency}, item.Revision)
		if err != nil {
			return err
		}

		current.inventories[sku] = changed
	}

	return nil
}

func (current *world) durableOrderID() string {
	for _, order := range current.orders {
		if order.CartID == current.placeCommand.CartID {
			return string(order.ID)
		}
	}

	return "absent"
}

func (current *world) firstAttempt(row map[string]string) error {
	expected := rows{{"Code": row["FirstCode"], "Orders": row["FirstOrders"]}}

	return expected.matches(map[string]string{"Code": value.Code(current.err), "Orders": strconv.Itoa(len(current.orders))})
}

func (current *world) prepareAtomicCheckout(row map[string]string) error {
	cart := current.carts[current.placeCommand.CartID]
	cart.Checkout.Value.ShippingMethod = sale.ShippingMethod(row["ShippingMethod"])
	cart.Checkout.Value.PaymentMethod = sale.PaymentMethod(row["PaymentMethod"])
	current.carts[cart.ID] = cart
	current.authorization = row["SimulatedAuthorization"]

	return current.evaluateAndConfirm()
}
