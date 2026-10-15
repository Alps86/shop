package opensearch_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) attributeSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine Produktprojektion mit expliziten Zusatzattributen übernommen wird$`, current.enrich)
	ctx.Step(`^die ergänzte Modellkarte entspricht$`, current.enrichedObserved)
}

func (current *world) enrich(table *godog.Table) {
	data := read(table)[0]
	id := value.ProductID(data["ProductID"])
	product := current.products[id]
	product.Attributes = catalog.Attributes{}
	product.AdditionalCategoryIDs = []value.CategoryID{product.PrimaryCategoryID}
	data.enrichProduct(&product)
	data.enrichVariant(&product.Variants[0])
	current.products[id] = product
	current.err = current.adapter.Apply(context.Background(), application.ProjectionEvent{ProductID: id, Revision: 2, Product: value.Optional[catalog.Product]{Present: true, Value: product}, Inventory: current.inventory[id]})
}

func (data row) enrichProduct(product *catalog.Product) {
	if data["ProductImage"] != "" {
		product.Images = []catalog.Image{{URL: data["ProductImage"]}}
	}

	if data["VerificationStatus"] != "" {
		product.Compliance.Context.VerificationStatus = data["VerificationStatus"]
	}

	if data["InvalidYear"] == "10000" {
		product.Compliance.Context.ValidFrom = value.Optional[time.Time]{Present: true, Value: time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)}
	}
}

func (data row) enrichVariant(variant *catalog.Variant) {
	if data["VariantImage"] != "" {
		variant.Images = []catalog.Image{{URL: data["VariantImage"]}}
	}

	if data["Measurement"] != "" {
		variant.Attributes["storageCapacity"] = catalog.AttributeValue{Type: "measurement", Measurement: value.Optional[catalog.Measurement]{Present: true, Value: catalog.Measurement{Value: value.Decimal(data["Measurement"]), Unit: data["Unit"]}}}
	}

	if data["Number"] != "" {
		variant.Attributes["storageCapacity"] = catalog.AttributeValue{Type: "number", Number: value.Optional[value.Decimal]{Present: true, Value: value.Decimal(data["Number"])}}
	}
}

func (current *world) enrichedObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	if len(current.result.Cards) == 0 {
		return value.NewError("MissingCard", "Cards")
	}

	card := current.result.Cards[0]

	return read(table)[0].matches(map[string]string{"Image": card.Image.URL, "AvailableQuantity": strconv.FormatInt(card.AvailableQuantity, 10), "TopAttributeCount": strconv.Itoa(len(card.TopAttributes))})
}
