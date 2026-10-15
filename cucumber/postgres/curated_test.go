package postgres_test

import (
	"strconv"

	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) curated(input *godog.Table) error {
	row := table(input)
	productFlag, firstErr := strconv.ParseBool(row["ProductFlag"])
	variantFlag, secondErr := strconv.ParseBool(row["VariantFlag"])
	if firstErr != nil {
		return firstErr
	}

	if secondErr != nil {
		return secondErr
	}

	current.manifest.Families[0].Attributes = []catalog.AttributeDefinition{{Code: "ready", Type: "boolean", Level: "product"}, {Code: "wireless", Type: "boolean", Level: "variant"}}
	current.manifest.Products[0].Attributes = catalog.Attributes{"ready": {Type: "boolean", Boolean: value.Optional[bool]{Present: true, Value: productFlag}}}
	current.manifest.Products[0].Variants[0].Attributes = catalog.Attributes{"wireless": {Type: "boolean", Boolean: value.Optional[bool]{Present: true, Value: variantFlag}}}
	current.manifest.Categories[0].FilterAttributes = []value.AttributeCode{"wireless"}
	current.manifest.Categories[0].TopAttributes = []value.AttributeCode{"ready", "wireless"}

	return nil
}
