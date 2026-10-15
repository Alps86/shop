package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"strings"
	"time"
)

func (product Product) ResolveCompliance(sku value.SKU, context value.RequestContext, at time.Time) (ResolvedCompliance, error) {
	if err := context.Validate(); err != nil {
		return ResolvedCompliance{}, err
	}

	resolved, err := product.ResolveRoles(sku)
	if err != nil {
		return ResolvedCompliance{}, err
	}

	if err := resolved.Validate(sku, context, at); err != nil {
		return ResolvedCompliance{}, err
	}

	return resolved, nil
}

func (variant Variant) ResolveRoles(product Product) ResolvedCompliance {
	resolved := ResolvedCompliance{Manufacturer: product.Manufacturer, ResponsiblePerson: product.ResponsiblePerson, Compliance: product.Compliance}
	if variant.Compliance.Present {
		resolved.Compliance = variant.Compliance.Value
	}

	if variant.Manufacturer.Present {
		resolved.Manufacturer = variant.Manufacturer.Value
	}

	if variant.ResponsiblePerson.Present {
		resolved.ResponsiblePerson = variant.ResponsiblePerson
	}

	return resolved
}

func (resolved ResolvedCompliance) Validate(sku value.SKU, context value.RequestContext, at time.Time) error {
	if err := resolved.Compliance.Validate(sku, context, at); err != nil {
		return err
	}

	if !resolved.Manufacturer.Complete() {
		return value.NewError("ComplianceIncomplete", "Manufacturer")
	}

	if !resolved.Manufacturer.inEuropeanUnion() {
		if !resolved.ResponsiblePerson.Present {
			return value.NewError("ComplianceIncomplete", "ResponsiblePerson")
		}

		if !resolved.ResponsiblePerson.Value.validResponsiblePerson() {
			return value.NewError("ComplianceIncomplete", "ResponsiblePerson")
		}
	}

	return nil
}

func (operator EconomicOperator) Complete() bool {
	fields := []string{operator.Name, operator.PostalAddress.Street, operator.PostalAddress.PostalCode, operator.PostalAddress.City, string(operator.PostalAddress.Country), operator.ElectronicAddress}
	for _, field := range fields {
		if strings.TrimSpace(field) == "" {
			return false
		}
	}

	_, err := value.ParseCountryCode(string(operator.PostalAddress.Country))

	return err == nil
}

func (compliance ProductCompliance) Validate(sku value.SKU, context value.RequestContext, at time.Time) error {
	if !compliance.Context.validAt(context, at) {
		return value.NewError("ComplianceIncomplete", "Context")
	}

	for _, field := range compliance.RequiredFields {
		if !compliance.hasField(field) {
			return value.NewError("ComplianceIncomplete", field)
		}
	}

	if err := compliance.validateNotices(context.Locale); err != nil {
		return err
	}

	return compliance.validateLabels(sku, context.Locale)
}

func (context ComplianceContext) validAt(request value.RequestContext, at time.Time) bool {
	if context.VerificationStatus != "verified" {
		return false
	}

	if context.Market != request.Market {
		return false
	}

	if context.Locale != request.Locale {
		return false
	}

	if context.ValidFrom.Present {
		if at.Before(context.ValidFrom.Value) {
			return false
		}
	}

	return !context.expired(at)
}

func (context ComplianceContext) expired(at time.Time) bool {
	if !context.ValidUntil.Present {
		return false
	}

	return !at.Before(context.ValidUntil.Value)
}

func (compliance ProductCompliance) validateNotices(locale value.Locale) error {
	notices := append(append([]RequiredNotice(nil), compliance.Notices...), compliance.Safety.Warnings...)
	for _, notice := range notices {
		if strings.TrimSpace(notice.Text[locale]) == "" {
			return value.NewError("ComplianceIncomplete", "RequiredNotice")
		}
	}

	return nil
}

func (compliance ProductCompliance) validateLabels(sku value.SKU, locale value.Locale) error {
	if compliance.EnergyLabel.Present {
		if err := compliance.EnergyLabel.Value.Validate(sku, locale); err != nil {
			return err
		}
	}

	if compliance.ChemicalLabel.Present {
		return compliance.ChemicalLabel.Value.Validate(locale)
	}

	return nil
}

// Membership snapshot: https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=legissum:member_states (2026-10-15).
func (operator EconomicOperator) inEuropeanUnion() bool {
	countries := "|AT|BE|BG|HR|CY|CZ|DK|EE|FI|FR|DE|GR|HU|IE|IT|LV|LT|LU|MT|NL|PL|PT|RO|SK|SI|ES|SE|"
	return strings.Contains(countries, "|"+string(operator.PostalAddress.Country)+"|")
}

func (operator EconomicOperator) validResponsiblePerson() bool {
	if !operator.Complete() {
		return false
	}

	return operator.inEuropeanUnion()
}

func (product Product) ResolveRoles(sku value.SKU) (ResolvedCompliance, error) {
	variant, err := product.Variant(sku)
	if err != nil {
		return ResolvedCompliance{}, err
	}

	return variant.ResolveRoles(product), nil
}
