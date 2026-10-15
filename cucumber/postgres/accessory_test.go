package postgres_test

import (
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) accessorySteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein kompatibles synthetisches Zubehör mit Rabatt und Paarstatistik ergänzt wird$`, current.addAccessory)
	ctx.Step(`^eine Importregel oder Statistikreferenz geändert wird$`, current.statisticsMutation)
}

func (current *world) addAccessory(input *godog.Table) error {
	row := table(input)
	price, priceErr := strconv.ParseInt(row["Price"], 10, 64)
	stock, stockErr := strconv.ParseInt(row["Stock"], 10, 64)
	discount, discountErr := strconv.ParseInt(row["Discount"], 10, 64)
	if err := errors.Join(priceErr, stockErr, discountErr); err != nil {
		return err
	}

	current.installAccessory(row, price, stock, discount)

	return current.installStatistics(row)
}

func (current *world) installAccessory(row map[string]string, price, stock, discount int64) {
	product := current.manifest.Products[0]
	product.ID = value.ProductID(row["ProductID"])
	product.Name = catalog.LocalizedText{"de-DE": row["Name"]}
	product.Variants = []catalog.Variant{{SKU: value.SKU(row["SKU"])}}
	current.manifest.Products[0].Accessories = []catalog.AccessoryReference{{ProductID: product.ID, CompatibleSKUs: []value.SKU{"PHONE"}}}
	current.manifest.Products = append(current.manifest.Products, product)
	current.manifest.Inventories = append(current.manifest.Inventories, sale.Inventory{SKU: value.SKU(row["SKU"]), Price: value.Money{MinorUnits: price, Currency: "EUR"}, StockQuantity: stock, Revision: 1})
	current.manifest.DiscountRules = []sale.DiscountRule{{ID: "rule-pair", Version: 1, MainProductID: "p-phone", AccessoryID: product.ID, Amount: value.Money{MinorUnits: discount, Currency: "EUR"}, Scope: sale.Pair}}
}

func (current *world) installStatistics(row map[string]string) error {
	start, startErr := time.Parse(time.RFC3339, row["WindowStart"])
	end, endErr := time.Parse(time.RFC3339, row["WindowEnd"])
	count, countErr := strconv.ParseInt(row["BasketCount"], 10, 64)
	if err := errors.Join(startErr, endErr, countErr); err != nil {
		return err
	}

	current.manifest.CoPurchaseStatistics = application.PairStatistics{{SourceProductID: "p-phone", TargetProductID: value.ProductID(row["ProductID"]), BasketCount: count, WindowStart: start, WindowEnd: end, DatasetVersion: row["StatisticsVersion"]}}

	return nil
}
