package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) productSteps(context *godog.ScenarioContext) {
	context.Step(`^die referenzierten Katalogdaten sind veröffentlicht$`, current.references)
	context.Step(`^ein Produkt veröffentlicht wird$`, current.publish)
	context.Step(`^lautet das Veröffentlichungsergebnis$`, current.publicationResult)
	context.Step(`^eine Katalogänderung validiert wird$`, current.catalogChange)
	context.Step(`^lautet das Validierungsergebnis$`, current.validatedResult)
	context.Step(`^ein ansonsten gültiges Produkt mit einem geänderten Feld publiziert wird$`, current.changedField)
	context.Step(`^lautet das Feldvalidierungsergebnis$`, current.validatedResult)
}

func (current *world) references(table *godog.Table) error {
	data := tableRow(table)
	current.store = catalog.NewCatalog()
	version, err := strconv.ParseUint(data["FamilyVersion"], 10, 64)
	if err != nil {
		return err
	}

	current.store.Families[value.ProductFamilyID(data["FamilyID"])] = catalog.ProductFamily{ID: value.ProductFamilyID(data["FamilyID"]), Version: version}
	current.store.Categories[value.CategoryID(data["CategoryID"])] = catalog.Category{ID: value.CategoryID(data["CategoryID"]), Version: 1}
	current.product = catalog.Product{Brand: data["Brand"], Name: catalog.LocalizedText{value.Locale(data["Locale"]): data["Name"]}}

	return nil
}

func (current *world) publish(table *godog.Table) error {
	data := tableRow(table)
	current.product.ID = value.ProductID(data["ID"])
	current.product.FamilyID = value.ProductFamilyID(data["FamilyID"])
	current.product.FamilyVersion = 1
	current.product.PrimaryCategoryID = value.CategoryID(data["PrimaryCategoryID"])
	current.product.Variants = nil
	for _, sku := range list(data["Variants"]) {
		current.product.Variants = append(current.product.Variants, catalog.Variant{SKU: value.SKU(sku)})
	}

	if data["ExistingSKUOwner"] != "none" {
		owner := value.ProductID(data["ExistingSKUOwner"])
		current.store.Products[owner] = catalog.Product{ID: owner, Variants: []catalog.Variant{{SKU: current.product.Variants[0].SKU}}}
	}

	current.store, current.err = current.store.Publish(current.product)

	return nil
}

func (current *world) publicationResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	products, variants := 0, 0
	if product, exists := current.store.Products[current.product.ID]; exists {
		products = 1
		variants = len(product.Variants)
	}

	if err := equal(strconv.Itoa(products), data["ProductCount"]); err != nil {
		return err
	}

	return equal(strconv.Itoa(variants), data["VariantCount"])
}

func validProduct() (catalog.Catalog, catalog.Product) {
	store := catalog.NewCatalog()
	store.Families["f-phone"] = catalog.ProductFamily{ID: "f-phone", Version: 1}
	store.Categories["c-phone"] = catalog.Category{ID: "c-phone", Version: 1}
	product := catalog.Product{ID: "p-phone", FamilyID: "f-phone", FamilyVersion: 1, PrimaryCategoryID: "c-phone", Brand: "Demo", Name: catalog.LocalizedText{"de-DE": "Testmodell"}, Variants: []catalog.Variant{{SKU: "PHONE"}}}

	return store, product
}

func (current *world) changedField(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	input := data["Value"]
	if input == "empty" {
		input = ""
	}

	setters := map[string]func(string){"Name.de-DE": current.setName, "Brand": current.setBrand, "ID": current.setID, "FamilyID": current.setFamily, "PrimaryCategoryID": current.setCategory, "Variant.GTIN": current.setGTIN}
	setters[data["Field"]](input)
	current.store, current.err = current.store.Publish(current.product)

	return nil
}

func (current *world) setName(input string) {
	current.product.Name["de-DE"] = input
}

func (current *world) setBrand(input string) {
	current.product.Brand = input
}

func (current *world) setID(input string) {
	current.product.ID = value.ProductID(input)
}

func (current *world) setFamily(input string) {
	current.product.FamilyID = value.ProductFamilyID(input)
}

func (current *world) setCategory(input string) {
	current.product.PrimaryCategoryID = value.CategoryID(input)
}

func (current *world) setGTIN(input string) {
	current.product.Variants[0].GTIN = input
}

func (current *world) validatedResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	return equal(strconv.FormatBool(current.err == nil), data["Published"])
}

func (current *world) catalogChange(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	operations := map[string]func(string) error{"categoryParent": current.categoryParent, "deleteCategory": current.deleteCategory, "accessory": current.accessory, "additionalCategory": current.additionalCategory, "familyVersion": current.familyVersion, "categoryAttributes": current.categoryAttributes}
	current.err = operations[data["Operation"]](data["Input"])

	return nil
}

func (current *world) categoryParent(input string) error {
	current.store = catalog.NewCatalog()
	for _, assignment := range strings.Split(input, ",") {
		parts := strings.Split(assignment, ":")
		category := catalog.Category{ID: value.CategoryID(parts[0]), Version: 1}
		if parts[1] != "none" {
			category.ParentID = value.Optional[value.CategoryID]{Present: true, Value: value.CategoryID(parts[1])}
		}

		current.store.Categories[category.ID] = category
	}

	for _, category := range current.store.Categories {
		if err := category.Validate(current.store); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) deleteCategory(input string) error {
	current.store.Products[current.product.ID] = current.product
	return current.store.CanDeleteCategory(value.CategoryID(strings.Split(input, ":")[1]))
}

func (current *world) accessory(input string) error {
	fields := parseAssignments(input)
	current.product.ID = value.ProductID(fields["source"])
	current.product.Accessories = []catalog.AccessoryReference{{ProductID: value.ProductID(fields["target"])}}

	return current.product.Validate(current.store)
}

func parseAssignments(input string) row {
	result := row{}
	for _, assignment := range strings.Split(input, ",") {
		parts := strings.SplitN(assignment, ":", 2)
		result[parts[0]] = parts[1]
	}

	return result
}

func (current *world) additionalCategory(input string) error {
	fields := parseAssignments(input)
	current.store.Categories[value.CategoryID(fields["primary"])] = catalog.Category{ID: value.CategoryID(fields["primary"]), Version: 1}
	current.product.PrimaryCategoryID = value.CategoryID(fields["primary"])
	current.product.AdditionalCategoryIDs = []value.CategoryID{value.CategoryID(fields["additional"])}

	return current.product.Validate(current.store)
}

func (current *world) familyVersion(input string) error {
	fields := parseAssignments(input)
	published, err := strconv.ParseUint(fields["published"], 10, 64)
	if err != nil {
		return err
	}

	family := current.store.Families[current.product.FamilyID]
	family.Version = published
	current.store.Families[family.ID] = family

	return current.product.Validate(current.store)
}

func (current *world) categoryAttributes(input string) error {
	fields := parseAssignments(input)
	filters, err := integer(fields["filters"])
	if err != nil {
		return err
	}

	top, err := integer(fields["top"])
	if err != nil {
		return err
	}

	category := catalog.Category{ID: "c-phone", Version: 1, FilterAttributes: make([]catalog.AttributeCode, filters), TopAttributes: make([]catalog.AttributeCode, top)}

	return category.Validate(current.store)
}
