package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
	"time"
)

func (current *world) complianceSteps(context *godog.ScenarioContext) {
	context.Step(`^ein Pflichtpaket nur eine behauptete Vollständigkeit enthält$`, current.claimedCompleteness)
	context.Step(`^ein ausdrücklich fiktives Pflichtdatenpaket ist vorbereitet$`, current.prepareCompliance)
	context.Step(`^das Angebot aufgelöst wird$`, current.resolveOffer)
	context.Step(`^lautet die Pflichtdatenfreigabe$`, current.complianceResult)
	context.Step(`^ein synthetisches Prüfangebot validiert wird$`, current.validatePackage)
	context.Step(`^lautet das Prüfungsergebnis$`, current.packageResult)
	context.Step(`^ein mengenbezogenes Testangebot bewertet wird$`, current.unitPrice)
	context.Step(`^lautet das Grundpreisergebnis$`, current.unitPriceResult)
	context.Step(`^ausdrücklich fiktive Lieferantenprüfdaten sind vorbereitet$`, current.prepareRoles)
	context.Step(`^die Variante besitzt optionale eigenständige Rollen$`, current.variantRoles)
	context.Step(`^die Rollen für das Angebot aufgelöst werden$`, current.resolveRoles)
	context.Step(`^lauten die getrennten öffentlichen Rollen$`, current.rolesResult)
	current.energySteps(context)
	current.boundarySteps(context)
}

func (current *world) energySteps(context *godog.ScenarioContext) {
	context.Step(`^eine strukturierte Pflichtdatengruppe geprüft wird$`, current.chemicalBoundary)
	context.Step(`^ein öffentliches Pflichtdatenpaket mit einer Änderung geprüft wird$`, current.complianceBoundary)
	context.Step(`^ein ausdrücklich synthetisches Energieprüfpaket validiert wird$`, current.validateEnergy)
	context.Step(`^lautet der Prüfvertrag ohne reale Klassifizierungsbehauptung$`, current.energyResult)
	context.Step(`^ein ausdrücklich synthetisches Energiedokument zugeordnet wird$`, current.assignDocument)
	context.Step(`^lautet das Pflichtdatenergebnis$`, current.documentResult)
	context.Step(`^ein ausschließlich fiktives Prüfangebot hat ausdrücklich freigegebene Testhinweise$`, current.prepareNotice)
	context.Step(`^das öffentliche Pflichtdatenpaket gelesen wird$`, current.readNotice)
	context.Step(`^enthält es sichtbare Prüfhinweise unabhängig vom PDF$`, current.noticeResult)
}

func syntheticOperator() catalog.EconomicOperator {
	return catalog.EconomicOperator{Name: "Testhersteller", PostalAddress: catalog.BusinessAddress{Street: "Testweg 1", PostalCode: "10000", City: "Teststadt", Country: "DE"}, ElectronicAddress: "manufacturer@example.invalid"}
}

func syntheticCompliance() catalog.ProductCompliance {
	return catalog.ProductCompliance{Context: catalog.ComplianceContext{Market: "DE", Locale: "de-DE", VerificationStatus: "verified"}}
}

func (current *world) prepareCompliance(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.product.ID = value.ProductID(data["ProductID"])
	current.product.Variants = []catalog.Variant{{SKU: value.SKU(data["SKU"])}}
	current.prepareOfferRoles()
	from, err := timestamp(data["ValidFrom"])
	if err != nil {
		return err
	}

	until, err := timestamp(data["ValidUntil"])
	if err != nil {
		return err
	}

	current.product.Compliance.Context = catalog.ComplianceContext{Market: value.CountryCode(data["Market"]), Locale: value.Locale(data["Locale"]), ValidFrom: value.Optional[time.Time]{Present: true, Value: from}, ValidUntil: value.Optional[time.Time]{Present: true, Value: until}, VerificationStatus: data["VerificationStatus"]}
	current.removeRequiredField(data["MissingField"])

	return nil
}

func (current *world) removeRequiredField(field string) {
	mutations := map[string]func(){"Manufacturer.ElectronicAddress": current.removeElectronicAddress, "ResponsiblePerson": current.removeResponsiblePerson, "GermanRequiredNotice": current.removeGermanNotice}
	if mutation, exists := mutations[field]; exists {
		mutation()
	}
}

func (current *world) removeElectronicAddress() {
	current.product.Manufacturer.ElectronicAddress = ""
}

func (current *world) removeResponsiblePerson() {
	current.product.Manufacturer.PostalAddress.Country = "US"
	current.product.ResponsiblePerson = value.Optional[catalog.EconomicOperator]{}
}

func (current *world) removeGermanNotice() {
	current.product.Compliance.Notices = []catalog.RequiredNotice{{Code: "test", Text: catalog.LocalizedText{"en": "synthetic"}}}
}

func (current *world) resolveOffer(table *godog.Table) error {
	data := tableRow(table)
	current.rows = data
	at, err := timestamp(data["At"])
	if err != nil {
		return err
	}

	context := value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: "EUR", Market: value.CountryCode(data["Market"])}
	current.roles, current.err = current.product.ResolveCompliance(value.SKU(data["SKU"]), context, at)

	return nil
}

