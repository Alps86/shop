package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) validateEnergy(table *godog.Table) error {
	data := tableRow(table)
	current.rows = data
	current.energy = catalog.EnergyLabel{SupplierModelIdentifier: data["Model"], EfficiencyClass: data["EfficiencyClass"], ScaleMin: data["ScaleMin"], ScaleMax: data["ScaleMax"], EPRELRegistrationNumber: data["EPRELRegistrationNumber"]}
	current.energy.LabelDocuments = []catalog.ProductDocument{{SKU: value.SKU(data["LabelSKU"]), Model: data["Model"], URL: data["LabelURL"], Locale: value.Locale(data["DocumentLocale"]), Kind: "energyLabel"}}
	current.energy.InformationSheets = []catalog.ProductDocument{{SKU: value.SKU(data["InformationSheetSKU"]), Model: data["Model"], URL: data["InformationSheetURL"], Locale: value.Locale(data["DocumentLocale"]), Kind: "productInformationSheet"}}
	current.err = current.energy.Validate(value.SKU(data["SKU"]), value.Locale(data["DocumentLocale"]))

	return nil
}

func (current *world) energyResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	actual := row{"Class": current.energy.EfficiencyClass, "Scale": current.energy.ScaleMin + "-" + current.energy.ScaleMax, "LabelLinked": strconv.FormatBool(len(current.energy.LabelDocuments) > 0), "InformationSheetLinked": strconv.FormatBool(len(current.energy.InformationSheets) > 0), "ClaimsRealCertification": strconv.FormatBool(!current.syntheticEnergyFixture())}
	for key, expected := range data {
		if key == "Code" {
			continue
		}

		if err := equal(actual[key], expected); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) assignDocument(table *godog.Table) error {
	data := tableRow(table)
	current.energy = catalog.EnergyLabel{SupplierModelIdentifier: data["Model"], EPRELRegistrationNumber: "synthetic", EfficiencyClass: "B", ScaleMin: "G", ScaleMax: "A", LabelDocuments: []catalog.ProductDocument{{SKU: value.SKU(data["DocumentSKU"]), Model: data["DocumentModel"], URL: "https://example.invalid/test.pdf", Locale: "de-DE"}}}
	current.err = current.energy.Validate(value.SKU(data["OfferSKU"]), "de-DE")

	return nil
}

func (current *world) documentResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	return equal(strconv.FormatBool(current.err == nil), data["Purchasable"])
}

func (current *world) prepareNotice(table *godog.Table) error {
	data := tableRow(table)
	current.rows = data
	current.compliance = syntheticCompliance()
	current.compliance.Notices = []catalog.RequiredNotice{{Code: "test", Text: catalog.LocalizedText{"de-DE": data["NoticeText"]}, Placement: []catalog.DisplayContext{"listing", "detail"}}}
	current.compliance.ChemicalLabel = value.Optional[catalog.ChemicalLabel]{Present: true, Value: syntheticChemical()}
	current.compliance.ChemicalLabel.Value.Pictograms = []catalog.HazardPictogram{{Code: data["PictogramCode"], ImageURL: data["PictogramURL"], AltText: catalog.LocalizedText{"de-DE": data["AltText"]}}}

	return nil
}

func (current *world) readNotice() error {
	current.product = catalog.Product{Manufacturer: syntheticOperator(), Compliance: current.compliance, Variants: []catalog.Variant{{SKU: "TEST"}}}
	current.roles, current.err = current.product.ResolveCompliance("TEST", value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, currentTime())

	return current.err
}

