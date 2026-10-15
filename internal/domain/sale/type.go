package sale

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"time"
)

type Inventory struct {
	SKU           value.SKU
	Price         value.Money
	StockQuantity int64
	Reservations  []Reservation
	Revision      uint64
}

type Reservation struct {
	ID       value.ReservationID
	OrderID  value.OrderID
	Quantity int64
}

type CartState string

type Cart struct {
	ID                value.CartID
	UserID            value.UserID
	Items             CartItems
	LastActivityAt    time.Time
	Revision          uint64
	State             CartState
	DefinitionVersion string
	RuleSetVersion    string
	Evaluation        value.Optional[CartEvaluation]
	Checkout          value.Optional[Checkout]
}

type CartItems []CartItem

type CartItem struct {
	SKU      value.SKU
	Quantity int64
}

type Checkout struct {
	CartID              value.CartID
	ShippingAddress     Address
	ShippingMethod      ShippingMethod
	PaymentMethod       PaymentMethod
	ConfirmedEvaluation value.Optional[CartEvaluation]
}

type Address struct {
	Recipient  string
	Street     string
	PostalCode string
	City       string
	Country    value.CountryCode
}

type ShippingMethod string

type PaymentMethod string

type DiscountScope string

type DiscountRule struct {
	ID            value.DiscountRuleID
	Version       uint64
	MainProductID value.ProductID
	AccessoryID   value.ProductID
	Amount        value.Money
	Scope         DiscountScope
}

type DiscountRules []DiscountRule

type PricedItem struct {
	SKU               value.SKU
	ProductID         value.ProductID
	Quantity          int64
	UnitPrice         value.Money
	InventoryRevision uint64
}

type PricedItems []PricedItem

type CartEvaluation struct {
	ID                 value.EvaluationID
	CartRevision       uint64
	EvaluatedAt        time.Time
	InventoryRevisions map[value.SKU]uint64
	RuleSetVersion     string
	Items              []EvaluatedCartItem
	Subtotal           value.Money
	Discounts          []AppliedDiscount
	ShippingCost       value.Money
	Taxes              []TaxAmount
	Total              value.Money
}

type EvaluatedCartItem struct {
	SKU       value.SKU
	Quantity  int64
	UnitPrice value.Money
	Discounts []AppliedDiscount
}

type AppliedDiscount struct {
	RuleID   value.DiscountRuleID
	Amount   value.Money
	ItemSKUs []value.SKU
}

type TaxAmount struct {
	Rate   value.Decimal
	Amount value.Money
}

type User struct {
	ID                value.UserID
	CategoryInterests []CategoryInterest
}

type CategoryInterest struct {
	CategoryID value.CategoryID
	Weight     value.Decimal
}

type Order struct {
	ID                value.OrderID
	CartID            value.CartID
	UserID            value.UserID
	PlacedAt          time.Time
	DefinitionVersion string
	RuleSetVersion    string
	Items             []OrderItem
	ShippingAddress   Address
	ShippingMethod    ShippingMethod
	PaymentMethod     PaymentMethod
	Subtotal          value.Money
	Discounts         []AppliedDiscount
	ShippingCost      value.Money
	Taxes             []TaxAmount
	Total             value.Money
}

type OrderItem struct {
	SKU               value.SKU
	ProductName       string
	VariantAttributes catalog.Attributes
	Quantity          int64
	UnitPrice         value.Money
	Discounts         []AppliedDiscount
}

type DeliveryStatus string

type FulfillmentUpdate struct {
	OrderID    value.OrderID
	ShipmentID value.ShipmentID
	CommandID  value.CommandID
	Lines      []FulfillmentLine
	Status     DeliveryStatus
}

type FulfillmentLine struct {
	SKU      value.SKU
	Quantity int64
}

type Fulfillment struct {
	OrderID   value.OrderID
	Purchased map[value.SKU]int64
	Updates   []FulfillmentUpdate
}

type discountAssignment struct {
	Items      PricedItems
	Used       map[value.SKU]int64
	Evaluation CartEvaluation
}

type AppliedDiscounts []AppliedDiscount

type EvaluatedCartItems []EvaluatedCartItem
