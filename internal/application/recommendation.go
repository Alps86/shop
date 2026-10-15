package application

import (
	"context"
	"sort"

	"github.com/Alps86/shop/internal/domain/value"
)

func (service RecommendationService) Recommend(ctx context.Context, query RecommendationQuery) (RecommendationResult, error) {
	candidates, degraded := service.candidates(ctx, query)
	if degraded {
		return RecommendationResult{Degraded: true}, nil
	}

	statistics, err := service.Source.Statistics(ctx, query.ProductID)
	if err != nil {
		statistics = nil
	}

	ranked, rankErr := query.rank(candidates, statistics)
	if rankErr != nil {
		return RecommendationResult{}, rankErr
	}

	return ranked.result(query.Limit, len(statistics) > 0, err != nil), nil
}

func (ranked rankedRecommendations) result(limit int, statistics, degraded bool) RecommendationResult {
	label := "Passendes Zubehör"
	if statistics {
		label = "Häufig zusammen gekauft"
	}

	if len(ranked) == 0 {
		label = ""
	}

	return RecommendationResult{Cards: ranked.cards(limit), Label: label, Degraded: degraded}
}

func (query RecommendationQuery) rank(candidates []RecommendationCandidate, statistics PairStatistics) (rankedRecommendations, error) {
	result := rankedRecommendations{}
	for _, candidate := range candidates {
		score, err := query.score(candidate, statistics)
		if err != nil {
			return nil, err
		}

		result = result.include(candidate, score)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].before(result[j])
	})

	return result, nil
}

func (query RecommendationQuery) score(candidate RecommendationCandidate, statistics PairStatistics) (int64, error) {
	count := statistics.count(query.ProductID, candidate.Card.ProductID)
	weight, err := query.weightFor(candidate.CategoryID)
	if err != nil {
		return 0, err
	}

	return count * (10000000 + weight), nil
}

func (query RecommendationQuery) weightFor(category value.CategoryID) (int64, error) {
	for _, interest := range query.Interests {
		if interest.CategoryID == category {
			return query.weight(interest.Weight)
		}
	}

	return 0, nil
}

func (query RecommendationQuery) weight(weight value.Decimal) (int64, error) {
	scaled, err := weight.Scaled()
	if err != nil {
		return 0, err
	}

	if scaled < 0 {
		return 0, value.NewError("InvalidInput", "Weight")
	}

	if scaled > 1000000 {
		return 0, value.NewError("InvalidInput", "Weight")
	}

	return scaled, nil
}

func (statistics PairStatistics) count(source, target value.ProductID) int64 {
	for _, entry := range statistics {
		if entry.matches(source, target) {
			return entry.BasketCount
		}
	}

	return 0
}

func (entry PairStatistic) matches(source, target value.ProductID) bool {
	forward := entry.SourceProductID == source && entry.TargetProductID == target
	reverse := entry.SourceProductID == target && entry.TargetProductID == source

	return forward || reverse
}

func (ranked rankedRecommendations) include(candidate RecommendationCandidate, score int64) rankedRecommendations {
	if !candidate.Compatible {
		return ranked
	}

	if candidate.Card.AvailableQuantity <= 0 {
		return ranked
	}

	for index, existing := range ranked {
		if existing.Candidate.Card.ProductID == candidate.Card.ProductID {
			return ranked.replace(index, candidate)
		}
	}

	return append(ranked, rankedRecommendation{Candidate: candidate, Score: score})
}

func (ranked rankedRecommendations) replace(index int, candidate RecommendationCandidate) rankedRecommendations {
	current := ranked[index].Candidate.Card
	if candidate.Card.UnitPrice.MinorUnits < current.UnitPrice.MinorUnits {
		ranked[index].Candidate = candidate
		return ranked
	}

	if candidate.Card.UnitPrice.MinorUnits != current.UnitPrice.MinorUnits {
		return ranked
	}

	if candidate.Card.SKU < current.SKU {
		ranked[index].Candidate = candidate
	}

	return ranked
}

func (entry rankedRecommendation) before(other rankedRecommendation) bool {
	if entry.Score != other.Score {
		return entry.Score > other.Score
	}

	return entry.Candidate.Card.ProductID < other.Candidate.Card.ProductID
}

func (ranked rankedRecommendations) cards(limit int) []ProductCard {
	limit = min(max(limit, 0), 3, len(ranked))
	cards := make([]ProductCard, 0, limit)
	for _, entry := range ranked[:limit] {
		cards = append(cards, entry.Candidate.Card)
	}

	return cards
}

func (service RecommendationService) candidates(ctx context.Context, query RecommendationQuery) ([]RecommendationCandidate, bool) {
	candidates, err := service.Source.Candidates(ctx, query)
	if err != nil {
		return nil, true
	}

	return candidates, false
}
