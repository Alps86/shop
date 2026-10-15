package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strings"
	"time"
)

func (current *world) attributeSteps(context *godog.ScenarioContext) {
	context.Step(`^eine Attributdefinition ist veröffentlicht$`, current.define)
	context.Step(`^ein Attributwert validiert wird$`, current.validateAttribute)
	context.Step(`^lautet das Attributergebnis$`, current.attributeResult)
	context.Step(`^ein Fachwert validiert wird$`, current.validateValue)
	context.Step(`^lautet das Wertergebnis$`, current.valueResult)
	context.Step(`^ein strukturierter Attributwert validiert wird$`, current.structuredAttribute)
	context.Step(`^lautet das strukturierte Attributergebnis$`, current.structuredResult)
	context.Step(`^eine Attributdefinition publiziert wird$`, current.publishDefinition)
	context.Step(`^lautet die Definitionsvalidierung$`, current.definitionResult)
	context.Step(`^ein Marketingtext besitzt Übersetzungen$`, current.marketingText)
	context.Step(`^eine Textauflösung angefragt wird$`, current.resolveText)
	context.Step(`^lautet das Textergebnis$`, current.textResult)
}

func (current *world) define(table *godog.Table) error {
	data := tableRow(table)
	current.definition = catalog.AttributeDefinition{Code: "test", Type: catalog.AttributeType(data["Type"]), Level: catalog.AttributeLevel(data["Level"]), Required: data.truth("Required")}
	current.definition.Minimum = optionalDecimal(data["Minimum"])
	current.definition.Maximum = optionalDecimal(data["Maximum"])
	for _, option := range list(data["Options"]) {
		current.definition.Options = append(current.definition.Options, catalog.AttributeOption{Code: catalog.OptionCode(option)})
	}

	return nil
}

func optionalDecimal(input string) value.Optional[value.Decimal] {
	if input == "absent" {
		return value.Optional[value.Decimal]{}
	}

	return value.Optional[value.Decimal]{Value: decimal(input), Present: true}
}

func (current *world) validateAttribute(table *godog.Table) error {
	data := tableRow(table)
	current.attribute = catalog.AttributeValue{Type: current.definition.Type, Number: value.Optional[value.Decimal]{Present: data.truth("Number.Present"), Value: decimal(data["Number.Value"])}, Boolean: value.Optional[bool]{Present: data.truth("Boolean.Present"), Value: data.truth("Boolean.Value")}}
	for _, option := range list(data["Options"]) {
		current.attribute.Options = append(current.attribute.Options, catalog.OptionCode(option))
	}

	if data.truth("ForeignBranch.Present") {
		current.attribute.Text = value.Optional[catalog.LocalizedText]{Present: true, Value: catalog.LocalizedText{"de-DE": "foreign"}}
	}

	if !data.truth("Present") {
		current.err = catalog.Attributes{}.Validate(catalog.ProductFamily{Attributes: []catalog.AttributeDefinition{current.definition}}, "variant")
		return nil
	}

	current.attribute = normalizeAttribute(current.attribute)
	current.attribute, current.err = current.definition.ValidateValue(current.attribute, catalog.AttributeLevel(data["Location"]))

	return nil
}

func (current *world) attributeResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	options := []string{}
	for _, option := range current.attribute.Options {
		options = append(options, string(option))
	}

	canonical := strings.Join(options, ",")
	if canonical == "" {
		canonical = "[]"
	}

	return equal(canonical, data["CanonicalOptions"])
}

func (current *world) validateValue(table *godog.Table) error {
	data := tableRow(table)
	handlers := map[string]func(row) error{"Measurement": current.measurementValue, "OptionalTime": current.optionalTime, "RequiredField": current.requiredField, "OptionalReplacement": current.optionalReplacement}
	current.err = nil
	current.canonical = "absent"

	return handlers[data["Type"]](data)
}

func (current *world) measurementValue(data row) error {
	parts := strings.Split(data["Input"], ":")
	result, err := (catalog.Measurement{Value: decimal(parts[0]), Unit: parts[1]}).Convert(data["TargetUnit"])
	current.err = err
	if err == nil {
		current.canonical = string(result.Value) + ":" + result.Unit
	}

	return nil
}

