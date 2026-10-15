package opensearch

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
)

func (response SearchResponse) result(query application.SearchQuery) (application.SearchResult, error) {
	result := application.SearchResult{TotalProducts: response.Hits.Total.Value, Page: query.Page, PageSize: 12}
	for _, hit := range response.Hits.Hits {
		card, err := hit.card()
		if err != nil {
			return application.SearchResult{}, err
		}

		result.Cards = append(result.Cards, card)
	}

	facets, err := response.facets(query)
	result.Facets = facets
	result.Suggestions = response.suggestions(query)

	return result, err
}

func (hit Hit) card() (application.ProductCard, error) {
	chosen := hit.Inner["chosen"].Hits.Hits
	if len(chosen) != 1 {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "ChosenVariant")
	}

	card := chosen[0].Source.Card
	if card.ProductID != hit.Source.ID {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "ProductID")
	}

	if card.SKU != chosen[0].Source.SKU {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "SKU")
	}

	return card, nil
}
