package sale_test

import (
	"fmt"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *scenario) bindCart(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein offener Warenkorb enthält eine Position$`, current.cartGiven)
	ctx.Step(`^eine Warenkorboperation ausgeführt wird$`, current.cartOperate)
	ctx.Step(`^lautet das Warenkorbergebnis$`, current.cartResult)
	ctx.Step(`^eine abgeschlossene Bestellung besitzt Kaufwerte$`, current.orderGiven)
	ctx.Step(`^Katalog und Inventory später geändert werden$`, current.changeSources)
	ctx.Step(`^enthält die öffentliche Kaufansicht unveränderte Werte$`, current.orderResult)
	ctx.Step(`^ein Cart besitzt einen terminalen Zustand$`, current.terminalGiven)
	ctx.Step(`^eine Warenkorbänderung angefordert wird$`, current.terminalChange)
	ctx.Step(`^lautet die unveränderte Antwort$`, current.terminalResult)
}

func (current *scenario) cartGiven(table *godog.Table) error {
	row := rows(table)[0]
	cart, err := sale.NewCart("c-1", "u-1", instant())
	if err != nil {
		return err
	}

	cart.Items = sale.CartItems{{SKU: value.SKU(row["SKU"]), Quantity: number(row["Quantity"])}}
	for index := int64(1); index < number(row["PositionCount"]); index++ {
		cart.Items = append(cart.Items, sale.CartItem{SKU: value.SKU(fmt.Sprintf("extra-%d", index)), Quantity: 1})
	}

	cart.Revision = unsigned(row["CartRevision"])
	current.cart = cart
	current.inventory = sale.Inventory{SKU: value.SKU(row["SKU"]), StockQuantity: 99, Reservations: []sale.Reservation{{Quantity: number(row["Reserved"])}}}

	return nil
}

func (current *scenario) cartOperate(table *godog.Table) error {
	row := rows(table)[0]
	sku := value.SKU(row["SKU"])
	revision := unsigned(row["ExpectedRevision"])
	if row["Operation"] == "remove" {
		current.cart, current.err = current.cart.Remove(sku, revision, instant())
		return nil
	}

	if row["Operation"] == "add" {
		current.cart, current.err = current.cart.Add(sku, number(row["Quantity"]), revision, instant())
		return nil
	}

	current.cart, current.err = current.cart.SetQuantity(sku, number(row["Quantity"]), revision, instant())

	return nil
}

func (current *scenario) cartResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err), "QuantityPHONE": text(current.cart.Items.Quantity("PHONE")), "PositionCount": strconv.Itoa(len(current.cart.Items)), "Reserved": text(current.inventory.ReservedQuantity()), "ConfirmationPresent": strconv.FormatBool(current.cart.Checkout.Value.ConfirmedEvaluation.Present)})
}

func (current *scenario) orderGiven(table *godog.Table) error {
	row := rows(table)[0]
	current.rows = rows(table)
	attribute := catalog.AttributeValue{Type: "select", Options: []catalog.OptionCode{catalog.OptionCode(row["Color"])}}
	current.order = sale.Order{ID: value.OrderID(row["OrderID"]), DefinitionVersion: row["DefinitionVersion"], RuleSetVersion: row["RuleSetVersion"], ShippingMethod: sale.ShippingMethod(row["ShippingMethod"]), PaymentMethod: sale.PaymentMethod(row["PaymentMethod"])}
	current.order.Items = []sale.OrderItem{{SKU: value.SKU(row["SKU"]), ProductName: row["ProductName"], VariantAttributes: catalog.Attributes{"color": attribute}, Quantity: number(row["Quantity"]), UnitPrice: money(row["UnitPrice"], "EUR"), Discounts: []sale.AppliedDiscount{{Amount: money(row["Discount"], "EUR")}}}}
	current.order.ShippingAddress = sale.Address{Recipient: row["Recipient"], Street: row["Street"], PostalCode: row["PostalCode"], City: row["City"], Country: value.CountryCode(row["Country"])}
	current.order.ShippingCost = money(row["ShippingCost"], "EUR")
	current.order.Taxes = []sale.TaxAmount{{Rate: "19", Amount: money(row["Taxes"], "EUR")}}
	current.order.Total = money(row["Total"], "EUR")
	current.snapshot = current.order.Snapshot()

	return nil
}

func (current *scenario) changeSources(table *godog.Table) {
	row := rows(table)[0]
	current.order.Items[0].ProductName = row["ProductName"]
	current.order.Items[0].VariantAttributes["color"] = catalog.AttributeValue{Type: "select", Options: []catalog.OptionCode{catalog.OptionCode(row["Color"])}}
	current.order.Items[0].UnitPrice = money(row["UnitPrice"], "EUR")
}

func (current *scenario) orderResult(table *godog.Table) error {
	order := current.snapshot
	item := order.Items[0]
	result := map[string]string{"OrderID": string(order.ID), "SKU": string(item.SKU), "ProductName": item.ProductName, "Color": string(item.VariantAttributes["color"].Options[0]), "Quantity": text(item.Quantity), "UnitPrice": text(item.UnitPrice.MinorUnits), "Discount": text(item.Discounts[0].Amount.MinorUnits), "Recipient": order.ShippingAddress.Recipient, "Street": order.ShippingAddress.Street, "PostalCode": order.ShippingAddress.PostalCode, "City": order.ShippingAddress.City, "Country": string(order.ShippingAddress.Country), "DefinitionVersion": order.DefinitionVersion, "RuleSetVersion": order.RuleSetVersion, "ShippingMethod": string(order.ShippingMethod), "PaymentMethod": string(order.PaymentMethod), "ShippingCost": text(order.ShippingCost.MinorUnits), "Taxes": text(order.Taxes[0].Amount.MinorUnits), "Total": text(order.Total.MinorUnits)}

	return check(table, result)
}

func (current *scenario) terminalGiven(table *godog.Table) {
	row := rows(table)[0]
	current.cart = sale.Cart{ID: "c-1", UserID: "u-1", State: sale.CartState(row["State"]), Revision: unsigned(row["CartRevision"]), LastActivityAt: instant(), Items: make(sale.CartItems, number(row["ItemCount"]))}
}

func (current *scenario) terminalChange(table *godog.Table) {
	row := rows(table)[0]
	current.cart, current.err = current.cart.Add(value.SKU(row["SKU"]), number(row["Quantity"]), current.cart.Revision, instant())
}

func (current *scenario) terminalResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err), "State": string(current.cart.State), "CartRevision": strconv.FormatUint(current.cart.Revision, 10), "ItemCount": strconv.Itoa(len(current.cart.Items))})
}
