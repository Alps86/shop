package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) importSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Produkt mit gemischten Pflichtdatenvarianten vorbereitet wird$`, current.mixedVariants)
	ctx.Step(`^ein Importfeld vor der Veröffentlichung geändert wird$`, current.mutateImport)
	ctx.Step(`^lautet die atomare Importbeobachtung$`, current.validationObserved)
}

func (current *world) mutateImport(input *godog.Table) error {
	row := table(input)
	mutations := map[string]func(string) error{"DatasetVersion": current.importVersion, "SourceURL": current.importSource, "LicenseNote": current.importLicense, "RetrievedAt": current.importTime, "FamilyVersion": current.importFamily, "CategoryParent": current.importParent, "Inventory": current.importInventory, "DocumentURL": current.importDocument, "SKU": current.importSKU, "ProductName": current.importName, "Discount": current.importDiscount, "FamilyReference": current.importReference, "AttributeLevel": current.importLevel, "Duplicate": current.importDuplicate}

	return mutations[row["Field"]](row["Value"])
}

func (current *world) validationObserved(input *godog.Table) error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	return check(table(input), map[string]string{"Code": value.Code(current.err), "Result": current.imported.Code, "Products": strconv.Itoa(len(snapshot.Products)), "Inventory": strconv.Itoa(len(snapshot.Inventory))})
}

func (current *world) importTime(_ string) error {
	current.manifest.RetrievedAt = time.Time{}
	return nil
}

func (current *world) importFamily(raw string) error {
	version, err := strconv.ParseUint(raw, 10, 64)
	current.manifest.Families[0].Version = version

	return err
}

func (current *world) importParent(raw string) error {
	current.manifest.Categories[0].ParentID = value.Optional[value.CategoryID]{Present: true, Value: value.CategoryID(raw)}
	return nil
}

func (current *world) importInventory(raw string) error {
	if raw == "orphan" {
		return current.orphanInventory()
	}

	if raw == "missing" {
		current.manifest.Inventories = nil
		return nil
	}

	if raw == "duplicate" {
		current.manifest.Inventories = append(current.manifest.Inventories, current.manifest.Inventories[0])
		return nil
	}

	return current.importStock(raw)
}

func (current *world) importStock(raw string) error {
	stock, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}

	current.manifest.Inventories[0].StockQuantity = stock

	return nil
}

func (current *world) importDocument(raw string) error {
	current.manifest.Products[0].Compliance.Documents = []catalog.ProductDocument{{Kind: "manual", Locale: "de-DE", SKU: "PHONE", URL: raw}}
	return nil
}

func (current *world) importSKU(raw string) error {
	current.manifest.Products[0].Variants = append(current.manifest.Products[0].Variants, catalog.Variant{SKU: value.SKU(raw)})
	return nil
}

func (current *world) importName(raw string) error {
	current.manifest.Products[0].Name["de-DE"] = raw
	return nil
}

func (current *world) importDiscount(raw string) error {
	amount, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}

	current.manifest.DiscountRules = []sale.DiscountRule{{ID: "rule-a", Version: 1, MainProductID: "p-phone", AccessoryID: "p-case", Amount: value.Money{MinorUnits: amount, Currency: "EUR"}, Scope: sale.Pair}}

	return nil
}

func (current *world) importVersion(raw string) error {
	current.manifest.DatasetVersion = raw
	return nil
}

func (current *world) importSource(raw string) error {
	current.manifest.SourceURL = raw
	return nil
}

func (current *world) importLicense(raw string) error {
	current.manifest.LicenseNote = raw
	return nil
}

func (current *world) mixedVariants(input *godog.Table) error {
	row := table(input)
	price, err := strconv.ParseInt(row["AdditionalPrice"], 10, 64)
	if err != nil {
		return err
	}

	stock, err := strconv.ParseInt(row["AdditionalStock"], 10, 64)
	if err != nil {
		return err
	}

	current.manifest.DatasetVersion = "demo-v2"
	current.manifest.Products[0].Compliance.Context.VerificationStatus = row["MainStatus"]
	compliance := current.manifest.Products[0].Compliance
	compliance.Context.VerificationStatus = row["AdditionalStatus"]
	sku := value.SKU(row["AdditionalSKU"])
	current.manifest.Products[0].Variants = append(current.manifest.Products[0].Variants, catalog.Variant{SKU: sku, Compliance: value.Optional[catalog.ProductCompliance]{Present: true, Value: compliance}})
	current.manifest.Inventories = append(current.manifest.Inventories, sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: stock, Revision: 1})

	return nil
}

func (current *world) importReference(raw string) error {
	current.manifest.Products[0].FamilyID = value.ProductFamilyID(raw)
	return nil
}

func (current *world) importLevel(raw string) error {
	current.manifest.Families[0].Attributes = []catalog.AttributeDefinition{{Code: "flag", Type: "boolean", Level: catalog.AttributeLevel(raw)}}
	current.manifest.Products[0].Attributes = map[value.AttributeCode]catalog.AttributeValue{"flag": {Type: "boolean", Boolean: value.Optional[bool]{Present: true, Value: false}}}

	return nil
}

func (current *world) importDuplicate(raw string) error {
	mutations := map[string]func(){"family": current.duplicateFamily, "category": current.duplicateCategory, "product": current.duplicateProduct}
	mutations[raw]()

	return nil
}

func (current *world) duplicateFamily() {
	current.manifest.Families = append(current.manifest.Families, current.manifest.Families[0])
}

func (current *world) duplicateCategory() {
	current.manifest.Categories = append(current.manifest.Categories, current.manifest.Categories[0])
}

func (current *world) duplicateProduct() {
	current.manifest.Products = append(current.manifest.Products, current.manifest.Products[0])
}

func (current *world) orphanInventory() error {
	inventory := current.manifest.Inventories[0]
	inventory.SKU = "ORPHAN"
	current.manifest.Inventories = append(current.manifest.Inventories, inventory)

	return nil
}
