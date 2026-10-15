package catalog

import "github.com/Alps86/shop/internal/domain/value"

func (product Product) validateVariants(catalog Catalog) error {
	seen := map[value.SKU]bool{}
	for _, variant := range product.Variants {
		if err := variant.Validate(); err != nil {
			return err
		}

		if seen[variant.SKU] {
			return value.NewError("DuplicateSKU", "SKU")
		}

		if catalog.skuOwnedByOther(variant.SKU, product.ID) {
			return value.NewError("DuplicateSKU", "SKU")
		}

		seen[variant.SKU] = true
	}

	return nil
}

func (catalog Catalog) skuOwnedByOther(sku value.SKU, id value.ProductID) bool {
	for _, product := range catalog.Products {
		if product.ID == id {
			continue
		}

		if _, err := product.Variant(sku); err == nil {
			return true
		}
	}

	return false
}

func (product Product) validateAttributes(family ProductFamily) error {
	if err := product.Attributes.Validate(family, "product"); err != nil {
		return err
	}

	for _, variant := range product.Variants {
		if err := variant.Attributes.Validate(family, "variant"); err != nil {
			return err
		}
	}

	return nil
}

func (variant Variant) Validate() error {
	if err := value.IdentifierSyntax(string(variant.SKU)).Validate(); err != nil {
		return err
	}

	if variant.GTIN == "" {
		return nil
	}

	if !variant.validGTIN() {
		return value.NewError("InvalidInput", "GTIN")
	}

	return nil
}

func (variant Variant) validGTIN() bool {
	gtin := variant.GTIN
	lengths := map[int]bool{8: true, 12: true, 13: true, 14: true}
	if !lengths[len(gtin)] {
		return false
	}

	sum := 0
	for index := len(gtin) - 1; index >= 0; index-- {
		digit := gtin[index] - '0'
		if digit > 9 {
			return false
		}

		weight := 1 + 2*((len(gtin)-1-index)%2)
		sum += int(digit) * weight
	}

	return sum%10 == 0
}