func (current *world) complianceResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	if err := equal(strconv.FormatBool(current.err == nil), data["Purchasable"]); err != nil {
		return err
	}

	fallback := false
	if current.err == nil {
		fallback = current.roles.Compliance.Context.Locale != value.Locale(current.rows["Locale"])
	}

	return equal(strconv.FormatBool(fallback), data["RequiredTextFallback"])
}

func (current *world) validatePackage(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.product.Manufacturer = syntheticOperator()
	current.product.Compliance = syntheticCompliance()
	packageData := syntheticCompliance()
	packageData.RequiredFields = list(data["RequiredFields"])
	current.supplyCompliance(&packageData, data["SuppliedFields"])
	current.rows = data
	if data["VariantOverride"] == "present" {
		current.product.Variants[0].Compliance = value.Optional[catalog.ProductCompliance]{Present: true, Value: packageData}
	}

	if data["VariantOverride"] == "absent" {
		current.product.Compliance = packageData
	}

	current.roles, current.err = current.product.ResolveCompliance("PHONE", value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))

	return nil
}

func (current *world) packageResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	resolved := strings.Join(current.roles.Compliance.RequiredFields, ",")

	return equal(resolved, data["ResolvedRequiredFields"])
}

func parseMeasurement(input string) catalog.Measurement {
	parts := strings.Split(input, ":")
	return catalog.Measurement{Value: decimal(parts[0]), Unit: parts[1]}
}

func (current *world) unitPrice(table *godog.Table) error {
	data := tableRow(table)
	price, err := strconv.ParseInt(data["Price"], 10, 64)
	if err != nil {
		return err
	}

	result, domainErr := parseMeasurement(data["NetContent"]).UnitPrice(value.Money{MinorUnits: price, Currency: value.Currency(data["Currency"])}, parseMeasurement(data["UnitPriceBasis"]))
	current.err = domainErr
	current.canonical = "absent"
	if domainErr == nil {
		current.canonical = strconv.FormatInt(result.MinorUnits, 10)
	}

	return nil
}

func (current *world) unitPriceResult(table *godog.Table) error {
	data := tableRow(table)
	if err := current.checkCode(data); err != nil {
		return err
	}

	if err := equal(current.canonical, data["UnitPrice"]); err != nil {
		return err
	}

	return equal("EUR/l", data["Unit"])
}

func (current *world) prepareRoles(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.product.Manufacturer = catalog.EconomicOperator{Name: data["Manufacturer.Name"], PostalAddress: catalog.BusinessAddress{Street: data["Manufacturer.Street"], PostalCode: data["Manufacturer.PostalCode"], City: data["Manufacturer.City"], Country: value.CountryCode(data["Manufacturer.Country"])}, ElectronicAddress: data["Manufacturer.ElectronicAddress"]}
	responsible := syntheticOperator()
	responsible.Name = data["ResponsiblePerson.Name"]
	responsible.PostalAddress.Country = value.CountryCode(data["ResponsiblePerson.Country"])
	responsible.ElectronicAddress = data["ResponsiblePerson.ElectronicAddress"]
	current.product.ResponsiblePerson = value.Optional[catalog.EconomicOperator]{Present: true, Value: responsible}
	current.product.Compliance = syntheticCompliance()
	current.product.Variants[0].SKU = "TEST"

	return nil
}

func (current *world) variantRoles(table *godog.Table) error {
	data := tableRow(table)
	manufacturer := syntheticOperator()
	manufacturer.Name = data["Manufacturer.Name"]
	responsible := syntheticOperator()
	responsible.Name = data["ResponsiblePerson.Name"]
	if data.truth("Manufacturer.Present") {
		current.product.Variants[0].Manufacturer = value.Optional[catalog.EconomicOperator]{Present: true, Value: manufacturer}
	}

	if data.truth("ResponsiblePerson.Present") {
		current.product.Variants[0].ResponsiblePerson = value.Optional[catalog.EconomicOperator]{Present: true, Value: responsible}
	}

	return nil
}

func (current *world) resolveRoles(table *godog.Table) error {
	data := tableRow(table)
	context := value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: "EUR", Market: value.CountryCode(data["Market"])}
	current.roles, current.err = current.product.ResolveCompliance(value.SKU(data["SKU"]), context, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))

	return current.err
}

func (current *world) rolesResult(table *godog.Table) error {
	data := tableRow(table)
	if err := equal(current.roles.Manufacturer.Name, data["Manufacturer.Name"]); err != nil {
		return err
	}

	return equal(current.roles.ResponsiblePerson.Value.Name, data["ResponsiblePerson.Name"])
}

func currentTime() time.Time {
	return time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
}

func (current *world) prepareOfferRoles() {
	current.product.Manufacturer = syntheticOperator()
	current.product.ResponsiblePerson = value.Optional[catalog.EconomicOperator]{Present: true, Value: syntheticOperator()}
	current.product.Compliance = syntheticCompliance()
}

func (current *world) claimedCompleteness(table *godog.Table) error {
	data := tableRow(table)
	current.store, current.product = validProduct()
	current.prepareOfferRoles()
	current.product.Compliance.RequiredFields = list(data["RequiredFields"])
	current.rows = data
	current.roles, current.err = current.product.ResolveCompliance("PHONE", value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, currentTime())

	return nil
}
