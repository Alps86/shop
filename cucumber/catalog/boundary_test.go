package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) boundarySteps(context *godog.ScenarioContext) {
	context.Step(`^eine Kategorienkonfiguration veröffentlicht werden soll$`, current.categoryConfiguration)
	context.Step(`^eine Elternkategorie mit vorhandenem Kind gelöscht werden soll$`, current.categoryChild)
	context.Step(`^ein Attribut mit verdecktem Zweig validiert wird$`, current.concealed)
	context.Step(`^eine öffentliche Attributprüfung angefragt wird$`, current.attributeBoundary)
	context.Step(`^eine öffentliche Textprüfung angefragt wird$`, current.textBoundary)
	context.Step(`^lautet das öffentliche Textergebnis$`, current.publicTextResult)
	context.Step(`^ein vorbereitetes Produkt mit einer Katalogbeziehung publiziert wird$`, current.productBoundary)
	context.Step(`^eine Kategorie auf Entfernung geprüft wird$`, current.categoryRemoval)
	context.Step(`^eine öffentliche Kategorieprüfung angefragt wird$`, current.categoryBoundary)
}

func (current *world) attributeBoundary(table *godog.Table) error {
	data := tableRow(table)
	current.definition = catalog.AttributeDefinition{Code: "test", Type: "number", Level: "product"}
	current.attribute = catalog.AttributeValue{Type: "number", Number: value.Optional[value.Decimal]{Present: true, Value: "5"}}
	mutations := map[string]func(string) error{"attributeInvalid": current.attributeInvalid, "attributeOtherLevel": current.attributeOtherLevel, "attributeRequiredPresent": current.attributeRequiredPresent, "attributeMeasurementInvalid": current.attributeMeasurementInvalid, "definitionCode": current.definitionCode, "optionCode": current.optionCode, "concealedMinimum": current.concealedMinimum, "minimum": current.minimum, "maximum": current.maximum, "maximumOnly": current.maximumOnly, "invalidDefinition": current.invalidDefinition, "typeMismatch": current.typeMismatch, "booleanWrongBranch": current.booleanWrongBranch, "measurementWrongBranch": current.measurementWrongBranch, "measurementUnknownUnit": current.measurementUnknownUnit, "measurementInvalidNumber": current.measurementInvalidNumber, "numberInvalid": current.numberInvalid, "numberAboveMaximum": current.numberAboveMaximum, "numberAtMaximum": current.numberAboveMaximum, "numberWrongBranch": current.numberWrongBranch, "requiredProductMissing": current.requiredProductMissing, "attributeUnknown": current.attributeUnknown, "attributeValid": current.attributeValid, "familyInvalidID": current.familyInvalidID, "familyZeroVersion": current.familyZeroVersion}
	current.err = mutations[data["Case"]](data["Input"])

	return nil
}

func (current *world) definitionCode(input string) error {
	current.definition.Code = catalog.AttributeCode(input)
	return current.definition.Validate()
}

func (current *world) optionCode(input string) error {
	current.definition.Options = []catalog.AttributeOption{{Code: catalog.OptionCode(input)}}
	return current.definition.Validate()
}

func (current *world) minimum(input string) error {
	current.definition.Minimum = optionalDecimal(input)
	return current.definition.Validate()
}

func (current *world) maximum(input string) error {
	current.definition.Minimum = optionalDecimal("0")
	current.definition.Maximum = optionalDecimal(input)

	return current.definition.Validate()
}

func (current *world) maximumOnly(input string) error {
	current.definition.Maximum = optionalDecimal(input)
	return current.definition.Validate()
}

func (current *world) invalidDefinition(input string) error {
	current.definition.Type = catalog.AttributeType(input)
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) typeMismatch(input string) error {
	current.attribute.Type = catalog.AttributeType(input)
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) booleanWrongBranch(_ string) error {
	current.definition.Type = "boolean"
	current.attribute.Type = "boolean"
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) measurementWrongBranch(_ string) error {
	current.definition.Type = "measurement"
	current.attribute.Type = "measurement"
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) measurementUnknownUnit(input string) error {
	_, err := (catalog.Measurement{Value: "1", Unit: input}).Convert("ml")
	return err
}

func (current *world) measurementInvalidNumber(input string) error {
	_, err := (catalog.Measurement{Value: value.Decimal(input), Unit: "ml"}).Convert("ml")
	return err
}

