package application

import (
	"slices"

	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
)

func (schema SearchSchema) validateValues(filters FilterSet, currency value.Currency) error {
	for code, options := range filters.Options {
		if !schema.validOptions(code, options) {
			return value.NewError("InvalidFilter", "Options")
		}
	}

	for code, interval := range filters.Ranges {
		if err := schema.validateRange(code, interval); err != nil {
			return err
		}
	}

	return filters.validatePrice(currency)
}

func (schema SearchSchema) validOptions(code value.AttributeCode, options []value.OptionCode) bool {
	if len(options) == 0 {
		return false
	}

	seen := map[value.OptionCode]bool{}
	for _, option := range options {
		if !schema.validOption(code, option, seen) {
			return false
		}

		seen[option] = true
	}

	return true
}

func (schema SearchSchema) validOption(code value.AttributeCode, option value.OptionCode, seen map[value.OptionCode]bool) bool {
	if seen[option] {
		return false
	}

	return slices.ContainsFunc(schema.Attributes[code].Options, func(candidate catalog.AttributeOption) bool {
		return candidate.Code == option
	})
}

func (schema SearchSchema) validateRange(code value.AttributeCode, interval NumericRange) error {
	definition := schema.Attributes[code]
	if !slices.Contains([]catalog.AttributeType{"number", "measurement"}, definition.Type) {
		return value.NewError("InvalidFilter", "RangeType")
	}

	if err := interval.Validate(); err != nil {
		return value.NewError("InvalidFilter", "Range")
	}

	return nil
}

func (interval NumericRange) Validate() error {
	minimum, err := interval.bound(interval.Minimum)
	if err != nil {
		return err
	}

	maximum, err := interval.bound(interval.Maximum)
	if err != nil {
		return err
	}

	if interval.reversed(minimum, maximum) {
		return value.NewError("InvalidFilter", "Range")
	}

	return nil
}

func (interval NumericRange) bound(bound value.Optional[value.Decimal]) (int64, error) {
	if err := bound.Validate(); err != nil {
		return 0, err
	}

	if !bound.Present {
		return 0, nil
	}

	return bound.Value.Scaled()
}

func (interval NumericRange) reversed(minimum, maximum int64) bool {
	return interval.Minimum.Present && interval.Maximum.Present && minimum > maximum
}

func (filters FilterSet) validatePrice(currency value.Currency) error {
	if err := filters.Price.Validate(); err != nil {
		return value.NewError("InvalidFilter", "Price")
	}

	if !filters.Price.Present {
		return nil
	}

	return filters.Price.Value.Validate(currency)
}

func (interval MoneyRange) Validate(currency value.Currency) error {
	for _, bound := range []value.Optional[value.Money]{interval.Minimum, interval.Maximum} {
		if err := interval.validateBound(bound, currency); err != nil {
			return err
		}
	}

	if interval.reversed() {
		return value.NewError("InvalidFilter", "Price")
	}

	return nil
}

func (interval MoneyRange) validateBound(bound value.Optional[value.Money], currency value.Currency) error {
	if err := bound.Validate(); err != nil {
		return value.NewError("InvalidFilter", "Price")
	}

	if !bound.Present {
		return nil
	}

	if bound.Value.Currency != currency {
		return value.NewError("InvalidFilter", "Currency")
	}

	if err := bound.Value.Validate(); err != nil {
		return value.NewError("InvalidFilter", "Price")
	}

	return nil
}

func (interval MoneyRange) reversed() bool {
	if !interval.Minimum.Present {
		return false
	}

	if !interval.Maximum.Present {
		return false
	}

	return interval.Minimum.Value.Currency != interval.Maximum.Value.Currency || interval.Minimum.Value.MinorUnits > interval.Maximum.Value.MinorUnits
}
