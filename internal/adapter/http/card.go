package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"slices"
)

func (renderer Renderer) cardViews(cards []application.ProductCard) ([]cardView, error) {
	views := make([]cardView, len(cards))
	for index, card := range cards {
		view, err := renderer.cardView(card)
		if err != nil {
			return nil, err
		}

		views[index] = view
	}

	return views, nil
}

func (renderer Renderer) cardView(card application.ProductCard) (cardView, error) {
	basePrice, err := renderer.basePrice(card)
	if err != nil {
		return cardView{}, err
	}

	return cardView{ProductCard: card, ImageAlt: renderer.localized(card.Image.AltText), Price: renderer.money(card.UnitPrice), BasePrice: basePrice, Attributes: renderer.cardAttributes(card.TopAttributes)}, nil
}

func (renderer Renderer) cardAttributes(attributes map[value.AttributeCode]catalog.AttributeValue) []cardAttributeView {
	codes := make([]value.AttributeCode, 0, len(attributes))
	for code := range attributes {
		codes = append(codes, code)
	}

	slices.Sort(codes)
	views := make([]cardAttributeView, len(codes))
	for index, code := range codes {
		views[index] = cardAttributeView{Code: code, Text: renderer.attribute(attributes[code])}
	}

	return views
}
