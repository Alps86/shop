package catalog_test

import (
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) roleReadSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^Rollen ohne behauptete Kaufprüfung aufgelöst werden$`, current.readRoles)
	ctx.Step(`^lautet die ungeprüfte Rollenansicht$`, current.readRolesObserved)
}

func (current *world) readRoles(input *godog.Table) {
	data := tableRow(input)
	current.product = catalog.Product{Manufacturer: catalog.EconomicOperator{Name: data["Manufacturer"]}, Compliance: catalog.ProductCompliance{Context: catalog.ComplianceContext{VerificationStatus: data["VerificationStatus"]}}, Variants: []catalog.Variant{{SKU: value.SKU(data["ExistingSKU"])}}}
	current.roles, current.err = current.product.ResolveRoles(value.SKU(data["SKU"]))
}

func (current *world) readRolesObserved(input *godog.Table) error {
	data := tableRow(input)
	if err := current.checkCode(data); err != nil {
		return err
	}

	if err := equal(current.roles.Manufacturer.Name, data["Manufacturer"]); err != nil {
		return err
	}

	return equal(current.roles.Compliance.Context.VerificationStatus, data["VerificationStatus"])
}
