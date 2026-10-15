package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"strings"
)

func (compliance ProductCompliance) hasField(field string) bool {
	energyFields := []string{"Class", "Scale", "Label", "InformationSheet", "Model", "EPREL"}
	for _, energyField := range energyFields {
		if field == energyField {
			return compliance.EnergyLabel.Present
		}
	}

	chemicalFields := []string{"VerifiedClassification", "Supplier", "Pictograms", "SignalWord", "Statements"}
	for _, chemicalField := range chemicalFields {
		if field == chemicalField {
			return compliance.ChemicalLabel.Present
		}
	}

	checks := map[string]bool{"Manufacturer": true, "Warnings": len(compliance.Safety.Warnings) > 0, "RuleReference": len(compliance.ApplicableRules) > 0, "Reason": strings.TrimSpace(compliance.NonApplicableReason) != "", "dataActInformation": compliance.hasDocument("dataActInformation"), "DisposalNotice": compliance.hasNotice("DisposalNotice"), "ApplicableBatteryNotice": compliance.hasNotice("ApplicableBatteryNotice")}

	return checks[field]
}

func (compliance ProductCompliance) hasDocument(kind DocumentKind) bool {
	for _, document := range compliance.Documents {
		if document.Kind == kind {
			return document.Locale == compliance.Context.Locale
		}
	}

	return false
}

func (compliance ProductCompliance) hasNotice(code string) bool {
	for _, notice := range compliance.Notices {
		if notice.Code == code {
			return strings.TrimSpace(notice.Text[compliance.Context.Locale]) != ""
		}
	}

	return false
}

func (chemical ChemicalLabel) Validate(locale value.Locale) error {
	if !chemical.Supplier.Complete() {
		return value.NewError("ComplianceIncomplete", "Chemical.Supplier")
	}

	if strings.TrimSpace(chemical.ProductIdentifier) == "" {
		return value.NewError("ComplianceIncomplete", "Chemical.ProductIdentifier")
	}

	return chemical.validateVisibleLabel(locale)
}

func (chemical ChemicalLabel) validateVisibleLabel(locale value.Locale) error {
	if strings.TrimSpace(chemical.SignalWord[locale]) == "" {
		return value.NewError("ComplianceIncomplete", "Chemical.SignalWord")
	}

	if len(chemical.Pictograms) == 0 {
		return value.NewError("ComplianceIncomplete", "Chemical.Pictograms")
	}

	if len(chemical.HazardStatements) == 0 {
		return value.NewError("ComplianceIncomplete", "Chemical.Statements")
	}

	return chemical.validatePictograms(locale)
}

func (chemical ChemicalLabel) validatePictograms(locale value.Locale) error {
	for _, pictogram := range chemical.Pictograms {
		if pictogram.Code == "" {
			return value.NewError("ComplianceIncomplete", "Pictogram.Code")
		}

		if !strings.HasPrefix(pictogram.ImageURL, "https://") {
			return value.NewError("ComplianceIncomplete", "Pictogram.ImageURL")
		}

		if pictogram.AltText[locale] == "" {
			return value.NewError("ComplianceIncomplete", "Pictogram.AltText")
		}
	}

	return chemical.validateStatements(locale)
}

func (chemical ChemicalLabel) validateStatements(locale value.Locale) error {
	statements := append(append(append([]CodedStatement(nil), chemical.HazardStatements...), chemical.PrecautionaryStatements...), chemical.SupplementalStatements...)
	for _, statement := range statements {
		if statement.Code == "" {
			return value.NewError("ComplianceIncomplete", "Statement.Code")
		}

		if strings.TrimSpace(statement.Text[locale]) == "" {
			return value.NewError("ComplianceIncomplete", "Statement.Text")
		}
	}

	return nil
}
