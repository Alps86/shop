package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"math/big"
	"sort"
)

func (rule DiscountRule) Validate() error {
	if err := rule.Amount.Validate(); err != nil {
		return err
	}

	if rule.invalidIdentity() {
		return value.NewError("InvalidInput", "DiscountRule")
	}

	if rule.Amount.MinorUnits == 0 {
		return value.NewError("InvalidMoney", "Amount")
	}

	if rule.Scope != Pair {
		if rule.Scope != PerOrder {
			return value.NewError("InvalidInput", "Scope")
		}
	}

	return nil
}

func (rule DiscountRule) invalidIdentity() bool {
	return value.IdentifierSyntax(string(rule.ID)).Validate() != nil || rule.Version == 0 || rule.invalidProducts()
}

func (rule DiscountRule) invalidProducts() bool {
	return value.IdentifierSyntax(string(rule.MainProductID)).Validate() != nil || value.IdentifierSyntax(string(rule.AccessoryID)).Validate() != nil || rule.MainProductID == rule.AccessoryID
}

func (items PricedItems) Evaluate(rules DiscountRules, shipping ShippingMethod) (CartEvaluation, error) {
	evaluation, err := items.base(shipping)
	if err != nil {
		return evaluation, err
	}

	assignment := discountAssignment{Items: items, Used: map[value.SKU]int64{}, Evaluation: evaluation}
	sorted := append(DiscountRules{}, rules...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	for _, rule := range sorted {
		if err := assignment.apply(rule); err != nil {
			return CartEvaluation{}, err
		}
	}

	return assignment.Evaluation.finish()
}

func (items PricedItems) base(shipping ShippingMethod) (CartEvaluation, error) {
	cost, err := shipping.Cost()
	if err != nil {
		return CartEvaluation{}, err
	}

	result := CartEvaluation{Subtotal: value.Money{Currency: cost.Currency}, ShippingCost: cost, InventoryRevisions: map[value.SKU]uint64{}}
	for _, item := range items {
		if err := result.add(item); err != nil {
			return CartEvaluation{}, err
		}
	}

	return result, nil
}

func (evaluation *CartEvaluation) add(item PricedItem) error {
	if err := item.validate(); err != nil {
		return err
	}

	amount, err := item.UnitPrice.Multiply(item.Quantity)
	if err != nil {
		return err
	}

	subtotal, err := evaluation.Subtotal.Add(amount)
	if err != nil {
		return err
	}

	evaluation.Subtotal = subtotal
	evaluation.Items = append(evaluation.Items, EvaluatedCartItem{SKU: item.SKU, Quantity: item.Quantity, UnitPrice: item.UnitPrice})
	evaluation.InventoryRevisions[item.SKU] = item.InventoryRevision

	return nil
}

func (item PricedItem) validate() error {
	if item.Quantity <= 0 {
		return value.NewError("InvalidInput", "Quantity")
	}

	if value.IdentifierSyntax(string(item.SKU)).Validate() != nil {
		return value.NewError("InvalidInput", "SKU")
	}

	return item.UnitPrice.Validate()
}

func (assignment *discountAssignment) apply(rule DiscountRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	if rule.Amount.Currency != assignment.Evaluation.Subtotal.Currency {
		return value.NewError("InvalidMoney", "Currency")
	}

	main := rule.pairQuantity(assignment.Items)
	for index, item := range assignment.Items {
		if item.ProductID == rule.AccessoryID {
			main = assignment.assign(index, item, rule, main)
		}
	}

	return nil
}

func (rule DiscountRule) pairQuantity(items PricedItems) int64 {
	quantity := items.productQuantity(rule.MainProductID)
	if rule.Scope == PerOrder {
		return min(quantity, 1)
	}

	return quantity
}

func (items PricedItems) productQuantity(id value.ProductID) int64 {
	var quantity int64
	for _, item := range items {
		if item.ProductID == id {
			quantity += item.Quantity
		}
	}

	return quantity
}

func (assignment *discountAssignment) assign(index int, item PricedItem, rule DiscountRule, main int64) int64 {
	quantity := min(main, item.Quantity-assignment.Used[item.SKU])
	if quantity <= 0 {
		return main
	}

	amount := min(item.UnitPrice.MinorUnits, rule.Amount.MinorUnits) * quantity
	discount := AppliedDiscount{RuleID: rule.ID, Amount: value.Money{MinorUnits: amount, Currency: rule.Amount.Currency}, ItemSKUs: []value.SKU{item.SKU}}
	assignment.Evaluation.Items[index].Discounts = append(assignment.Evaluation.Items[index].Discounts, discount)
	assignment.Evaluation.Discounts = append(assignment.Evaluation.Discounts, discount)
	assignment.Used[item.SKU] += quantity

	return main - quantity
}

func (evaluation CartEvaluation) finish() (CartEvaluation, error) {
	total := evaluation.Subtotal.MinorUnits
	var taxes int64
	for _, item := range evaluation.Items {
		gross := item.Quantity*item.UnitPrice.MinorUnits - item.DiscountAmount()
		total -= item.DiscountAmount()
		taxes += evaluation.includedTax(gross)
	}

	amount, err := (value.Money{MinorUnits: total, Currency: evaluation.Subtotal.Currency}).Add(evaluation.ShippingCost)
	if err != nil {
		return CartEvaluation{}, err
	}

	evaluation.Total = amount
	evaluation.Taxes = []TaxAmount{{Rate: "19", Amount: value.Money{MinorUnits: taxes + evaluation.includedTax(evaluation.ShippingCost.MinorUnits), Currency: amount.Currency}}}
	evaluation.Discounts = AppliedDiscounts(evaluation.Discounts).aggregate()

	return evaluation, nil
}

func (evaluation CartEvaluation) includedTax(gross int64) int64 {
	numerator := new(big.Int).Mul(big.NewInt(gross), big.NewInt(19))
	numerator.Add(numerator, big.NewInt(59))

	return numerator.Div(numerator, big.NewInt(119)).Int64()
}

func (item EvaluatedCartItem) DiscountAmount() int64 {
	var amount int64
	for _, discount := range item.Discounts {
		amount += discount.Amount.MinorUnits
	}

	return amount
}
