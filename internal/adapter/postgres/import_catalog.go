package postgres

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
)

func (manifest ImportManifest) configure(config *catalog.Catalog) error {
	if err := manifest.configureFamilies(config); err != nil {
		return err
	}

	if err := manifest.configureCategories(config); err != nil {
		return err
	}

	return manifest.validateCategories(*config)
}

func (manifest ImportManifest) configureFamilies(config *catalog.Catalog) error {
	for _, family := range manifest.Families {
		if _, exists := config.Families[family.ID]; exists {
			return value.NewError("Conflict", "Families.ID")
		}

		if err := family.Validate(); err != nil {
			return err
		}

		config.Families[family.ID] = family
	}

	return nil
}

func (manifest ImportManifest) configureCategories(config *catalog.Catalog) error {
	for _, category := range manifest.Categories {
		if _, exists := config.Categories[category.ID]; exists {
			return value.NewError("Conflict", "Categories.ID")
		}

		config.Categories[category.ID] = category
	}

	return nil
}

func (manifest ImportManifest) validateCategories(config catalog.Catalog) error {
	for _, category := range manifest.Categories {
		if err := category.Validate(config); err != nil {
			return err
		}
	}

	return nil
}
