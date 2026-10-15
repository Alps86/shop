package sale

import (
	"slices"

	"github.com/Alps86/shop/internal/domain/value"
)

func (discounts AppliedDiscounts) aggregate() []AppliedDiscount {
	result := []AppliedDiscount{}
	indexes := map[value.DiscountRuleID]int{}
	for _, discount := range discounts {
		if index, exists := indexes[discount.RuleID]; exists {
			result[index] = result[index].merged(discount)
			continue
		}

		indexes[discount.RuleID] = len(result)
		result = append(result, discount.clone())
	}

	return result
}

func (discount AppliedDiscount) merged(other AppliedDiscount) AppliedDiscount {
	discount.Amount.MinorUnits += other.Amount.MinorUnits
	for _, sku := range other.ItemSKUs {
		if !slices.Contains(discount.ItemSKUs, sku) {
			discount.ItemSKUs = append(discount.ItemSKUs, sku)
		}
	}

	return discount
}