func (current *world) noticeResult(table *godog.Table) error {
	data := tableRow(table)
	public := current.roles.Compliance
	actual := row{"NoticeText": public.Notices[0].Text["de-DE"], "PictogramCode": public.ChemicalLabel.Value.Pictograms[0].Code, "AltText": public.ChemicalLabel.Value.Pictograms[0].AltText["de-DE"], "VisibleWithoutPDF": strconv.FormatBool(len(public.Notices) > 0), "ClaimsRealHazardClassification": strconv.FormatBool(current.rows["DataClassification"] != "synthetic-only")}
	for key, expected := range data {
		if err := equal(actual[key], expected); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) complianceBoundary(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.product.Manufacturer = syntheticOperator()
	current.product.Compliance = syntheticCompliance()
	current.product.Manufacturer.PostalAddress.Country = "US"
	current.product.ResponsiblePerson = value.Optional[catalog.EconomicOperator]{Present: true, Value: syntheticOperator()}
	current.energy = catalog.EnergyLabel{SupplierModelIdentifier: "MODEL", EPRELRegistrationNumber: "synthetic", EfficiencyClass: "B", ScaleMin: "G", ScaleMax: "A", LabelDocuments: []catalog.ProductDocument{{SKU: "PHONE", Model: "MODEL", Locale: "de-DE", URL: "https://example.invalid/label.pdf"}}, InformationSheets: []catalog.ProductDocument{{SKU: "PHONE", Model: "MODEL", Locale: "de-DE", URL: "https://example.invalid/sheet.pdf"}}}
	current.mutateCompliance(data)
	sku := value.SKU("PHONE")
	if data["Field"] == "SKU" {
		sku = value.SKU(data["Value"])
	}

	current.roles, current.err = current.product.ResolveCompliance(sku, value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, currentTime())

	return nil
}

func (current *world) mutateCompliance(data row) {
	setters := map[string]func(string){"ResponsiblePerson.Country": current.responsibleCountry, "Manufacturer.Country": current.manufacturerCountry, "ResponsiblePerson.Name": current.responsibleName, "Context.Market": current.complianceMarket, "Context.Locale": current.complianceLocale, "Energy.Model": current.energyModel, "Energy.Class": current.energyClass, "Energy.ScaleMin": current.energyScaleMin, "Energy.ScaleMax": current.energyScaleMax, "Energy.LabelModel": current.energyLabelModel, "Energy.LabelLocale": current.energyLabelLocale, "Energy.LabelURL": current.energyLabelURL, "Energy.InformationSheet": current.energySheet}
	if setter, exists := setters[data["Field"]]; exists {
		setter(data["Value"])
	}

	if strings.HasPrefix(data["Field"], "Energy") {
		current.product.Compliance.EnergyLabel = value.Optional[catalog.EnergyLabel]{Present: true, Value: current.energy}
	}
}

func (current *world) responsibleName(_ string) {
	current.product.ResponsiblePerson.Value.Name = ""
}

func (current *world) complianceMarket(input string) {
	current.product.Compliance.Context.Market = value.CountryCode(input)
}

func (current *world) complianceLocale(input string) {
	current.product.Compliance.Context.Locale = value.Locale(input)
}

func (current *world) energyModel(_ string) {
	current.energy.SupplierModelIdentifier = ""
}

func (current *world) energyClass(input string) {
	if input == "empty" {
		input = ""
	}

	current.energy.EfficiencyClass = input
}

func (current *world) energyScaleMin(input string) {
	current.energy.ScaleMin = input
}

func (current *world) energyScaleMax(input string) {
	current.energy.ScaleMax = input
}

func (current *world) energyLabelModel(input string) {
	current.energy.LabelDocuments[0].Model = input
}

func (current *world) energyLabelLocale(input string) {
	current.energy.LabelDocuments[0].Locale = value.Locale(input)
}

func (current *world) energyLabelURL(input string) {
	current.energy.LabelDocuments[0].URL = input
}

func (current *world) energySheet(_ string) {
	current.energy.InformationSheets = nil
}

func (current *world) responsibleCountry(input string) {
	current.product.ResponsiblePerson.Value.PostalAddress.Country = value.CountryCode(input)
}

func (current *world) manufacturerCountry(input string) {
	current.product.Manufacturer.PostalAddress.Country = value.CountryCode(input)
	current.product.ResponsiblePerson = value.Optional[catalog.EconomicOperator]{}
}

func (current *world) syntheticEnergyFixture() bool {
	if !strings.HasPrefix(current.energy.EPRELRegistrationNumber, "synthetic-") {
		return false
	}

	for _, document := range append(append([]catalog.ProductDocument(nil), current.energy.LabelDocuments...), current.energy.InformationSheets...) {
		if !strings.HasPrefix(document.URL, "https://example.invalid/") {
			return false
		}
	}

	return true
}
