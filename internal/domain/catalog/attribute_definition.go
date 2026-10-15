package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"slices"
)

func (definition AttributeDefinition) Validate() error {
	validTypes := []AttributeType{"text", "number", "measurement", "select", "multiSelect", "boolean"}
	if !slices.Contains(validTypes, definition.Type) {
		return value.NewError("InvalidAttribute", "Type")
	}

	if !slices.Contains([]AttributeLevel{"product", "variant"}, definition.Level) {
		return value.NewError("InvalidAttribute", "Level")
	}

	if err := value.IdentifierSyntax(string(definition.Code)).Validate(); err != nil {
		return value.NewError("InvalidAttribute", "Code")
	}

	return definition.validateOptions()
}

func (definition AttributeDefinition) validateOptions() error {
	seen := map[OptionCode]bool{}
	for _, option := range definition.Options {
		if seen[option.Code] {
			return value.NewError("InvalidAttribute", "Options")
		}

		seen[option.Code] = true
		if err := value.IdentifierSyntax(string(option.Code)).Validate(); err != nil {
			return value.NewError("InvalidAttribute", "Options")
		}
	}

	return definition.validateBounds()
}

func (definition AttributeDefinition) validateBounds() error {
	if definition.unsupportedBounds() {
		return value.NewError("InvalidAttribute", "Bounds")
	}

	if err := definition.validateBound(definition.Minimum); err != nil {
		return err
	}

	return definition.validateMaximum()
}

func (definition AttributeDefinition) validateMaximum() error {
	if err := definition.validateBound(definition.Maximum); err != nil {
		return err
	}

	if definition.invalidBounds() {
		return value.NewError("InvalidAttribute", "Bounds")
	}

	return nil
}

func (definition AttributeDefinition) unsupportedBounds() bool {
	if definition.Type == "number" {
		return false
	}

	if definition.Type == "measurement" {
		return false
	}

	return definition.Minimum.Present || definition.Maximum.Present
}

func (definition AttributeDefinition) invalidBounds() bool {
	if !definition.Minimum.Present {
		return false
	}

	if !definition.Maximum.Present {
		return false
	}

	comparison, err := definition.Minimum.Value.Compare(definition.Maximum.Value)

	return comparison > 0 || err != nil
}

func (definition AttributeDefinition) validateBound(bound value.Optional[value.Decimal]) error {
	if err := bound.Validate(); err != nil {
		return value.NewError("InvalidAttribute", "Bounds")
	}

	if !bound.Present {
		return nil
	}

	if _, err := value.ParseDecimal(string(bound.Value)); err != nil {
		return value.NewError("InvalidAttribute", "Bounds")
	}

	return nil
}