func (current *world) numberInvalid(input string) error {
	current.attribute.Number.Value = value.Decimal(input)
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) numberAboveMaximum(input string) error {
	current.definition.Maximum = optionalDecimal("10")
	current.attribute.Number.Value = value.Decimal(input)
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) numberWrongBranch(_ string) error {
	current.attribute.Number = value.Optional[value.Decimal]{}
	current.attribute.Text = value.Optional[catalog.LocalizedText]{Present: true, Value: catalog.LocalizedText{"de-DE": "text"}}
	_, err := current.definition.ValidateValue(current.attribute, "product")

	return err
}

func (current *world) requiredProductMissing(_ string) error {
	current.definition.Required = true
	return catalog.Attributes{}.Validate(catalog.ProductFamily{Attributes: []catalog.AttributeDefinition{current.definition}}, "product")
}

func (current *world) attributeUnknown(input string) error {
	attributes := catalog.Attributes{catalog.AttributeCode(input): current.attribute}
	return attributes.Validate(catalog.ProductFamily{Attributes: []catalog.AttributeDefinition{current.definition}}, "product")
}

func (current *world) attributeValid(_ string) error {
	attributes := catalog.Attributes{"test": current.attribute}
	return attributes.Validate(catalog.ProductFamily{Attributes: []catalog.AttributeDefinition{current.definition}}, "product")
}

func (current *world) familyInvalidID(input string) error {
	return (catalog.ProductFamily{ID: value.ProductFamilyID(input), Version: 1}).Validate()
}

func (current *world) familyZeroVersion(_ string) error {
	return (catalog.ProductFamily{ID: "family"}).Validate()
}

func (current *world) textBoundary(table *godog.Table) error {
	data := tableRow(table)
	text := data["Text"]
	if text == "empty" {
		text = ""
	}

	current.text = catalog.LocalizedText{value.Locale(data["Locale"]): text}
	current.err = nil
	if !current.text.Valid() {
		current.err = value.NewError("InvalidInput", "Text")
	}

	current.resolved, current.source = current.text.Resolve(value.Locale(data["RequestedLocale"]))

	return nil
}

func (current *world) publicTextResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	if current.resolved == "" {
		current.resolved = "absent"
	}

	if current.source == "" {
		current.source = "absent"
	}

	return current.textResult(table)
}

func (current *world) productBoundary(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.store.Categories["c-extra"] = catalog.Category{ID: "c-extra", Version: 1}
	current.store.Products["p-accessory"] = catalog.Product{ID: "p-accessory", Published: true, Variants: []catalog.Variant{{SKU: "CABLE"}}}
	setters := map[string]func(string){"unpublishedAccessory": current.unpublishedAccessory, "additional": current.additional, "accessory": current.setAccessory, "accessoryCompatibility": current.compatibility, "duplicateAccessory": current.duplicateAccessory, "variantSKU": current.variantSKU, "gtin": current.setGTIN, "productAttribute": current.setProductAttribute, "variantAttribute": current.setVariantAttribute, "existingProduct": current.existingProduct, "republish": current.republish}
	setters[data["Case"]](data["Input"])
	current.store, current.err = current.store.Publish(current.product)

	return nil
}

func (current *world) additional(input string) {
	current.product.AdditionalCategoryIDs = []value.CategoryID{value.CategoryID(input)}
}

func (current *world) setAccessory(input string) {
	current.product.Accessories = []catalog.AccessoryReference{{ProductID: value.ProductID(input)}}
}

func (current *world) compatibility(input string) {
	current.setAccessory("p-accessory")
	for _, sku := range strings.Split(input, ",") {
		current.product.Accessories[0].CompatibleSKUs = append(current.product.Accessories[0].CompatibleSKUs, value.SKU(sku))
	}
}

func (current *world) duplicateAccessory(input string) {
	current.setAccessory(input)
	current.product.Accessories = append(current.product.Accessories, current.product.Accessories[0])
}

func (current *world) variantSKU(input string) {
	current.product.Variants[0].SKU = value.SKU(input)
}

func (current *world) setProductAttribute(input string) {
	current.product.Attributes = catalog.Attributes{catalog.AttributeCode(input): {Type: "text"}}
}

func (current *world) setVariantAttribute(input string) {
	current.product.Variants[0].Attributes = catalog.Attributes{catalog.AttributeCode(input): {Type: "text"}}
}

func (current *world) existingProduct(input string) {
	current.store.Products[value.ProductID(input)] = catalog.Product{ID: value.ProductID(input), Variants: []catalog.Variant{{SKU: "OTHER"}}}
}

