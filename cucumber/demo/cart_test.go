package demo_test

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) cartObserved(table *godog.Table) error {
	service := current.service()
	guest, err := service.Open(context.Background(), current.session())
	if err != nil {
		return err
	}

	view, err := service.GetCart(context.Background(), guest.Session, guest.CartID.Value)
	if err != nil {
		return err
	}

	return current.checkCart(view, guest, read(table)[0])
}

func (current *world) checkCart(view application.CartView, guest application.GuestView, data map[string]string) error {
	var quantity, discount int64
	for _, item := range view.Cart.Items {
		quantity += item.Quantity
	}

	for _, applied := range view.Cart.Evaluation.Value.Discounts {
		discount += applied.Amount.MinorUnits
	}

	observed := map[string]string{"Items": strconv.Itoa(len(view.Cart.Items)), "Quantity": strconv.FormatInt(quantity, 10), "DiscountMinor": strconv.FormatInt(discount, 10), "TotalMinor": strconv.FormatInt(view.Cart.Evaluation.Value.Total.MinorUnits, 10), "Profile": string(guest.Profile)}
	observed["NewCart"] = strconv.FormatBool(view.Cart.ID != current.order.Purchase.CartID)
	for key, expected := range data {
		if observed[key] != expected {
			return fmt.Errorf("Gastwarenkorb %s=%q statt %q", key, observed[key], expected)
		}
	}

	return nil
}
