package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"maps"
	"strings"
)

func NewCatalog() Catalog {
	return Catalog{Products: map[value.ProductID]Product{}, Families: map[value.ProductFamilyID]ProductFamily{}, Categories: map[value.CategoryID]Category{}}
}

func (catalog Catalog) Publish(product Product) (Catalog, error) {
	if err := product.Validate(catalog); err != nil {
		return catalog, err
	}

	catalog.Products = maps.Clone(catalog.Products)
	product.Published = true
	product.Revision = catalog.Products[product.ID].Revision + 1
	catalog.Products[product.ID] = product

	return catalog, nil
}

func (product Product) Validate(catalog Catalog) error {
	if err := product.validateRequired(); err != nil {
		return err
	}

	if err := product.validateReferences(catalog); err != nil {
		return err
	}

	if err := product.validateVariants(catalog); err != nil {
		return err
	}

	return product.validateAttributes(catalog.Families[product.FamilyID])
}

func (product Product) validateRequired() error {
	for _, identifier := range []string{string(product.ID), string(product.FamilyID), string(product.PrimaryCategoryID)} {
		if err := value.IdentifierSyntax(identifier).Validate(); err != nil {
			return err
		}
	}

	if !product.Name.Valid() {
		return value.NewError("InvalidInput", "Name")
	}

	if strings.TrimSpace(product.Brand) == "" {
		return value.NewError("InvalidInput", "Brand")
	}

	if len(product.Variants) == 0 {
		return value.NewError("InvalidInput", "Variants")
	}

	return nil
}

func (product Product) validateReferences(catalog Catalog) error {
	family, exists := catalog.Families[product.FamilyID]
	if !exists {
		return value.NewError("InvalidReference", "FamilyID")
	}

	if family.Version != product.FamilyVersion {
		return value.NewError("Conflict", "FamilyVersion")
	}

	if _, exists := catalog.Categories[product.PrimaryCategoryID]; !exists {
		return value.NewError("InvalidReference", "PrimaryCategoryID")
	}

	if err := product.validateCategories(catalog); err != nil {
		return err
	}

	return product.validateAccessories(catalog)
}

func (product Product) validateCategories(catalog Catalog) error {
	seen := map[value.CategoryID]bool{product.PrimaryCategoryID: true}
	for _, id := range product.AdditionalCategoryIDs {
		if seen[id] {
			return value.NewError("InvalidReference", "AdditionalCategoryIDs")
		}

		if _, exists := catalog.Categories[id]; !exists {
			return value.NewError("InvalidReference", "AdditionalCategoryIDs")
		}

		seen[id] = true
	}

	return nil
}

func (product Product) validateAccessories(catalog Catalog) error {
	seen := map[value.ProductID]bool{product.ID: true}
	for _, accessory := range product.Accessories {
		if seen[accessory.ProductID] {
			return value.NewError("InvalidReference", "Accessories")
		}

		if !catalog.Products[accessory.ProductID].Published {
			return value.NewError("InvalidReference", "Accessories")
		}

		if err := product.validateCompatibility(accessory); err != nil {
			return err
		}

		seen[accessory.ProductID] = true
	}

	return nil
}

func (product Product) validateCompatibility(accessory AccessoryReference) error {
	seen := map[value.SKU]bool{}
	for _, sku := range accessory.CompatibleSKUs {
		if seen[sku] {
			return value.NewError("InvalidReference", "CompatibleSKUs")
		}

		if _, err := product.Variant(sku); err != nil {
			return value.NewError("InvalidReference", "CompatibleSKUs")
		}

		seen[sku] = true
	}

	return nil
}

func (product Product) Variant(sku value.SKU) (Variant, error) {
	for _, variant := range product.Variants {
		if variant.SKU == sku {
			return variant, nil
		}
	}

	return Variant{}, value.NewError("InvalidReference", "SKU")
}
