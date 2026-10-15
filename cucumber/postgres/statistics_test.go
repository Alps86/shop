package postgres_test

import (
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) statisticsMutation(input *godog.Table) error {
	row := table(input)
	mutations := map[string]func(string) error{"Source": current.statisticSource, "Target": current.statisticTarget, "Count": current.statisticCount, "Version": current.statisticVersion, "Window": current.statisticWindow, "MainRule": current.ruleSource, "AccessoryRule": current.ruleTarget}

	return mutations[row["Field"]](row["Value"])
}

func (current *world) statisticSource(raw string) error {
	current.manifest.CoPurchaseStatistics[0].SourceProductID = value.ProductID(raw)
	return nil
}

func (current *world) statisticTarget(raw string) error {
	current.manifest.CoPurchaseStatistics[0].TargetProductID = value.ProductID(raw)
	return nil
}

func (current *world) statisticCount(raw string) error {
	count, err := strconv.ParseInt(raw, 10, 64)
	current.manifest.CoPurchaseStatistics[0].BasketCount = count

	return err
}

func (current *world) statisticVersion(raw string) error {
	current.manifest.CoPurchaseStatistics[0].DatasetVersion = raw
	return nil
}

func (current *world) statisticWindow(_ string) error {
	current.manifest.CoPurchaseStatistics[0].WindowStart = current.manifest.CoPurchaseStatistics[0].WindowEnd
	return nil
}

func (current *world) ruleSource(raw string) error {
	current.manifest.DiscountRules[0].MainProductID = value.ProductID(raw)
	return nil
}

func (current *world) ruleTarget(raw string) error {
	current.manifest.DiscountRules[0].AccessoryID = value.ProductID(raw)
	return nil
}
