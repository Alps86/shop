package value

import (
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"strings"
)

func ParseLocale(input string) (Locale, error) {
	tag, err := language.Parse(input)
	if err != nil {
		return "", NewError("InvalidLocale", "Locale")
	}

	if tag == language.Und {
		return "", NewError("InvalidLocale", "Locale")
	}

	return Locale(tag.String()), nil
}

func ParseCountryCode(input string) (CountryCode, error) {
	if len(input) != 2 {
		return "", NewError("InvalidInput", "CountryCode")
	}

	region, err := language.ParseRegion(strings.ToUpper(input))
	if err != nil {
		return "", NewError("InvalidInput", "CountryCode")
	}

	code := CountryCode(region.String())
	if !code.assigned(region) {
		return "", NewError("InvalidInput", "CountryCode")
	}

	return code, nil
}

func ParseCurrency(input string) (Currency, error) {
	unit, err := currency.ParseISO(strings.ToUpper(input))
	if err != nil {
		return "", NewError("InvalidMoney", "Currency")
	}

	return Currency(unit.String()), nil
}

func (context RequestContext) Validate() error {
	if context.Locale != "de-DE" {
		return NewError("UnsupportedContext", "Locale")
	}

	if context.Currency != "EUR" {
		return NewError("UnsupportedContext", "Currency")
	}

	if context.Market != "DE" {
		return NewError("UnsupportedContext", "Market")
	}

	return nil
}

func (code CountryCode) assigned(region language.Region) bool {
	if region.IsPrivateUse() {
		return false
	}

	return region.IsCountry()
}