func (current *world) optionalTime(data row) error {
	parts := strings.Split(data["Input"], ",Value:")
	moment := value.Optional[time.Time]{Present: parts[0] == "Present:true"}
	if parts[1] != "zero" {
		parsed, err := timestamp(parts[1])
		if err != nil {
			return err
		}

		moment.Value = parsed
	}

	current.err = moment.Validate()
	if moment.Present {
		current.canonical = moment.Value.Format(time.RFC3339)
	}

	return nil
}

func (current *world) requiredField(data row) error {
	current.err = value.RequiredJSON([]byte(data["Input"])).Validate()
	return nil
}

func (current *world) optionalReplacement(data row) error {
	_, current.err = value.ParseOptionalReplacement[string]([]byte(data["Input"]))
	return nil
}

func (current *world) valueResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	return equal(current.canonical, data["Canonical"])
}

func (current *world) structuredAttribute(table *godog.Table) error {
	data := tableRow(table)
	definition := catalog.AttributeDefinition{Code: "test", Type: catalog.AttributeType(data["DefinitionType"]), Level: "product", Unit: data["CanonicalUnit"], Minimum: optionalDecimal(data["Minimum"]), Maximum: optionalDecimal(data["Maximum"])}
	text := catalog.LocalizedText{}
	if data["Text.de-DE"] != "empty" {
		text["de-DE"] = data["Text.de-DE"]
	}

	attribute := catalog.AttributeValue{Type: definition.Type, Text: value.Optional[catalog.LocalizedText]{Present: data.truth("Text.Present"), Value: text}, Measurement: value.Optional[catalog.Measurement]{Present: data.truth("Measurement.Present"), Value: catalog.Measurement{Value: decimal(data["Measurement.Value"]), Unit: data["Measurement.Unit"]}}, Number: value.Optional[value.Decimal]{Present: data.truth("Number.Present")}}
	current.attribute, current.err = definition.ValidateValue(normalizeAttribute(attribute), "product")

	return nil
}

func (current *world) structuredResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	text := "absent"
	measurement := "absent"
	if current.attribute.Text.Present {
		text = current.attribute.Text.Value["de-DE"]
	}

	if current.attribute.Measurement.Present {
		measurement = string(current.attribute.Measurement.Value.Value) + ":" + current.attribute.Measurement.Value.Unit
	}

	if err := equal(text, data["ResolvedText"]); err != nil {
		return err
	}

	return equal(measurement, data["CanonicalMeasurement"])
}

func (current *world) publishDefinition(table *godog.Table) error {
	data := tableRow(table)
	family := catalog.ProductFamily{ID: "test", Version: 1}
	for _, code := range list(data["Codes"]) {
		definition := catalog.AttributeDefinition{Code: catalog.AttributeCode(code), Type: catalog.AttributeType(data["Type"]), Level: catalog.AttributeLevel(data["Level"]), Minimum: optionalDecimal(data["Minimum"]), Maximum: optionalDecimal(data["Maximum"])}
		for _, option := range list(data["OptionCodes"]) {
			definition.Options = append(definition.Options, catalog.AttributeOption{Code: catalog.OptionCode(option)})
		}

		family.Attributes = append(family.Attributes, definition)
	}

	current.err = family.Validate()

	return nil
}

func (current *world) definitionResult(table *godog.Table) error {
	return current.checkCode(tableRow(table))
}

func (current *world) marketingText(table *godog.Table) error {
	current.text = catalog.LocalizedText{}
	for locale, text := range tableRow(table) {
		current.text[value.Locale(locale)] = text
	}

	return nil
}

func (current *world) resolveText(table *godog.Table) error {
	data := tableRow(table)
	current.resolved, current.source = current.text.Resolve(value.Locale(data["RequestedLocale"]))

	return nil
}

func (current *world) textResult(table *godog.Table) error {
	data := tableRow(table)
	if err := equal(current.resolved, data["Text"]); err != nil {
		return err
	}

	return equal(string(current.source), data["SourceLocale"])
}

func normalizeAttribute(attribute catalog.AttributeValue) catalog.AttributeValue {
	if !attribute.Number.Present {
		attribute.Number.Value = ""
	}

	if !attribute.Measurement.Present {
		attribute.Measurement.Value = catalog.Measurement{}
	}

	if !attribute.Text.Present {
		attribute.Text.Value = nil
	}

	return attribute
}
