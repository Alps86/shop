package value

import "math"

func NewMoney(amount int64, currency Currency) (Money, error) {
	canonical, err := ParseCurrency(string(currency))
	if err != nil {
		return Money{}, err
	}

	money := Money{MinorUnits: amount, Currency: canonical}

	return money, money.Validate()
}

func (money Money) Validate() error {
	_, err := ParseCurrency(string(money.Currency))
	if err != nil {
		return err
	}

	if money.MinorUnits < 0 {
		return NewError("InvalidMoney", "MinorUnits")
	}

	return nil
}

func (money Money) Add(other Money) (Money, error) {
	if err := money.compatible(other); err != nil {
		return Money{}, err
	}

	if money.MinorUnits > math.MaxInt64-other.MinorUnits {
		return Money{}, NewError("AmountOverflow", "overflow")
	}

	return NewMoney(money.MinorUnits+other.MinorUnits, money.Currency)
}

func (money Money) compatible(other Money) error {
	if err := money.Validate(); err != nil {
		return err
	}

	if err := other.Validate(); err != nil {
		return err
	}

	if money.Currency != other.Currency {
		return NewError("InvalidMoney", "Currency")
	}

	return nil
}

func (money Money) Multiply(quantity int64) (Money, error) {
	if err := money.Validate(); err != nil {
		return Money{}, err
	}

	if quantity < 0 {
		return Money{}, NewError("InvalidMoney", "Quantity")
	}

	if money.overflows(quantity) {
		return Money{}, NewError("AmountOverflow", "overflow")
	}

	return NewMoney(money.MinorUnits*quantity, money.Currency)
}

func (money Money) overflows(quantity int64) bool {
	if quantity == 0 {
		return false
	}

	return money.MinorUnits > math.MaxInt64/quantity
}
