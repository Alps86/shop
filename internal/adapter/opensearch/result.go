package opensearch

import (
	"encoding/json"
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

	id, err := hit.Fields.productID()
	if err != nil {
		return application.ProductCard{}, err
	}

	return chosen[0].Fields.card(id)
}

func (fields StoredProductFields) productID() (value.ProductID, error) {
	if len(fields.ProductID) != 1 {
		return "", value.NewError("SearchUnavailable", "ProductID")
	}

	id := fields.ProductID[0]
	if err := value.IdentifierSyntax(string(id)).Validate(); err != nil {
		return "", value.NewError("SearchUnavailable", "ProductID")
	}

	return id, nil
}

func (fields StoredVariantFields) validate() error {
	if len(fields.SKU) != 1 {
		return value.NewError("SearchUnavailable", "SKU")
	}

	if err := value.IdentifierSyntax(string(fields.SKU[0])).Validate(); err != nil {
		return value.NewError("SearchUnavailable", "SKU")
	}

	if len(fields.CardPayload) != 1 {
		return value.NewError("SearchUnavailable", "CardPayload")
	}

	return nil
}

func (fields StoredVariantFields) card(id value.ProductID) (application.ProductCard, error) {
	if err := fields.validate(); err != nil {
		return application.ProductCard{}, err
	}

	var card application.ProductCard
	if err := json.Unmarshal(fields.CardPayload[0], &card); err != nil {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "CardPayload")
	}

	if card.ProductID != id {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "ProductID")
	}

	if card.SKU != fields.SKU[0] {
		return application.ProductCard{}, value.NewError("SearchUnavailable", "SKU")
	}

	return card, nil
}

func (response SearchResponse) searchResult(query application.SearchQuery, multi bool) (application.SearchResult, error) {
	if !multi {
		return response.result(query)
	}

	result, err := response.result(query)
	if err != nil {
		return application.SearchResult{}, err
	}

	return result, nil
}