func (current *world) republish(_ string) {
	current.store.Products[current.product.ID] = current.product
}

func (current *world) categoryRemoval(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	if data["Reference"] == "additional" {
		current.product.AdditionalCategoryIDs = []value.CategoryID{"c-extra"}
	}

	if data["Reference"] != "none" {
		current.store.Products[current.product.ID] = current.product
	}

	current.err = current.store.CanDeleteCategory("c-extra")

	return nil
}

func (current *world) categoryBoundary(table *godog.Table) error {
	data := tableRow(table)
	category := catalog.Category{ID: value.CategoryID(data["ID"]), Version: 1}
	if data["Parent"] != "absent" {
		category.ParentID = value.Optional[value.CategoryID]{Present: true, Value: value.CategoryID(data["Parent"])}
	}

	current.err = category.Validate(catalog.NewCatalog())

	return nil
}

func (current *world) concealed(table *godog.Table) error {
	data := tableRow(table)
	definition := catalog.AttributeDefinition{Code: "flag", Type: "boolean", Level: "product"}
	attribute := catalog.AttributeValue{Type: "boolean", Boolean: value.Optional[bool]{Present: true}}
	if data["Value"] != "absent" {
		attribute.Text = value.Optional[catalog.LocalizedText]{Present: data.truth("Present"), Value: catalog.LocalizedText{"de-DE": data["Value"]}}
	}

	_, current.err = definition.ValidateValue(attribute, "product")

	return nil
}

func (current *world) attributeInvalid(_ string) error {
	current.attribute.Number.Value = "invalid"
	return current.attributeValid("")
}

func (current *world) attributeOtherLevel(_ string) error {
	current.definition.Level = "variant"
	return catalog.Attributes{}.Validate(catalog.ProductFamily{Attributes: []catalog.AttributeDefinition{current.definition}}, "product")
}

func (current *world) attributeRequiredPresent(_ string) error {
	current.definition.Required = true
	return current.attributeValid("")
}

func (current *world) attributeMeasurementInvalid(_ string) error {
	current.definition.Type = "measurement"
	current.definition.Unit = "l"
	attribute := catalog.AttributeValue{Type: "measurement", Measurement: value.Optional[catalog.Measurement]{Present: true, Value: catalog.Measurement{Value: "1", Unit: "g"}}}
	_, err := current.definition.ValidateValue(attribute, "product")

	return err
}

func (current *world) concealedMinimum(input string) error {
	current.definition.Minimum = value.Optional[value.Decimal]{Value: value.Decimal(input)}
	return current.definition.Validate()
}

func (current *world) categoryConfiguration(table *godog.Table) error {
	data := tableRow(table)
	version, err := strconv.ParseUint(data["Version"], 10, 64)
	if err != nil {
		return err
	}

	store := catalog.NewCatalog()
	family := configuredFamily(data["ExistingAttributes"])
	store.Families[family.ID] = family
	category := catalog.Category{ID: "category", Version: version, FilterAttributes: attributeCodes(data["Filters"]), TopAttributes: attributeCodes(data["TopAttributes"])}

	current.err = category.Validate(store)

	return nil
}

func (current *world) categoryChild(table *godog.Table) error {
	data := tableRow(table)
	store := catalog.NewCatalog()
	parent := value.CategoryID(data["ParentID"])
	child := value.CategoryID(data["ChildID"])
	store.Categories[child] = catalog.Category{ID: child, ParentID: value.Optional[value.CategoryID]{Present: true, Value: parent}}
	current.err = store.CanDeleteCategory(parent)

	return nil
}

func (current *world) unpublishedAccessory(input string) {
	current.setAccessory(input)
	accessory := current.store.Products[value.ProductID(input)]
	accessory.Published = false
	current.store.Products[accessory.ID] = accessory
}

func configuredFamily(input string) catalog.ProductFamily {
	family := catalog.ProductFamily{ID: "family", Version: 1}
	for _, code := range list(input) {
		family.Attributes = append(family.Attributes, catalog.AttributeDefinition{Code: catalog.AttributeCode(code), Type: "text", Level: "product"})
	}

	return family
}

func attributeCodes(input string) []catalog.AttributeCode {
	result := []catalog.AttributeCode{}
	for _, code := range list(input) {
		result = append(result, catalog.AttributeCode(code))
	}

	return result
}
