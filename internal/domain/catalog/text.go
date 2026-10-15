package catalog

import (
	"github.com/Alps86/shop/internal/domain/value"
	"maps"
	"strings"
)

func (text LocalizedText) Resolve(locale value.Locale) (string, value.Locale) {
	candidates := []value.Locale{locale, value.Locale(strings.Split(string(locale), "-")[0]), "de-DE"}
	for _, candidate := range candidates {
		if content := text[candidate]; strings.TrimSpace(content) != "" {
			return content, candidate
		}
	}

	return "", ""
}

func (text LocalizedText) Valid() bool {
	for locale, content := range text {
		if _, err := value.ParseLocale(string(locale)); err != nil {
			return false
		}

		if strings.TrimSpace(content) == "" {
			return false
		}
	}

	return len(text) > 0
}

func (attributes Attributes) Clone() Attributes {
	result := make(Attributes, len(attributes))
	for code, attribute := range attributes {
		attribute.Options = append([]OptionCode(nil), attribute.Options...)
		attribute.Text.Value = maps.Clone(attribute.Text.Value)
		result[code] = attribute
	}

	return result
}
