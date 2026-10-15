package sale

import (
	"github.com/Alps86/shop/internal/domain/value"
	"regexp"
	"strings"
)

const (
	Standard   ShippingMethod = "standard"
	Express    ShippingMethod = "express"
	Invoice    PaymentMethod  = "invoice"
	CreditCard PaymentMethod  = "creditCard"
	Pair       DiscountScope  = "pair"
	PerOrder   DiscountScope  = "order"
)

func (address Address) Validate() error {
	for field, content := range map[string]string{"Recipient": address.Recipient, "Street": address.Street, "City": address.City} {
		if strings.TrimSpace(content) == "" {
			return value.NewError("InvalidInput", field)
		}
	}

	if !regexp.MustCompile(`^[0-9]{5}$`).MatchString(address.PostalCode) {
		return value.NewError("InvalidInput", "PostalCode")
	}

	if address.Country != "DE" {
		return value.NewError("UnsupportedContext", "Country")
	}

	return nil
}

func (method ShippingMethod) Cost() (value.Money, error) {
	if method == Standard {
		return value.Money{MinorUnits: 490, Currency: "EUR"}, nil
	}

	if method == Express {
		return value.Money{MinorUnits: 1490, Currency: "EUR"}, nil
	}

	return value.Money{}, value.NewError("InvalidInput", "ShippingMethod")
}

func (checkout Checkout) Validate() error {
	if err := checkout.ShippingAddress.Validate(); err != nil {
		return err
	}

	if _, err := checkout.ShippingMethod.Cost(); err != nil {
		return err
	}

	if checkout.PaymentMethod == Invoice {
		return nil
	}

	if checkout.PaymentMethod == CreditCard {
		return nil
	}

	return value.NewError("InvalidInput", "PaymentMethod")
}
