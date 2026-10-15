package postgres_test

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) prepare(input *godog.Table) error {
	row := table(input)
	price, err := strconv.ParseInt(row["Price"], 10, 64)
	if err != nil {
		return err
	}

	quantity, err := strconv.ParseInt(row["StockQuantity"], 10, 64)
	if err != nil {
		return err
	}

	product := newProduct(row)
	current.manifest = postgres.ImportManifest{DatasetVersion: row["DatasetVersion"], SourceURL: "https://example.invalid/synthetic", RetrievedAt: time.Now().UTC(), LicenseNote: "synthetic-only", Products: []catalog.Product{product}, Families: []catalog.ProductFamily{{ID: "f-phone", Version: 1}}, Categories: current.fixtureCategories(row), Inventories: []sale.Inventory{{SKU: value.SKU(row["SKU"]), Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: quantity, Revision: 1}}}

	return nil
}

func newProduct(row map[string]string) catalog.Product {
	manufacturer := catalog.EconomicOperator{Name: "Fiktiver Testhersteller", PostalAddress: catalog.BusinessAddress{Street: "Testweg 1", PostalCode: "12345", City: "Teststadt", Country: "DE"}, ElectronicAddress: "synthetic@example.invalid"}
	compliance := catalog.ProductCompliance{Context: catalog.ComplianceContext{Market: "DE", Locale: "de-DE", VerificationStatus: "verified"}}

	return catalog.Product{ID: value.ProductID(row["ProductID"]), FamilyID: "f-phone", FamilyVersion: 1, Name: catalog.LocalizedText{"de-DE": row["Name"]}, Brand: "Synthetic", PrimaryCategoryID: "mobile", Variants: []catalog.Variant{{SKU: value.SKU(row["SKU"])}}, Manufacturer: manufacturer, Compliance: compliance, Revision: 1, Published: true, Slug: "testmodell"}
}

func (current *world) publish() {
	current.imported, current.err = current.store.Import(context.Background(), current.manifest)
}

func (current *world) catalog() error {
	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	current.inventoryPrice = int64(len(snapshot.Products))

	return nil
}

func (current *world) importObserved(input *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	return check(table(input), map[string]string{"Code": value.Code(current.err), "ProductCount": strconv.FormatInt(current.inventoryPrice, 10)})
}

func (current *world) fixtureCategories(row map[string]string) []catalog.Category {
	raw := row["CategoryIDs"]
	if raw == "" {
		raw = "mobile"
	}

	categories := []catalog.Category{}
	for _, id := range strings.Split(raw, ",") {
		categories = append(categories, catalog.Category{ID: value.CategoryID(id), Version: 1})
	}

	return categories
}
