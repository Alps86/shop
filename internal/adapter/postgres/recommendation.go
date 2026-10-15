package postgres

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"slices"
)

func (repository RecommendationRepository) Candidates(ctx context.Context, query application.RecommendationQuery) ([]application.RecommendationCandidate, error) {
	source := catalog.Product{}
	if err := newRecord("product", string(query.ProductID), &source).read(ctx, repository.store.pool); err != nil {
		return nil, err
	}

	candidates := []application.RecommendationCandidate{}
	for _, link := range source.Accessories {
		if !repository.compatible(link, query.SKU) {
			continue
		}

		additions, err := repository.productCandidates(ctx, link.ProductID, query)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, additions...)
	}

	return candidates, nil
}

func (repository RecommendationRepository) Statistics(ctx context.Context, source value.ProductID) (application.PairStatistics, error) {
	statistics := application.PairStatistics{}
	if err := newRecord("statistics", "synthetic", &statistics).read(ctx, repository.store.pool); err != nil {
		return nil, err
	}

	result := application.PairStatistics{}
	for _, statistic := range statistics {
		if statistic.SourceProductID == source {
			result = append(result, statistic)
		}
	}

	return result, nil
}

func (repository RecommendationRepository) productCandidates(ctx context.Context, id value.ProductID, query application.RecommendationQuery) ([]application.RecommendationCandidate, error) {
	detail, err := repository.store.Catalog().Product(ctx, id, value.Optional[value.SKU]{}, query.Context)
	if err != nil {
		return nil, err
	}

	candidates := []application.RecommendationCandidate{}
	for _, card := range detail.Cards {
		candidates = append(candidates, application.RecommendationCandidate{Card: card, CategoryID: detail.Product.PrimaryCategoryID, Compatible: true})
	}

	return candidates, nil
}

func (repository RecommendationRepository) compatible(link catalog.AccessoryReference, sku value.SKU) bool {
	return len(link.CompatibleSKUs) == 0 || slices.Contains(link.CompatibleSKUs, sku)
}
