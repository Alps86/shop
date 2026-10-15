package opensearch

import (
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"math/big"
)

func (ranking Ranking) query(base JSON) JSON {
	if len(ranking.Interests) == 0 {
		return base
	}

	// A matching zero score avoids OpenSearch's implicit score of one for unmatched functions.
	functions := []JSON{{"script_score": JSON{"script": JSON{"source": "0.0"}}}}
	for _, interest := range ranking.Interests {
		weight := ranking.weight(interest.Weight)
		if weight == 0 {
			continue
		}

		functions = append(functions, JSON{"filter": JSON{"term": JSON{"category_ids": interest.CategoryID}}, "weight": weight})
	}

	if len(functions) == 1 {
		return base
	}

	return JSON{"function_score": JSON{"query": base, "functions": functions, "boost_mode": "sum", "score_mode": "sum"}}
}

func (ranking Ranking) weight(weight value.Decimal) float64 {
	number, err := weight.Scaled()
	if err != nil {
		return 0
	}

	return float64(max(0, min(1000000, number))) / 10000000
}

func (config Config) margin(inventory sale.Inventory) int64 {
	cost, exists := config.PurchaseCosts[inventory.SKU]
	if !exists {
		return 0
	}

	if !config.validCost(cost, inventory.Price) {
		return 0
	}

	difference := big.NewInt(inventory.Price.MinorUnits - cost.MinorUnits)
	difference.Mul(difference, big.NewInt(1000000))
	difference.Quo(difference, big.NewInt(inventory.Price.MinorUnits))

	return difference.Int64()
}

func (config Config) validCost(cost, price value.Money) bool {
	return cost.Currency == price.Currency && cost.MinorUnits >= 0 && cost.MinorUnits < price.MinorUnits
}
