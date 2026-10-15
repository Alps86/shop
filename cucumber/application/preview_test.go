package application_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) registerPreview(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Zubehörangebot mit Rabattregel vorliegt$`, current.previewOffer)
	ctx.Step(`^eine reine Zubehörvorschau angefordert wird$`, current.requestPreview)
	ctx.Step(`^lautet die Vorschau ohne gespeicherte Cartänderung$`, current.previewObserved)
}

func (current *world) previewOffer(input *godog.Table) error {
	row := read(input)[0]
	price, err := strconv.ParseInt(row["UnitPrice"], 10, 64)
	if err != nil {
		return err
	}

	amount, err := strconv.ParseInt(row["Discount"], 10, 64)
	if err != nil {
		return err
	}

	sku := value.SKU(row["SKU"])
	current.inventoryProducts[sku] = value.ProductID(row["AccessoryID"])
	current.inventories[sku] = sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: 3, Revision: 1}
	current.previewRules = sale.DiscountRules{{ID: "rule-preview", Version: 1, MainProductID: value.ProductID(row["MainProductID"]), AccessoryID: value.ProductID(row["AccessoryID"]), Amount: value.Money{MinorUnits: amount, Currency: "EUR"}, Scope: sale.Pair}}

	return nil
}

func (current *world) requestPreview(input *godog.Table) {
	row := read(input)[0]
	current.fault = row["Fault"]
	current.preview, current.err = current.service.PreviewAccessory(context.Background(), application.Session(row["Session"]), value.CartID(row["CartID"]), value.SKU(row["SKU"]))
}

func (current *world) previewObserved(input *godog.Table) error {
	cart := current.carts["c-a"]
	discount := int64(0)
	for _, applied := range current.preview.Discounts {
		discount += applied.Amount.MinorUnits
	}

	return read(input).matches(map[string]string{"Code": value.Code(current.err), "Total": strconv.FormatInt(current.preview.Total.MinorUnits, 10), "Discount": strconv.FormatInt(discount, 10), "StoredMainQuantity": strconv.FormatInt(cart.Items.Quantity("PHONE"), 10), "StoredAccessoryQuantity": strconv.FormatInt(cart.Items.Quantity("CASE"), 10), "StoredRevision": strconv.FormatUint(cart.Revision, 10), "StoredConfirmed": strconv.FormatBool(cart.Checkout.Value.ConfirmedEvaluation.Present), "StoredActivityUnchanged": strconv.FormatBool(cart.LastActivityAt.Equal(current.now))})
}
