package value

type ProductID string

type ProductFamilyID string

type CategoryID string

type CartID string

type UserID string

type OrderID string

type ReservationID string

type DiscountRuleID string

type ShipmentID string

type CommandID string

type EvaluationID string

type SKU string

type Currency string

type Locale string

type CountryCode string

type Decimal string

type Optional[T any] struct {
	Value   T
	Present bool
}

type Money struct {
	MinorUnits int64
	Currency   Currency
}

type RequestContext struct {
	Locale   Locale
	Currency Currency
	Market   CountryCode
}

type AttributeCode string

type OptionCode string

type IdentifierSyntax string

type RequiredJSON []byte
