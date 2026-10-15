package sale

import "slices"

func (evaluation CartEvaluation) MatchesPrices(other CartEvaluation) bool {
	return evaluation.sameAmounts(other) && evaluation.sameRules(other) && EvaluatedCartItems(evaluation.Items).Matches(other.Items)
}

func (evaluation CartEvaluation) sameAmounts(other CartEvaluation) bool {
	return evaluation.Subtotal == other.Subtotal && evaluation.ShippingCost == other.ShippingCost && evaluation.Total == other.Total
}

func (evaluation CartEvaluation) sameRules(other CartEvaluation) bool {
	return evaluation.RuleSetVersion == other.RuleSetVersion && slices.Equal(evaluation.Taxes, other.Taxes) && AppliedDiscounts(evaluation.Discounts).Matches(other.Discounts)
}

func (items EvaluatedCartItems) Matches(other []EvaluatedCartItem) bool {
	if len(items) != len(other) {
		return false
	}

	for _, item := range items {
		if !item.matchesIn(other) {
			return false
		}
	}

	return true
}

func (item EvaluatedCartItem) matchesIn(items []EvaluatedCartItem) bool {
	for _, other := range items {
		if item.SKU == other.SKU {
			return item.matches(other)
		}
	}

	return false
}

func (item EvaluatedCartItem) matches(other EvaluatedCartItem) bool {
	return item.Quantity == other.Quantity && item.UnitPrice == other.UnitPrice && AppliedDiscounts(item.Discounts).Matches(other.Discounts)
}

func (discounts AppliedDiscounts) Matches(other []AppliedDiscount) bool {
	return slices.EqualFunc(discounts, other, func(left, right AppliedDiscount) bool {
		return left.RuleID == right.RuleID && left.Amount == right.Amount && slices.Equal(left.ItemSKUs, right.ItemSKUs)
	})
}
