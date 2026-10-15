package sale_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *scenario) bindCheckout(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Gastwarenkorb angelegt wird$`, current.cartCreate)
	ctx.Step(`^die Bewertung als Bestellung bestätigt wird$`, current.orderFromEvaluation)
	ctx.Step(`^die ursprünglichen Rabattdaten geändert werden$`, current.mutateDiscount)
	ctx.Step(`^besitzt die Kaufkopie unveränderte Rabattdaten$`, current.snapshotDiscount)
	ctx.Step(`^eine Bestellung mit expliziten Kaufkopien angefordert wird$`, current.orderExplicit)
	ctx.Step(`^eine zeitbezogene Cartoperation ausgeführt wird$`, current.cartTimed)
	ctx.Step(`^lautet der zeitbezogene Cartzustand$`, current.cartTimedResult)
	ctx.Step(`^Checkoutdaten validiert werden$`, current.checkoutValidate)
	ctx.Step(`^der fachliche Kaufablauf schrittweise ausgeführt wird$`, current.purchaseSequence)
	ctx.Step(`^enthält der erzeugte Kauf seine eigenen Preiswerte$`, current.purchaseResult)
}

func (current *scenario) cartCreate(table *godog.Table) {
	row := rows(table)[0]
	current.cart, current.err = sale.NewCart(value.CartID(input(row["CartID"])), value.UserID(input(row["UserID"])), instant())
}

func (current *scenario) cartTimed(table *godog.Table) {
	row := rows(table)[0]
	now := instant().Add(time.Duration(number(row["HoursSinceActivity"])) * time.Hour)
	sku := value.SKU(input(row["SKU"]))
	revision := unsigned(row["ExpectedRevision"])
	operations := map[string]func(){
		"add": func() {
			current.cart, current.err = current.cart.Add(sku, number(row["Quantity"]), revision, now)
		},
		"remove": func() {
			current.cart, current.err = current.cart.Remove(sku, revision, now)
		},
		"expire": func() {
			current.cart, current.err = current.cart.Expire(now)
		},
	}
	operations[row["Operation"]]()
}

func (current *scenario) cartTimedResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err), "State": string(current.cart.State), "ItemCount": strconv.Itoa(len(current.cart.Items)), "QuantityPHONE": text(current.cart.Items.Quantity("PHONE"))})
}

func (current *scenario) checkoutValidate(table *godog.Table) {
	row := rows(table)[0]
	checkout := checkoutInput(row)
	current.err = checkout.Validate()
}

func (current *scenario) purchaseSequence(table *godog.Table) error {
	for _, row := range rows(table) {
		current.purchaseAction(row)
		if value.Code(current.err) != row["ExpectedCode"] {
			return fmt.Errorf("%s: erwartet %s, erhalten %w", row["Operation"], row["ExpectedCode"], current.err)
		}

		if string(current.cart.State) != row["ExpectedState"] {
			return fmt.Errorf("%s: Zustand %s", row["Operation"], current.cart.State)
		}
	}

	return nil
}

func (current *scenario) purchaseAction(row map[string]string) {
	actions := map[string]func(map[string]string){
		"configure":        current.configureAction,
		"configureInvalid": current.configureAction,
		"evaluate":         current.setEvaluation,
		"confirm":          current.confirmAction,
		"markOrdered":      current.markOrderedAction,
		"createOrder":      current.createOrderAction,
		"createEmptyOrder": current.createOrderAction,
		"expire":           current.expireAction,
	}

	actions[row["Operation"]](row)
}

func (current *scenario) configureAction(row map[string]string) {
	current.configure(unsigned(row["Revision"]), row["Operation"] == "configureInvalid")
}

func (current *scenario) confirmAction(row map[string]string) {
	current.cart, current.err = current.cart.Confirm(value.EvaluationID(row["EvaluationID"]), unsigned(row["Revision"]), instant())
}

func (current *scenario) markOrderedAction(_ map[string]string) {
	current.cart, current.err = current.cart.MarkOrdered()
}

func (current *scenario) expireAction(_ map[string]string) {
	current.cart, current.err = current.cart.Expire(instant())
}

func (current *scenario) createOrderAction(row map[string]string) {
	current.createOrder(row["Operation"] == "createEmptyOrder")
}

func (current *scenario) configure(revision uint64, invalid bool) {
	checkout := sale.Checkout{ShippingAddress: sale.Address{Recipient: "Demo", Street: "Testweg 1", PostalCode: "12345", City: "Teststadt", Country: "DE"}, ShippingMethod: sale.Standard, PaymentMethod: sale.Invoice}
	if invalid {
		checkout.ShippingAddress.Recipient = ""
	}

	current.cart, current.err = current.cart.ConfigureCheckout(checkout, revision, instant())
}

func (current *scenario) setEvaluation(row map[string]string) {
	evaluation, err := (sale.PricedItems{{SKU: "PHONE", ProductID: "p-phone", Quantity: 1, UnitPrice: money("10000", "EUR"), InventoryRevision: 1}}).Evaluate(nil, sale.Standard)
	if err != nil {
		current.err = err
		return
	}

	evaluation.ID = value.EvaluationID(row["EvaluationID"])
	evaluation.CartRevision = unsigned(row["Revision"])
	evaluation.RuleSetVersion = row["RuleSetVersion"]
	current.cart, current.err = current.cart.SetEvaluation(evaluation)
}

func (current *scenario) createOrder(empty bool) {
	items := []sale.OrderItem{{SKU: "PHONE", ProductName: "Testmodell", Quantity: 1, UnitPrice: money("10000", "EUR")}}
	if empty {
		items = nil
	}

	current.order, current.err = sale.NewOrder("o-new", current.cart, items, instant())
}

func (current *scenario) purchaseResult(table *godog.Table) error {
	item := current.order.Items[0]

	return check(table, map[string]string{"OrderID": string(current.order.ID), "SKU": string(item.SKU), "Quantity": text(item.Quantity), "UnitPrice": text(item.UnitPrice.MinorUnits), "Total": text(current.order.Total.MinorUnits)})
}

func (current *scenario) orderExplicit(table *godog.Table) {
	row := rows(table)[0]
	items := []sale.OrderItem{}
	for index := int64(0); index < number(row["ItemCount"]); index++ {
		items = append(items, sale.OrderItem{SKU: value.SKU(row["SKU"]), Quantity: number(row["Quantity"]), UnitPrice: money(row["UnitPrice"], "EUR")})
	}

	current.cart.State = sale.CartState(row["State"])
	now := instant().Add(time.Duration(number(row["HoursSinceActivity"])) * time.Hour)
	current.order, current.err = sale.NewOrder(value.OrderID(input(row["OrderID"])), current.cart, items, now)
}

func (current *scenario) orderFromEvaluation(table *godog.Table) error {
	row := rows(table)[0]
	cart, err := sale.NewCart(value.CartID(row["CartID"]), value.UserID(row["UserID"]), instant())
	if err != nil {
		return err
	}

	current.cart = cart
	current.cart, current.err = current.cart.ConfigureCheckout(checkoutInput(row), current.cart.Revision, instant())
	if current.err != nil {
		return current.err
	}

	current.evaluation.ID = value.EvaluationID(row["EvaluationID"])
	current.evaluation.CartRevision = current.cart.Revision
	current.evaluation.RuleSetVersion = current.cart.RuleSetVersion

	return current.confirmOrder(value.OrderID(row["OrderID"]))
}

func (current *scenario) confirmOrder(id value.OrderID) error {
	current.cart, current.err = current.cart.SetEvaluation(current.evaluation)
	if current.err != nil {
		return current.err
	}

	current.cart, current.err = current.cart.Confirm(current.evaluation.ID, current.cart.Revision, instant())
	if current.err != nil {
		return current.err
	}

	items := []sale.OrderItem{}
	for _, item := range current.evaluation.Items {
		items = append(items, sale.OrderItem{SKU: item.SKU, Quantity: item.Quantity, UnitPrice: item.UnitPrice, Discounts: item.Discounts})
	}

	current.order, current.err = sale.NewOrder(id, current.cart, items, instant())

	return current.err
}

func (current *scenario) mutateDiscount(table *godog.Table) {
	row := rows(table)[0]
	current.evaluation.Items[1].Discounts[0].Amount.MinorUnits = number(row["Amount"])
	current.evaluation.Items[1].Discounts[0].ItemSKUs[0] = value.SKU(row["SKU"])
}

func (current *scenario) snapshotDiscount(table *godog.Table) error {
	discount := current.order.Items[1].Discounts[0]

	return check(table, map[string]string{"OrderID": string(current.order.ID), "RuleID": string(discount.RuleID), "Amount": text(discount.Amount.MinorUnits), "SKU": string(discount.ItemSKUs[0])})
}

func checkoutInput(row map[string]string) sale.Checkout {
	return sale.Checkout{ShippingAddress: sale.Address{Recipient: input(row["Recipient"]), Street: input(row["Street"]), PostalCode: row["PostalCode"], City: input(row["City"]), Country: value.CountryCode(row["Country"])}, ShippingMethod: sale.ShippingMethod(row["ShippingMethod"]), PaymentMethod: sale.PaymentMethod(row["PaymentMethod"])}
}
