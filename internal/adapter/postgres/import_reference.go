package postgres

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
)

func (manifest ImportManifest) validateRules() error {
	products := map[value.ProductID]bool{}
	for _, product := range manifest.Products {
		if products[product.ID] {
			return value.NewError("Conflict", "Products.ProductID")
		}

		products[product.ID] = true
	}

	for _, rule := range manifest.DiscountRules {
		if !products[rule.MainProductID] {
			return value.NewError("InvalidReference", "DiscountRule.MainProductID")
		}

		if !products[rule.AccessoryID] {
			return value.NewError("InvalidReference", "DiscountRule.AccessoryID")
		}
	}

	return manifest.validateStatistics(products)
}

func (manifest ImportManifest) validateStatistics(products map[value.ProductID]bool) error {
	for _, statistic := range manifest.CoPurchaseStatistics {
		if !products[statistic.SourceProductID] {
			return value.NewError("InvalidReference", "Statistics.SourceProductID")
		}

		if !products[statistic.TargetProductID] {
			return value.NewError("InvalidReference", "Statistics.TargetProductID")
		}

		if err := manifest.validateStatistic(statistic); err != nil {
			return err
		}
	}

	return nil
}

func (manifest ImportManifest) validateStatistic(statistic application.PairStatistic) error {
	if statistic.BasketCount < 0 {
		return value.NewError("InvalidInput", "Statistics.BasketCount")
	}

	if !statistic.WindowStart.Before(statistic.WindowEnd) {
		return value.NewError("InvalidInput", "Statistics.Window")
	}

	if err := value.IdentifierSyntax(statistic.DatasetVersion).Validate(); err != nil {
		return err
	}

	return nil
}
