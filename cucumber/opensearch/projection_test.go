package opensearch_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (current *world) projectionSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine synthetische Suchquelle mit vollständigen Lieferantendaten besteht$`, current.source)
	ctx.Step(`^der synthetische Produktkatalog wird tatsächlich projiziert$`, current.project)
	ctx.Step(`^lauten die tatsächlichen Modellkarten$`, current.cards)
}

func (current *world) Now() time.Time {
	return current.now
}

func (current *world) source(table *godog.Table) error {
	data := read(table)[0]
	data["Index"] += "-" + strings.ToLower(rand.Text()[:10])
	now, err := time.Parse(time.RFC3339, data["At"])
	if err != nil {
		return err
	}

	current.now = now
	current.products = map[value.ProductID]catalog.Product{}
	current.inventory = map[value.ProductID][]sale.Inventory{}
	current.config = opensearch.Config{Endpoint: os.Getenv("SHOP_TEST_OPENSEARCH_URL"), Index: data["Index"], Context: value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: value.Currency(data["Currency"]), Market: value.CountryCode(data["Market"])}, Clock: current}
	current.capture = &captureTransport{}
	current.config.HTTP = &http.Client{Transport: current.capture, Timeout: 5 * time.Second}
	current.template = catalog.Product{FamilyID: value.ProductFamilyID(data["FamilyID"]), FamilyVersion: 1, PrimaryCategoryID: value.CategoryID(data["CategoryID"]), Manufacturer: catalog.EconomicOperator{Name: data["Manufacturer"], PostalAddress: catalog.BusinessAddress{Street: data["Street"], PostalCode: data["PostalCode"], City: data["City"], Country: value.CountryCode(data["Country"])}, ElectronicAddress: data["Email"]}, Compliance: catalog.ProductCompliance{Context: catalog.ComplianceContext{Market: current.config.Context.Market, Locale: current.config.Context.Locale, VerificationStatus: "verified"}}}
	current.adapter = opensearch.New(current.config)

	return current.adapter.Initialize(context.Background())
}

func (current *world) project(table *godog.Table) error {
	for _, data := range read(table) {
		if err := current.addProduct(data); err != nil {
			return err
		}
	}

	return current.publish()
}

func (current *world) addProduct(data row) error {
	id := value.ProductID(data["ProductID"])
	product, exists := current.products[id]
	if !exists {
		product = current.template
		product.ID = id
		product.Name = catalog.LocalizedText{current.config.Context.Locale: data["Name"]}
		product.Slug = data["Slug"]
		product.Brand = data["Brand"]
		product.Published = data["Published"] == "true"
	}

	sku := value.SKU(data["SKU"])
	product.Variants = append(product.Variants, catalog.Variant{SKU: sku, Attributes: catalog.Attributes{"color": {Type: "select", Options: []catalog.OptionCode{catalog.OptionCode(data["Color"])}}, "storageCapacity": {Type: "number", Number: value.Optional[value.Decimal]{Present: true, Value: value.Decimal(data["StorageGB"])}}}})
	current.products[id] = product

	return current.addInventory(id, sku, data)
}

func (current *world) addInventory(id value.ProductID, sku value.SKU, data row) error {
	price, err := strconv.ParseInt(data["Price"], 10, 64)
	if err != nil {
		return err
	}

	stock, err := strconv.ParseInt(data["Stock"], 10, 64)
	if err != nil {
		return err
	}

	current.inventory[id] = append(current.inventory[id], sale.Inventory{SKU: sku, Price: value.Money{MinorUnits: price, Currency: current.config.Context.Currency}, StockQuantity: stock, Revision: 1})

	return nil
}

func (current *world) cards(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	ids := []string{}
	for _, card := range current.result.Cards {
		ids = append(ids, string(card.ProductID))
	}

	sort.Strings(ids)
	data := read(table)[0]
	if strings.Join(ids, ",") != data["ProductIDs"] {
		return fmt.Errorf("Modellkarten %v statt %s", ids, data["ProductIDs"])
	}

	if strconv.FormatInt(current.result.TotalProducts, 10) != data["TotalProducts"] {
		return fmt.Errorf("Modellanzahl %d statt %s", current.result.TotalProducts, data["TotalProducts"])
	}

	return nil
}
