package postgres

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"net/url"
)

func (manifest ImportManifest) validatePublication() error {
	if err := manifest.validateMetadata(); err != nil {
		return err
	}

	if err := manifest.validateInventory(); err != nil {
		return err
	}

	for _, rule := range manifest.DiscountRules {
		if err := rule.Validate(); err != nil {
			return err
		}
	}

	if err := manifest.validateRules(); err != nil {
		return err
	}

	return manifest.validateReferences()
}

func (manifest ImportManifest) validateMetadata() error {
	if err := manifest.validateSource(); err != nil {
		return err
	}

	if manifest.RetrievedAt.IsZero() {
		return value.NewError("InvalidInput", "RetrievedAt")
	}

	if manifest.LicenseNote == "" {
		return value.NewError("InvalidInput", "LicenseNote")
	}

	return nil
}

func (manifest ImportManifest) validateSource() error {
	source, err := url.Parse(manifest.SourceURL)
	if err != nil {
		return value.NewError("InvalidInput", "SourceURL")
	}

	if source.Scheme != "https" {
		return value.NewError("InvalidInput", "SourceURL")
	}

	if source.Host == "" {
		return value.NewError("InvalidInput", "SourceURL")
	}

	return nil
}

func (manifest ImportManifest) validateReferences() error {
	inventories := map[value.SKU]bool{}
	for _, inventory := range manifest.Inventories {
		if inventories[inventory.SKU] {
			return value.NewError("DuplicateSKU", "Inventories.SKU")
		}

		inventories[inventory.SKU] = true
	}

	if err := manifest.validateProductInventory(inventories); err != nil {
		return err
	}

	if len(inventories) > 0 {
		return value.NewError("InvalidReference", "Inventories.SKU")
	}

	return nil
}

func (manifest ImportManifest) validateProductInventory(inventories map[value.SKU]bool) error {
	for _, product := range manifest.Products {
		if err := manifest.validateVariantInventory(product, inventories); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) validateVariantInventory(product catalog.Product, inventories map[value.SKU]bool) error {
	for _, variant := range product.Variants {
		if !inventories[variant.SKU] {
			return value.NewError("InvalidReference", "Inventories.SKU")
		}

		delete(inventories, variant.SKU)
	}

	return nil
}
