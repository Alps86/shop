package application

import (
	"time"

	"github.com/Alps86/shop/internal/domain/value"
)

func (statistics PairStatistics) Validate(products []value.ProductID) error {
	seen := map[string]bool{}
	for _, entry := range statistics {
		if err := entry.validate(products); err != nil {
			return err
		}

		if !entry.sameWindow(statistics[0]) {
			return value.NewError("InvalidInput", "Window")
		}

		if seen[entry.key()] {
			return value.NewError("InvalidInput", "Pair")
		}

		seen[entry.key()] = true
	}

	return nil
}

func (entry PairStatistic) validate(products []value.ProductID) error {
	if entry.invalidValues() {
		return value.NewError("InvalidInput", "Statistic")
	}

	for _, id := range []value.ProductID{entry.SourceProductID, entry.TargetProductID} {
		if !entry.known(products, id) {
			return value.NewError("InvalidReference", "ProductID")
		}
	}

	return nil
}

func (entry PairStatistic) invalidValues() bool {
	return entry.BasketCount < 0 || entry.SourceProductID == entry.TargetProductID || entry.DatasetVersion == "" || entry.WindowEnd.Sub(entry.WindowStart) != 90*24*time.Hour
}

func (entry PairStatistic) sameWindow(other PairStatistic) bool {
	return entry.WindowStart.Equal(other.WindowStart) && entry.WindowEnd.Equal(other.WindowEnd) && entry.DatasetVersion == other.DatasetVersion
}

func (entry PairStatistic) known(products []value.ProductID, id value.ProductID) bool {
	for _, product := range products {
		if product == id {
			return true
		}
	}

	return false
}

func (entry PairStatistic) key() string {
	ids := []string{string(entry.SourceProductID), string(entry.TargetProductID)}
	if ids[0] > ids[1] {
		ids[0], ids[1] = ids[1], ids[0]
	}

	return ids[0] + ":" + ids[1]
}
