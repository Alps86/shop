package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
)

type row map[string]string

type world struct {
	err        error
	definition catalog.AttributeDefinition
	attribute  catalog.AttributeValue
	store      catalog.Catalog
	product    catalog.Product
	text       catalog.LocalizedText
	resolved   string
	source     value.Locale
	canonical  string
	rows       row
	compliance catalog.ProductCompliance
	roles      catalog.ResolvedCompliance
	energy     catalog.EnergyLabel
}
