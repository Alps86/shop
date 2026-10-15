package postgres

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (repository EvaluationRepository) Evaluate(ctx context.Context, cart sale.Cart) (sale.CartEvaluation, error) {
	return repository.evaluate(ctx, repository.store.pool, cart)
}

func (repository EvaluationRepository) evaluate(ctx context.Context, db database, cart sale.Cart) (sale.CartEvaluation, error) {
	items, err := repository.items(ctx, db, cart)
	if err != nil {
		return sale.CartEvaluation{}, err
	}

	rules := sale.DiscountRules{}
	if err := newRecord("ruleset", cart.RuleSetVersion, &rules).read(ctx, db); err != nil {
		return sale.CartEvaluation{}, err
	}

	evaluation, err := items.Evaluate(rules, repository.shipping(cart))
	evaluation.ID = value.EvaluationID("e-" + rand.Text())
	evaluation.CartRevision = cart.Revision
	evaluation.RuleSetVersion = cart.RuleSetVersion
	evaluation.EvaluatedAt = time.Now().UTC()

	return evaluation, err
}

func (repository EvaluationRepository) shipping(cart sale.Cart) sale.ShippingMethod {
	if !cart.Checkout.Present {
		return sale.Standard
	}

	return cart.Checkout.Value.ShippingMethod
}

func (repository EvaluationRepository) items(ctx context.Context, db database, cart sale.Cart) (sale.PricedItems, error) {
	items := sale.PricedItems{}
	for _, item := range cart.Items {
		priced, err := repository.item(ctx, db, item)
		if err != nil {
			return nil, err
		}

		items = append(items, priced)
	}

	return items, nil
}

func (repository EvaluationRepository) item(ctx context.Context, db database, item sale.CartItem) (sale.PricedItem, error) {
	product, err := repository.store.productForSKU(ctx, db, item.SKU)
	if err != nil {
		return sale.PricedItem{}, err
	}

	if _, err := product.ResolveCompliance(item.SKU, repository.store.config.Context, time.Now().UTC()); err != nil {
		return sale.PricedItem{}, err
	}

	inventory := sale.Inventory{}
	if err := newRecord("inventory", string(item.SKU), &inventory).read(ctx, db); err != nil {
		return sale.PricedItem{}, err
	}

	return sale.PricedItem{SKU: item.SKU, ProductID: product.ID, Quantity: item.Quantity, UnitPrice: inventory.Price, InventoryRevision: inventory.Revision}, nil
}

func (store *Store) productForSKU(ctx context.Context, db database, sku value.SKU) (catalog.Product, error) {
	var id value.ProductID
	if err := db.QueryRow(ctx, "SELECT product_id FROM sku_registry WHERE sku=$1", sku).Scan(&id); err != nil {
		return catalog.Product{}, databaseError(err)
	}

	product := catalog.Product{}
	err := newRecord("product", string(id), &product).read(ctx, db)

	return product, err
}
