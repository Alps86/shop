package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) chemicalBoundary(table *godog.Table) error {
	data := tableRow(table)
	current.rows = data
	current.store, current.product = validProduct()
	current.prepareOfferRoles()
	current.compliance = syntheticCompliance()
	current.compliance.ChemicalLabel = value.Optional[catalog.ChemicalLabel]{Present: true, Value: syntheticChemical()}
	current.compliance.RequiredFields = []string{"VerifiedClassification", "Supplier", "Pictograms", "SignalWord", "Statements"}
	mutations := map[string]func(){"supplier": current.clearSupplier, "identifier": current.clearIdentifier, "signal": current.clearSignal, "pictograms": current.clearPictograms, "statements": current.clearStatements, "pictogramCode": current.clearPictogramCode, "pictogramURL": current.changePictogramURL, "pictogramAlt": current.clearPictogramAlt, "statementCode": current.clearStatementCode, "statementText": current.clearStatementText, "dataActInformation": current.dataAct, "DisposalNotice": current.disposalNotice, "ApplicableBatteryNotice": current.batteryNotice, "unknownRequirement": current.unknownRequirement}
	if mutation, exists := mutations[data["Case"]]; exists {
		mutation()
	}

	current.product.Compliance = current.compliance
	current.roles, current.err = current.product.ResolveCompliance("PHONE", value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, currentTime())

	return nil
}

func (current *world) clearSupplier() {
	current.compliance.ChemicalLabel.Value.Supplier.Name = ""
}

func (current *world) clearIdentifier() {
	current.compliance.ChemicalLabel.Value.ProductIdentifier = ""
}

func (current *world) clearSignal() {
	current.compliance.ChemicalLabel.Value.SignalWord = nil
}

func (current *world) clearPictograms() {
	current.compliance.ChemicalLabel.Value.Pictograms = nil
}

func (current *world) clearStatements() {
	current.compliance.ChemicalLabel.Value.HazardStatements = nil
}

func (current *world) clearPictogramCode() {
	current.compliance.ChemicalLabel.Value.Pictograms[0].Code = ""
}

func (current *world) changePictogramURL() {
	current.compliance.ChemicalLabel.Value.Pictograms[0].ImageURL = current.rows["Value"]
}

func (current *world) clearPictogramAlt() {
	current.compliance.ChemicalLabel.Value.Pictograms[0].AltText = nil
}

func (current *world) clearStatementCode() {
	current.compliance.ChemicalLabel.Value.HazardStatements[0].Code = ""
}

func (current *world) clearStatementText() {
	current.compliance.ChemicalLabel.Value.HazardStatements[0].Text = nil
}

func (current *world) dataAct() {
	current.compliance.RequiredFields = []string{"dataActInformation"}
	locale := value.Locale("de-DE")
	if current.rows["Value"] != "valid" {
		locale = value.Locale(current.rows["Value"])
	}

	current.compliance.Documents = []catalog.ProductDocument{{Kind: "instructions", Locale: "de-DE"}, {Kind: "dataActInformation", Locale: locale, URL: "https://example.invalid/data.pdf"}}
}

func (current *world) disposalNotice() {
	current.compliance.RequiredFields = []string{"DisposalNotice"}
	code := "DisposalNotice"
	if current.rows["Value"] != "valid" {
		code = current.rows["Value"]
	}

	current.compliance.Notices = []catalog.RequiredNotice{{Code: code, Text: catalog.LocalizedText{"de-DE": "Fiktiver Entsorgungshinweis"}}}
}

func (current *world) batteryNotice() {
	current.compliance.RequiredFields = []string{"ApplicableBatteryNotice"}
	current.compliance.Notices = []catalog.RequiredNotice{{Code: "ApplicableBatteryNotice", Text: catalog.LocalizedText{"de-DE": "Fiktiver Batteriehinweis"}}}
}

func (current *world) unknownRequirement() {
	current.compliance.RequiredFields = []string{"unknown"}
}
