package performance_test

import (
	"context"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (current *world) readSnapshotProduct(ctx context.Context) (snapshotProduct, string) {
	if current.store == nil {
		return snapshotProduct{}, "PostgreSQL observer not connected"
	}

	state, err := current.store.Catalog().Snapshot(ctx)
	if err != nil {
		return snapshotProduct{}, snapshotError(err)
	}

	for _, product := range state.Products {
		if product.ID == "bench-product-01" {
			return publicSnapshotProduct(product, state.Inventory), ""
		}
	}

	return snapshotProduct{}, "Product missing from subsequent PostgreSQL snapshot"
}

func publicSnapshotProduct(product catalog.Product, inventory []sale.Inventory) snapshotProduct {
	result := snapshotProduct{ID: product.ID, Revision: product.Revision, Published: product.Published, Name: product.Name, PrimaryCategoryID: product.PrimaryCategoryID, AdditionalCategoryIDs: product.AdditionalCategoryIDs, Context: product.Compliance.Context, ApplicableRules: product.Compliance.ApplicableRules}
	for _, variant := range product.Variants {
		validity := snapshotValidity{Context: variant.Compliance.Value.Context, ApplicableRules: variant.Compliance.Value.ApplicableRules}
		result.Variants = append(result.Variants, snapshotVariant{SKU: variant.SKU, Compliance: value.Optional[snapshotValidity]{Present: variant.Compliance.Present, Value: validity}})
		result.addInventory(variant.SKU, inventory)
	}

	return result
}

func (result *snapshotProduct) addInventory(sku value.SKU, inventory []sale.Inventory) {
	for _, item := range inventory {
		if item.SKU == sku {
			result.Inventory = append(result.Inventory, publicSnapshotInventory(item))
		}
	}
}

func publicSnapshotInventory(inventory sale.Inventory) snapshotInventory {
	result := snapshotInventory{SKU: inventory.SKU, Revision: inventory.Revision, Price: inventory.Price, StockQuantity: inventory.StockQuantity}
	for _, reservation := range inventory.Reservations {
		result.ReservedQuantity += reservation.Quantity
	}

	return result
}
