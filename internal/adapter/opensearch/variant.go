package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"maps"
	"net/url"
)

func NewVariant(product catalog.Product, variant catalog.Variant, inventory sale.Inventory, config Config) (Variant, error) {
	if err := inventory.Validate(); err != nil {
		return Variant{}, err
	}

	if inventory.Price.Currency != config.Context.Currency {
		return Variant{}, value.NewError("UnsupportedContext", "Currency")
	}

	projected := Variant{SKU: variant.SKU, Price: inventory.Price.MinorUnits, Available: inventory.AvailableQuantity(), Options: map[string][]value.OptionCode{}, Numbers: map[string]int64{}}
	projected.Card = NewCard(product, variant, inventory, config)
	projected.Available = projected.Card.AvailableQuantity
	projected.Margin = config.margin(inventory)
	attributes := maps.Clone(product.Attributes)
	if attributes == nil {
		attributes = catalog.Attributes{}
	}

	maps.Copy(attributes, variant.Attributes)

	return projected.withAttributes(attributes, config)
}

func NewCard(product catalog.Product, variant catalog.Variant, inventory sale.Inventory, config Config) application.ProductCard {
	name, _ := product.Name.Resolve(config.Context.Locale)
	card := application.ProductCard{ProductID: product.ID, Name: name, SKU: variant.SKU, UnitPrice: inventory.Price, AvailableQuantity: inventory.AvailableQuantity(), VariantCount: len(product.Variants), Compliance: product.Compliance}
	card.URL = "/" + string(config.Context.Locale) + "/produkt/" + url.PathEscape(product.Slug+"--"+string(product.ID)) + "?sku=" + url.QueryEscape(string(variant.SKU))
	card.Image = config.image(product, variant)
	card.TopAttributes = config.topAttributes(product, variant)
	resolved, err := product.ResolveCompliance(variant.SKU, config.Context, config.Clock.Now())
	if err != nil {
		card.AvailableQuantity = 0
		return card
	}

	card.Compliance = resolved.Compliance

	return card
}

func (config Config) image(product catalog.Product, variant catalog.Variant) catalog.Image {
	if len(variant.Images) > 0 {
		return variant.Images[0]
	}

	if len(product.Images) > 0 {
		return product.Images[0]
	}

	return catalog.Image{}
}

func (variant Variant) withAttributes(attributes catalog.Attributes, config Config) (Variant, error) {
	for code, attribute := range attributes {
		if err := variant.addAttribute(code, attribute, config); err != nil {
			return Variant{}, err
		}
	}

	return variant, nil
}

func (variant Variant) addAttribute(code value.AttributeCode, attribute catalog.AttributeValue, config Config) error {
	if len(attribute.Options) > 0 {
		variant.Options[string(code)] = attribute.Options
	}

	if attribute.Number.Present {
		number, err := attribute.Number.Value.Scaled()
		variant.Numbers[string(code)] = number

		return err
	}

	if attribute.Measurement.Present {
		return variant.addMeasurement(code, attribute.Measurement.Value, config)
	}

	return nil
}

func (variant Variant) addMeasurement(code value.AttributeCode, measurement catalog.Measurement, config Config) error {
	target := measurement.Unit
	if definition, exists := config.Schema.Attributes[code]; exists {
		target = definition.Unit
	}

	converted, err := measurement.Convert(target)
	if err != nil {
		return err
	}

	number, err := converted.Value.Scaled()
	variant.Numbers[string(code)] = number

	return err
}

func (config Config) topAttributes(product catalog.Product, variant catalog.Variant) map[value.AttributeCode]catalog.AttributeValue {
	attributes := maps.Clone(product.Attributes)
	if attributes == nil {
		attributes = catalog.Attributes{}
	}

	maps.Copy(attributes, variant.Attributes)
	result := map[value.AttributeCode]catalog.AttributeValue{}
	codes := config.Categories[product.PrimaryCategoryID].TopAttributes
	for _, code := range codes[:min(3, len(codes))] {
		if attribute, exists := attributes[code]; exists {
			result[code] = attribute
		}
	}

	return result
}
