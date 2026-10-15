package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"strings"
)

func (energy EnergyLabel) Validate(sku value.SKU, locale value.Locale) error {
	for _, field := range []string{energy.SupplierModelIdentifier, energy.EPRELRegistrationNumber, energy.EfficiencyClass, energy.ScaleMin, energy.ScaleMax} {
		if strings.TrimSpace(field) == "" {
			return value.NewError("ComplianceIncomplete", "EnergyLabel")
		}
	}

	if !energy.validClass() {
		return value.NewError("ComplianceIncomplete", "EfficiencyClass")
	}

	if !energy.documentsMatch(energy.LabelDocuments, sku, locale) {
		return value.NewError("ComplianceIncomplete", "Label")
	}

	if !energy.documentsMatch(energy.InformationSheets, sku, locale) {
		return value.NewError("ComplianceIncomplete", "InformationSheet")
	}

	return nil
}

func (energy EnergyLabel) validClass() bool {
	if len(energy.EfficiencyClass) != 1 {
		return false
	}

	if len(energy.ScaleMin) != 1 {
		return false
	}

	if len(energy.ScaleMax) != 1 {
		return false
	}

	return energy.EfficiencyClass >= energy.ScaleMax && energy.EfficiencyClass <= energy.ScaleMin
}

func (energy EnergyLabel) documentsMatch(documents []ProductDocument, sku value.SKU, locale value.Locale) bool {
	for _, document := range documents {
		if document.matches(sku, locale, energy.SupplierModelIdentifier) {
			return true
		}
	}

	return false
}

func (document ProductDocument) matches(sku value.SKU, locale value.Locale, model string) bool {
	if document.SKU != sku {
		return false
	}

	if document.Model != model {
		return false
	}

	if document.Locale != locale {
		return false
	}

	return strings.HasPrefix(document.URL, "https://")
}
