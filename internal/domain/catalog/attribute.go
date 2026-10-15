package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"slices"
)

func (definition AttributeDefinition) ValidateValue(attribute AttributeValue, level AttributeLevel) (AttributeValue, error) {
	if err := definition.Validate(); err != nil {
		return AttributeValue{}, err
	}

	if definition.Level != level {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Level")
	}

	if attribute.Type != definition.Type {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Type")
	}

	if !attribute.validPresence() {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Presence")
	}

	return definition.validateBranch(attribute)
}

func (attribute AttributeValue) branchCount() int {
	count := 0
	for _, present := range []bool{attribute.Text.Present, attribute.Number.Present, attribute.Measurement.Present, attribute.Boolean.Present, len(attribute.Options) > 0} {
		if present {
			count++
		}
	}

	return count
}

func (definition AttributeDefinition) validateBranch(attribute AttributeValue) (AttributeValue, error) {
	validators := map[AttributeType]func(AttributeValue) (AttributeValue, error){
		"text":        definition.validateText,
		"number":      definition.validateNumber,
		"measurement": definition.validateMeasurement,
		"boolean":     definition.validateBoolean,
		"select":      definition.validateSelection,
		"multiSelect": definition.validateSelection,
	}

	return validators[definition.Type](attribute)
}

func (definition AttributeDefinition) validateText(attribute AttributeValue) (AttributeValue, error) {
	if !attribute.Text.Value.Valid() {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Text")
	}

	return attribute, nil
}

func (definition AttributeDefinition) validateBoolean(attribute AttributeValue) (AttributeValue, error) {
	if !attribute.Boolean.Present {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Boolean")
	}

	return attribute, nil
}

func (definition AttributeDefinition) validateNumber(attribute AttributeValue) (AttributeValue, error) {
	if !attribute.Number.Present {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Number")
	}

	number, err := value.ParseDecimal(string(attribute.Number.Value))
	if err != nil {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Number")
	}

	if !definition.inBounds(number) {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Bounds")
	}

	attribute.Number.Value = number

	return attribute, nil
}

func (definition AttributeDefinition) inBounds(number value.Decimal) bool {
	return !definition.belowMinimum(number) && !definition.aboveMaximum(number)
}

func (definition AttributeDefinition) belowMinimum(number value.Decimal) bool {
	if !definition.Minimum.Present {
		return false
	}

	comparison, err := number.Compare(definition.Minimum.Value)

	return comparison < 0 || err != nil
}

func (definition AttributeDefinition) aboveMaximum(number value.Decimal) bool {
	if !definition.Maximum.Present {
		return false
	}

	comparison, err := number.Compare(definition.Maximum.Value)

	return comparison > 0 || err != nil
}

func (definition AttributeDefinition) validateMeasurement(attribute AttributeValue) (AttributeValue, error) {
	if !attribute.Measurement.Present {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Measurement")
	}

	converted, err := attribute.Measurement.Value.Convert(definition.Unit)
	if err != nil {
		return AttributeValue{}, err
	}

	if !definition.inBounds(converted.Value) {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Bounds")
	}

	attribute.Measurement.Value = converted

	return attribute, nil
}

func (definition AttributeDefinition) validateSelection(attribute AttributeValue) (AttributeValue, error) {
	if !definition.validSelectionCount(len(attribute.Options)) {
		return AttributeValue{}, value.NewError("InvalidAttribute", "Options")
	}

	allowed := map[OptionCode]bool{}
	for _, option := range definition.Options {
		allowed[option.Code] = true
	}

	selected := append(OptionCodes(nil), attribute.Options...)
	slices.Sort(selected)
	for index := range selected {
		if selected.invalid(allowed, index) {
			return AttributeValue{}, value.NewError("InvalidAttribute", "Options")
		}
	}

	attribute.Options = selected

	return attribute, nil
}

func (options OptionCodes) duplicate(index int) bool {
	if index == 0 {
		return false
	}

	return options[index-1] == options[index]
}

func (definition AttributeDefinition) validSelectionCount(count int) bool {
	if definition.Type == "select" {
		return count == 1
	}

	return count > 0 && count <= len(definition.Options)
}

func (options OptionCodes) invalid(allowed map[OptionCode]bool, index int) bool {
	if !allowed[options[index]] {
		return true
	}

	return options.duplicate(index)
}

func (attribute AttributeValue) validPresence() bool {
	errors := []error{attribute.Text.Validate(), attribute.Number.Validate(), attribute.Measurement.Validate(), attribute.Boolean.Validate()}
	for _, err := range errors {
		if err != nil {
			return false
		}
	}

	return attribute.branchCount() == 1
}
