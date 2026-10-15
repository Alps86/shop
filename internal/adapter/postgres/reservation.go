package postgres

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (attempt *purchaseAttempt) lockInventory(ctx context.Context) error {
	items := append(sale.CartItems{}, attempt.cart.Items...)
	slices.SortFunc(items, func(left, right sale.CartItem) int {
		return strings.Compare(string(left.SKU), string(right.SKU))
	})

	for _, item := range items {
		if err := attempt.lockItem(ctx, item); err != nil {
			return err
		}
	}

	return nil
}

func (attempt *purchaseAttempt) lockItem(ctx context.Context, item sale.CartItem) error {
	inventory := sale.Inventory{}
	if err := newRecord("inventory", string(item.SKU), &inventory).lock(ctx, attempt.tx); err != nil {
		return err
	}

	product, err := attempt.repository.store.productForSKU(ctx, attempt.tx, item.SKU)
	if err != nil {
		return err
	}

	return attempt.acceptItem(ctx, item, product, inventory)
}

func (attempt *purchaseAttempt) acceptItem(ctx context.Context, item sale.CartItem, product catalog.Product, inventory sale.Inventory) error {
	if _, err := product.ResolveCompliance(item.SKU, attempt.repository.store.config.Context, time.Now().UTC()); err != nil {
		return err
	}

	proposed, err := inventory.Reserve(attempt.orderID, item.Quantity, inventory.Revision)
	if value.Code(err) == "Unavailable" {
		return attempt.unavailable(ctx, product, item.SKU)
	}

	if err != nil {
		return err
	}

	attempt.inventory[item.SKU] = proposed
	attempt.products[item.SKU] = product

	return nil
}

func (attempt *purchaseAttempt) unavailable(ctx context.Context, product catalog.Product, sku value.SKU) error {
	cards, err := attempt.repository.store.Catalog().cards(ctx, product, attempt.repository.store.config.Context)
	if err != nil {
		return err
	}

	attempt.result = application.NewCartPurchaseResult(attempt.cart)
	attempt.result.UnavailableSKU = value.Optional[value.SKU]{Present: true, Value: sku}
	for _, card := range cards {
		if attempt.availableAlternative(card, sku) {
			attempt.result.Alternatives = append(attempt.result.Alternatives, card)
		}
	}

	return value.NewError("Unavailable", string(sku))
}

func (attempt *purchaseAttempt) availableAlternative(card application.ProductCard, sku value.SKU) bool {
	return card.SKU != sku && card.AvailableQuantity > 0
}

func (attempt *purchaseAttempt) reserve(ctx context.Context) error {
	ids := []value.ProductID{}
	for _, item := range attempt.cart.Items {
		if err := attempt.reserveItem(ctx, item); err != nil {
			return err
		}

		ids = append(ids, attempt.products[item.SKU].ID)
	}

	slices.Sort(ids)
	for _, id := range slices.Compact(ids) {
		if err := attempt.repository.store.project(ctx, attempt.tx, id); err != nil {
			return err
		}
	}

	return nil
}

func (attempt *purchaseAttempt) reserveItem(ctx context.Context, item sale.CartItem) error {
	return newRecord("inventory", string(item.SKU), attempt.inventory[item.SKU]).save(ctx, attempt.tx)
}
