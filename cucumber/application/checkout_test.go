package application_test

import (
	"context"
	"reflect"
	"strconv"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerCheckout(ctx *godog.ScenarioContext) {
	current.registerPurchaseSteps(ctx)
	ctx.Step(`^der maßgebliche Variantenbestand lautet$`, current.variantInventory)
	ctx.Step(`^der Gast den Abschluss über den Prozesseingang anfordert$`, current.processPlace)
	ctx.Step(`^enthält der Prozessaufruf ausschließlich autorisierte Fachkennungen$`, current.processObserved)
	ctx.Step(`^die Bewertung ist ausdrücklich bestätigt$`, current.confirmed)
	ctx.Step(`^vor dem Abschluss eine Änderung stattfindet$`, current.changed)
	ctx.Step(`^der alte Abschlussversuch ausgeführt wird$`, current.oldPlace)
	ctx.Step(`^Checkoutdaten validiert werden$`, current.checkoutInvalid)
	ctx.Step(`^lautet das Validierungsergebnis$`, current.checkoutValidated)
	ctx.Step(`^ein vorbereiteter Cart ohne Bestätigung abgeschlossen wird$`, current.unconfirmed)
}

func (current *world) registerPurchaseSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein eigener Gastcart ist vollständig vorbereitet$`, current.prepared)
	ctx.Step(`^der Gast den bewerteten Gesamtpreis bestätigt und abschließt$`, current.confirmAndPlace)
	ctx.Step(`^lautet das Abschlussergebnis$`, current.purchased)
	ctx.Step(`^die bestätigte SKU wurde zwischenzeitlich unverfügbar$`, current.unavailableSKU)
	ctx.Step(`^der bestätigte Abschluss versucht wird$`, current.oldPlace)
	ctx.Step(`^wird ein unveränderter Cart mit Ersatzangebot geliefert$`, current.purchased)
	ctx.Step(`^der Gast die Ersatzvariante wählt und den neuen Preis bestätigt$`, current.replacement)
	ctx.Step(`^entsteht die Bestellung mit der aktiv gewählten SKU$`, current.purchased)
	ctx.Step(`^der maßgebliche Preis wurde nach alter Bestätigung geändert$`, current.repriced)
	ctx.Step(`^der Gast die neue Bewertung bestätigt und abschließt$`, current.newPricePlace)
}

func (current *world) prepared(table *godog.Table) error {
	row := read(table)[0]
	quantity, err := strconv.ParseInt(row["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	price, err := strconv.ParseInt(row["UnitPrice"], 10, 64)
	if err != nil {
		return err
	}

	return current.prepareStock(row, quantity, price)
}

func (current *world) prepareStock(row map[string]string, quantity, price int64) error {
	stock, err := strconv.ParseInt(row["StockQuantity"], 10, 64)
	if err != nil {
		return err
	}

	revision, err := strconv.ParseUint(row["CartRevision"], 10, 64)
	if err != nil {
		return err
	}

	return current.prepareCart(row, quantity, price, stock, revision)
}

func (current *world) prepareCart(row map[string]string, quantity, price, stock int64, revision uint64) error {
	id, user, sku := value.CartID(row["CartID"]), value.UserID(row["UserID"]), value.SKU(row["SKU"])
	current.sessions["a"] = application.SessionIdentity{UserID: user, IssuedAt: current.now}
	current.inventoryProducts[sku] = "p-phone"
	current.inventories[sku] = sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: stock, Revision: 1}
	checkout := sale.Checkout{CartID: id, ShippingAddress: sale.Address{Recipient: row["Recipient"], Street: row["Street"], PostalCode: row["PostalCode"], City: row["City"], Country: value.CountryCode(row["Country"])}, ShippingMethod: sale.Standard, PaymentMethod: sale.Invoice}
	current.carts[id] = sale.Cart{ID: id, UserID: user, Items: sale.CartItems{{SKU: sku, Quantity: quantity}}, Revision: revision, State: sale.CheckingOut, LastActivityAt: current.now, DefinitionVersion: "cart-v1", RuleSetVersion: "demo-v1", Checkout: value.Optional[sale.Checkout]{Value: checkout, Present: true}}
	current.placeCommand = application.PlaceCommand{UserID: "u-a", CartID: id, CommandID: "cmd-a", EvaluationID: value.EvaluationID(row["EvaluationID"])}

	return current.evaluateAndConfirm()
}

func (current *world) evaluateAndConfirm() error {
	cart := current.carts[current.placeCommand.CartID]
	evaluation, err := (evaluator{current}).Evaluate(context.Background(), cart)
	if err != nil {
		return err
	}

	evaluation.ID = current.placeCommand.EvaluationID
	cart, err = cart.SetEvaluation(evaluation)
	if err != nil {
		return err
	}

	cart, err = cart.Confirm(evaluation.ID, cart.Revision, current.now)
	current.carts[cart.ID] = cart

	return err
}

func (current *world) confirmAndPlace(table *godog.Table) error {
	row := read(table)[0]
	cart := current.carts["c-a"]
	cart.Checkout.Value.ShippingMethod = sale.ShippingMethod(row["ShippingMethod"])
	cart.Checkout.Value.PaymentMethod = sale.PaymentMethod(row["PaymentMethod"])
	current.carts["c-a"] = cart
	current.authorization = row["SimulatedAuthorization"]
	current.placeCommand.CommandID = value.CommandID(row["CommandID"])
	current.placeCommand.EvaluationID = value.EvaluationID(row["EvaluationID"])
	if err := current.evaluateAndConfirm(); err != nil {
		return err
	}

	current.place()

	return nil
}

func (current *world) place() {
	current.purchaseResult, current.err = current.service.ApplyPlace(context.Background(), current.placeCommand)
}

func (current *world) oldPlace(table *godog.Table) {
	row := read(table)[0]
	current.placeCommand.CommandID = value.CommandID(row["CommandID"])
	current.place()
}

func (current *world) confirmed(table *godog.Table) error {
	cart := current.carts["c-a"]
	evaluation := cart.Checkout.Value.ConfirmedEvaluation.Value

	return read(table).matches(map[string]string{"EvaluationID": string(evaluation.ID), "CartRevision": strconv.FormatUint(cart.Revision, 10), "Total": strconv.FormatInt(evaluation.Total.MinorUnits, 10)})
}

func (current *world) changed(table *godog.Table) error {
	row := read(table)[0]
	if row["Change"] == "inventoryPrice" {
		price, err := strconv.ParseInt(row["NewValue"], 10, 64)
		inventory := current.inventories["PHONE"]
		inventory.Price.MinorUnits = price
		inventory.Revision++
		current.inventories["PHONE"] = inventory

		return err
	}

	return current.changeCart(row)
}

func (current *world) changeCart(row map[string]string) error {
	cart := current.carts["c-a"]
	command := application.ChangeCommand{UserID: "u-a", CartID: cart.ID, ExpectedRevision: cart.Revision, CommandID: "cmd-change", Operation: row["Change"]}
	command.ShippingMethod = sale.ShippingMethod(row["NewValue"])
	command.PaymentMethod = sale.PaymentMethod(row["NewValue"])
	command.Address = cart.Checkout.Value.ShippingAddress
	command.Address.Street = row["NewValue"]
	if row["Change"] == "quantity" {
		command.Operation = "setQuantity"
		command.SKU = "PHONE"
		command.Quantity = 2
	}

	_, err := current.service.ApplyChange(context.Background(), command)

	return err
}

func (current *world) purchased(table *godog.Table) error {
	cart := current.carts["c-a"]
	total := "absent"
	sku := ""
	if order, exists := current.orders[current.purchaseResult.OrderID]; exists {
		total = strconv.FormatInt(order.Total.MinorUnits, 10)
		sku = string(order.Items[0].SKU)
	}

	reserved := current.reservedQuantity()

	evaluation := current.purchaseResult.Cart.Value.Cart.Evaluation.Value

	actual := map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.Itoa(len(current.orders)), "Reserved": strconv.FormatInt(reserved, 10), "Total": total, "AuthorizationCalls": strconv.Itoa(current.authorizations), "RealPaymentCalls": "0", "NewTotal": strconv.FormatInt(evaluation.Total.MinorUnits, 10), "ConfirmationPresent": strconv.FormatBool(cart.Checkout.Value.ConfirmedEvaluation.Present), "CartSKU": string(cart.Items[0].SKU), "ProposedSKU": current.proposedSKU(), "SKU": sku, "CartPresent": strconv.FormatBool(current.purchaseResult.Cart.Present), "ReturnedConfirmationPresent": strconv.FormatBool(current.purchaseResult.Cart.Value.Cart.Checkout.Value.ConfirmedEvaluation.Present), "ReturnedEvaluationPresent": strconv.FormatBool(current.purchaseResult.Cart.Value.Cart.Evaluation.Present), "UnavailableSKU": string(current.purchaseResult.UnavailableSKU.Value), "AlternativeCount": strconv.Itoa(len(current.purchaseResult.Alternatives))}

	return read(table).matches(actual)
}

func (current *world) checkoutInvalid(table *godog.Table) {
	row := read(table)[0]
	checkout := current.carts["c-a"].Checkout.Value
	field := reflect.ValueOf(&checkout.ShippingAddress).Elem().FieldByName(row["Field"])
	if !field.IsValid() {
		field = reflect.ValueOf(&checkout).Elem().FieldByName(row["Field"])
	}

	content := row["Value"]
	if content == "empty" {
		content = ""
	}

	field.SetString(content)
	current.err = checkout.Validate()
}

func (current *world) checkoutValidated(table *godog.Table) error {
	return read(table).matches(map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.Itoa(len(current.orders))})
}

func (current *world) unconfirmed(_ *godog.Table) {
	cart := current.carts["c-a"]
	cart.Checkout.Value.ConfirmedEvaluation = value.Optional[sale.CartEvaluation]{}
	current.carts["c-a"] = cart
	current.place()
}

func (current *world) unavailableSKU(table *godog.Table) error {
	row := read(table)[0]
	price, err := strconv.ParseInt(row["AlternativePrice"], 10, 64)
	if err != nil {
		return err
	}

	inventory := current.inventories[value.SKU(row["SKU"])]
	inventory.StockQuantity = 0
	current.inventories[inventory.SKU] = inventory
	sku := value.SKU(row["SameProductAlternative"])
	current.inventoryProducts[sku] = "p-phone"
	current.inventories[sku] = sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: 2, Revision: 1}

	return nil
}

func (current *world) replacement(table *godog.Table) error {
	row := read(table)[0]
	cart := current.carts["c-a"]
	cart.Items = sale.CartItems{{SKU: value.SKU(row["SKU"]), Quantity: 1}}
	cart.Revision++
	current.carts[cart.ID] = cart
	current.placeCommand.CommandID = value.CommandID(row["CommandID"])
	current.placeCommand.EvaluationID = "e-b"
	if err := current.evaluateAndConfirm(); err != nil {
		return err
	}

	current.place()

	return nil
}

func (current *world) repriced(table *godog.Table) error {
	row := read(table)[0]
	total, err := strconv.ParseInt(row["NewTotal"], 10, 64)
	if err != nil {
		return err
	}

	inventory := current.inventories["PHONE"]
	inventory.Price.MinorUnits = total - 490
	current.inventories["PHONE"] = inventory
	cart := current.carts["c-a"]
	cart.Revision = 5
	current.carts["c-a"] = cart
	current.placeCommand.EvaluationID = value.EvaluationID(row["NewEvaluationID"])

	return current.evaluateAndConfirm()
}

func (current *world) newPricePlace(table *godog.Table) {
	row := read(table)[0]
	current.placeCommand.EvaluationID = value.EvaluationID(row["EvaluationID"])
	current.placeCommand.CommandID = value.CommandID(row["CommandID"])
	current.place()
}

func (current *world) proposedSKU() string {
	if len(current.purchaseResult.Alternatives) == 0 {
		return "absent"
	}

	return string(current.purchaseResult.Alternatives[0].SKU)
}

func (current *world) variantInventory(table *godog.Table) error {
	for _, row := range read(table) {
		price, err := strconv.ParseInt(row["UnitPrice"], 10, 64)
		if err != nil {
			return err
		}

		quantity, err := strconv.ParseInt(row["StockQuantity"], 10, 64)
		if err != nil {
			return err
		}

		sku := value.SKU(row["SKU"])
		current.inventories[sku] = sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: quantity, Revision: 1}
		current.inventoryProducts[sku] = value.ProductID(row["ProductID"])
	}

	return nil
}
