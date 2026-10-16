package demo_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"os"
	"reflect"
)

func (current *world) importSearchProduct(table *godog.Table) error {
	data := read(table)[0]
	if err := current.prepareSearchImport(data); err != nil {
		return err
	}

	manifest, err := current.searchImportManifest(data)
	if err != nil {
		return err
	}

	return current.applySearchImport(manifest, data)
}

func (current *world) applySearchImport(manifest postgres.ImportManifest, data map[string]string) error {
	result, err := current.deployment.store.Import(context.Background(), manifest)
	if err != nil {
		return err
	}

	current.searchCache.product = value.ProductID(data["ProductID"])
	if result.Code != data["Code"] {
		return fmt.Errorf("Import %q statt %q", result.Code, data["Code"])
	}

	return nil
}

func (current *world) prepareSearchImport(data map[string]string) error {
	bytes, err := os.ReadFile("../../" + data["SeedFile"])
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &current.searchCache.seed); err != nil {
		return err
	}

	return current.prepareSearchBaseline(data)
}

func (current *world) prepareSearchBaseline(data map[string]string) error {
	baseline, err := current.deployment.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	current.searchCache.baseline = baseline
	current.searchCache.slot = current.deployment.config.Slot
	if current.deployment.external {
		current.searchCache.slot = data["ComposeSlot"]
	}

	current.searchCache.checkpoint, err = current.deployment.store.Checkpoint(context.Background(), current.searchCache.slot)

	return err
}

func (current *world) searchImportManifest(data map[string]string) (postgres.ImportManifest, error) {
	manifest := current.searchCache.seed.ImportManifest
	baseline := current.searchCache.baseline
	manifest.Products, manifest.Families, manifest.Categories = baseline.Products, baseline.Families, baseline.Categories
	manifest.Inventories, manifest.DiscountRules = baseline.Inventory, baseline.DiscountRules
	manifest.DatasetVersion = data["DatasetVersion"] + "-" + rand.Text()
	product, inventory, err := current.searchProduct(data)
	if err != nil {
		return manifest, err
	}

	manifest.Products = append(append([]catalog.Product{}, manifest.Products...), product)
	manifest.Inventories = current.searchImportInventories(inventory)

	return manifest, nil
}

func (current *world) searchImportInventories(inventory sale.Inventory) []sale.Inventory {
	baseline := current.searchCache.baseline.Inventory
	for _, original := range baseline {
		if original.SKU == inventory.SKU {
			return append([]sale.Inventory{}, baseline...)
		}
	}

	return append(append([]sale.Inventory{}, baseline...), inventory)
}

func (current *world) retainedSearchInventory(table *godog.Table) error {
	snapshot, err := current.deployment.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	sku := value.SKU(read(table)[0]["SKU"])
	for _, inventory := range snapshot.Inventory {
		if inventory.SKU == sku {
			return nil
		}
	}

	return fmt.Errorf("verbliebener Lagerdatensatz %q fehlt", sku)
}

func (current *world) searchProduct(data map[string]string) (catalog.Product, sale.Inventory, error) {
	for _, product := range current.searchCache.baseline.Products {
		if product.ID == value.ProductID(data["TemplateProductID"]) {
			return current.renameSearchProduct(product, data)
		}
	}

	return catalog.Product{}, sale.Inventory{}, fmt.Errorf("Vorlage %q fehlt", data["TemplateProductID"])
}

func (current *world) renameSearchProduct(product catalog.Product, data map[string]string) (catalog.Product, sale.Inventory, error) {
	if len(product.Variants) != 1 {
		return product, sale.Inventory{}, fmt.Errorf("Vorlage braucht genau eine Variante")
	}

	inventory, err := current.searchInventory(product.Variants[0].SKU)
	if err != nil {
		return product, inventory, err
	}

	product.ID = value.ProductID(data["ProductID"])
	product.Name = catalog.LocalizedText{value.Locale(data["Locale"]): data["Name"]}
	product.Slug = data["Slug"]
	product.Variants = append([]catalog.Variant{}, product.Variants...)
	product.Variants[0].SKU = value.SKU(data["SKU"])
	inventory.SKU = value.SKU(data["SKU"])
	inventory.Reservations = nil

	return product, inventory, nil
}

func (current *world) searchInventory(sku value.SKU) (sale.Inventory, error) {
	for _, inventory := range current.searchCache.baseline.Inventory {
		if inventory.SKU == sku {
			return inventory, nil
		}
	}

	return sale.Inventory{}, fmt.Errorf("Vorlagenbestand %q fehlt", sku)
}

func (current *world) deleteSearchProduct(table *godog.Table) error {
	id := value.ProductID(read(table)[0]["ProductID"])
	if id != current.searchCache.product {
		return fmt.Errorf("Löschung gehört nicht zum eingefügten Produkt")
	}

	checkpoint, err := current.deployment.store.Checkpoint(context.Background(), current.searchCache.slot)
	if err != nil {
		return err
	}

	current.searchCache.checkpoint = checkpoint

	return current.removeSearchProduct(id)
}

func (current *world) removeSearchProduct(id value.ProductID) error {
	if err := current.deployment.store.DeleteProduct(context.Background(), id); err != nil {
		return err
	}

	current.searchCache.product = ""

	return nil
}

func (current *world) observeSearchInventory(table *godog.Table) error {
	snapshot, err := current.deployment.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	if err := current.originalSearchInventory(snapshot.Inventory); err != nil {
		return err
	}

	for _, product := range snapshot.Products {
		if string(product.ID) == read(table)[0]["ProductID"] {
			return fmt.Errorf("gelöschtes Produkt bleibt im Katalog")
		}
	}

	return nil
}

func (current *world) originalSearchInventory(inventories []sale.Inventory) error {
	actual := make(map[value.SKU]sale.Inventory)
	for _, inventory := range inventories {
		actual[inventory.SKU] = inventory
	}

	for _, original := range current.searchCache.baseline.Inventory {
		if !reflect.DeepEqual(original, actual[original.SKU]) {
			return fmt.Errorf("Bestand/Reservierung für %s verändert", original.SKU)
		}
	}

	return nil
}

func (current *world) cleanupSearchImport() error {
	if current.searchCache.product == "" {
		return nil
	}

	return current.deployment.store.DeleteProduct(context.Background(), current.searchCache.product)
}
