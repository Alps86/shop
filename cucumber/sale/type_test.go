package sale_test

import (
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

type scenario struct {
	matched     bool
	inventory   sale.Inventory
	fulfillment sale.Fulfillment
	cart        sale.Cart
	order       sale.Order
	snapshot    sale.Order
	rules       sale.DiscountRules
	items       sale.PricedItems
	evaluation  sale.CartEvaluation
	money       value.Money
	err         error
	rows        []map[string]string
}
