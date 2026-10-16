package opensearch_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) correctionSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die öffentliche Wortquelle erneut projiziert wird$`, current.reviseWordSource)
	ctx.Step(`^die deutsche Suchbeschreibung projiziert wird$`, current.describeWordSource)
}

func (current *world) reviseWordSource(table *godog.Table) error {
	data := read(table)[0]
	id := value.ProductID(data["ProductID"])
	revision, err := strconv.ParseUint(data["Revision"], 10, 64)
	if err != nil {
		return err
	}

	product := current.products[id]
	product.Published = data["Published"] == "true"
	optional := value.Optional[catalog.Product]{}
	if data["Deleted"] != "true" {
		optional = value.Optional[catalog.Product]{Present: true, Value: product}
	}

	return current.adapter.Apply(context.Background(), application.ProjectionEvent{ProductID: id, Revision: revision, Product: optional, Inventory: current.inventory[id]})
}

func (current *world) describeWordSource(table *godog.Table) error {
	data := read(table)[0]
	id := value.ProductID(data["ProductID"])
	revision, err := strconv.ParseUint(data["Revision"], 10, 64)
	if err != nil {
		return err
	}

	product := current.products[id]
	product.Description = catalog.LocalizedText{current.config.Context.Locale: data["Description"]}

	return current.adapter.Apply(context.Background(), application.ProjectionEvent{ProductID: id, Revision: revision, Product: value.Optional[catalog.Product]{Present: true, Value: product}, Inventory: current.inventory[id]})
}
