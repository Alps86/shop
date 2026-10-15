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

func (current *world) productPresentation(input *godog.Table) {
	row := table(input)
	current.manifest.Products[0].Images = []catalog.Image{{URL: row["ProductImage"], AltText: catalog.LocalizedText{"de-DE": row["AltText"]}}}
	if row["VariantImage"] != "" {
		current.manifest.Products[0].Variants[0].Images = []catalog.Image{{URL: row["VariantImage"], AltText: catalog.LocalizedText{"de-DE": row["AltText"]}}}
	}

	child := catalog.Category{ID: value.CategoryID(row["ChildID"]), Version: 1, ParentID: value.Optional[value.CategoryID]{Present: true, Value: "mobile"}}
	current.manifest.Categories = append(current.manifest.Categories, child)
}
