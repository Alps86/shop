package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
)

func (current *world) supplyCompliance(target *catalog.ProductCompliance, input string) {
	for _, field := range list(input) {
		current.supplyField(target, field)
	}
}

func (current *world) supplyField(target *catalog.ProductCompliance, field string) {
	if field == "RuleReference" {
		target.ApplicableRules = []catalog.RuleReference{{Code: "synthetic-rule", Market: "DE"}}
	}

	if field == "Reason" {
		target.NonApplicableReason = "Ausdrücklich synthetische Nichtanwendbarkeit"
	}

	if field == "Warnings" {
		target.Safety.Warnings = []catalog.RequiredNotice{{Code: "test", Text: catalog.LocalizedText{"de-DE": "Fiktiver Testhinweis"}}}
	}

	current.supplyEnergyMetadata(target, field)
}

func (current *world) supplyEnergyMetadata(target *catalog.ProductCompliance, field string) {
	if field == "Class" {
		target.EnergyLabel.Present = true
		target.EnergyLabel.Value.EfficiencyClass = "B"
	}

	if field == "Scale" {
		target.EnergyLabel.Value.ScaleMin = "G"
		target.EnergyLabel.Value.ScaleMax = "A"
	}

	if field == "Model" {
		target.EnergyLabel.Value.SupplierModelIdentifier = "SYNTHETIC-MODEL"
	}

	if field == "EPREL" {
		target.EnergyLabel.Value.EPRELRegistrationNumber = "synthetic-test-number"
	}

	current.supplyEnergyDocuments(target, field)
}

func (current *world) supplyEnergyDocuments(target *catalog.ProductCompliance, field string) {
	document := catalog.ProductDocument{SKU: "PHONE", Model: "SYNTHETIC-MODEL", Locale: "de-DE", URL: "https://example.invalid/synthetic.pdf"}
	if field == "Label" {
		target.EnergyLabel.Value.LabelDocuments = []catalog.ProductDocument{document}
	}

	if field == "InformationSheet" {
		target.EnergyLabel.Value.InformationSheets = []catalog.ProductDocument{document}
	}
}

func syntheticChemical() catalog.ChemicalLabel {
	return catalog.ChemicalLabel{ProductIdentifier: "SYNTHETIC-TEST", Supplier: syntheticOperator(), Pictograms: []catalog.HazardPictogram{{Code: "TEST-PIC", ImageURL: "https://example.invalid/test.svg", AltText: catalog.LocalizedText{"de-DE": "Fiktives Testpiktogramm"}}}, SignalWord: catalog.LocalizedText{"de-DE": "Fiktives Testsignalwort"}, HazardStatements: []catalog.CodedStatement{{Code: "TEST-STATEMENT", Text: catalog.LocalizedText{"de-DE": "Fiktive Testaussage"}}}}
}
