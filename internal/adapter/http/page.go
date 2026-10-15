package http

import (
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"slices"
	"strings"
)

func newPage(kind, title string) Page {
	return Page{Kind: kind, Title: title, Status: 200, Robots: "index,follow", CacheControl: "public,max-age=5", Context: value.RequestContext{Locale: "de-DE", Market: "DE", Currency: "EUR"}}
}

func (renderer Renderer) money(amount value.Money) string {
	return fmt.Sprintf("%d,%02d %s", amount.MinorUnits/100, amount.MinorUnits%100, amount.Currency)
}

func (renderer Renderer) localized(text catalog.LocalizedText) string {
	return text["de-DE"]
}

func (renderer Renderer) productURL(product catalog.Product) string {
	return "/de-DE/produkt/" + product.Slug + "--" + string(product.ID)
}

func (renderer Renderer) categoryURL(category catalog.Category) string {
	return "/de-DE/kategorie/" + category.Slug + "--" + string(category.ID)
}

func (renderer Renderer) attribute(attribute catalog.AttributeValue) string {
	if attribute.Text.Present {
		return renderer.localized(attribute.Text.Value)
	}

	if attribute.Number.Present {
		return string(attribute.Number.Value)
	}

	if attribute.Measurement.Present {
		return string(attribute.Measurement.Value.Value) + " " + attribute.Measurement.Value.Unit
	}

	if attribute.Boolean.Present {
		return fmt.Sprint(attribute.Boolean.Value)
	}

	return renderer.attributeOptions(attribute)
}

func (handler *Handler) validateLocale(request *http.Request) error {
	locale := strings.Split(request.URL.Path, "/")[1]
	if _, err := value.ParseLocale(locale); err != nil {
		return newHTTPError(400, "Ungültige Sprache")
	}

	if locale != "de-DE" {
		return newHTTPError(404, "Sprache nicht verfügbar")
	}

	return nil
}

func (renderer Renderer) basePrice(card application.ProductCard) (string, error) {
	if !card.Compliance.NetContent.Present {
		return "", nil
	}

	if !card.Compliance.UnitPriceBasis.Present {
		return "", nil
	}

	basis := card.Compliance.UnitPriceBasis.Value
	price, err := card.Compliance.NetContent.Value.UnitPrice(card.UnitPrice, basis)
	if err != nil {
		return "", err
	}

	return renderer.money(price) + " / " + string(basis.Value) + " " + basis.Unit, nil
}

func (renderer Renderer) attributeOptions(attribute catalog.AttributeValue) string {
	values := []string{}
	for _, option := range attribute.Options {
		values = append(values, string(option))
	}

	return strings.Join(values, ", ")
}

func (renderer Renderer) can(view application.CartView, action string) bool {
	return slices.Contains(view.Actions, action)
}
