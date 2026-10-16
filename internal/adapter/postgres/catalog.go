package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
)

func (repository CatalogRepository) Product(ctx context.Context, id value.ProductID, selected value.Optional[value.SKU], request value.RequestContext) (application.ProductDetail, error) {
	if err := request.Validate(); err != nil {
		return application.ProductDetail{}, err
	}

	product := catalog.Product{}
	if err := newRecord("product", string(id), &product).read(ctx, repository.store.pool); err != nil {
		return application.ProductDetail{}, err
	}

	cards, err := repository.cards(ctx, product, request)
	if err != nil {
		return application.ProductDetail{}, err
	}

	return repository.detail(product, cards, selected)
}

func (repository CatalogRepository) cards(ctx context.Context, product catalog.Product, request value.RequestContext) ([]application.ProductCard, error) {
	inventories, err := repository.store.productInventory(ctx, repository.store.pool, product)
	if err != nil {
		return nil, err
	}

	category := catalog.Category{}
	if err := newRecord("category", string(product.PrimaryCategoryID), &category).read(ctx, repository.store.pool); err != nil {
		return nil, err
	}

	return repository.variantCards(product, inventories, request, category.TopAttributes), nil
}

func (repository CatalogRepository) variantCards(product catalog.Product, inventories []sale.Inventory, request value.RequestContext, codes []value.AttributeCode) []application.ProductCard {
	cards := make([]application.ProductCard, 0, len(product.Variants))
	for index, variant := range product.Variants {
		card := repository.card(product, variant, inventories[index], request)
		card.TopAttributes = repository.topAttributes(product, variant, codes)

		cards = append(cards, card)
	}

	return cards
}

func (repository CatalogRepository) card(product catalog.Product, variant catalog.Variant, inventory sale.Inventory, request value.RequestContext) application.ProductCard {
	compliance := variant.ResolveRoles(product)

	quantity := inventory.AvailableQuantity()
	if compliance.Validate(variant.SKU, request, time.Now().UTC()) != nil {
		quantity = 0
	}

	name, _ := product.Name.Resolve(request.Locale)
	image := catalog.Image{}
	if len(product.Images) != 0 {
		image = product.Images[0]
	}

	if len(variant.Images) != 0 {
		image = variant.Images[0]
	}

	return application.ProductCard{ProductID: product.ID, URL: fmt.Sprintf("/%s/produkt/%s--%s?sku=%s", request.Locale, product.Slug, product.ID, variant.SKU), Name: name, SKU: variant.SKU, Image: image, UnitPrice: inventory.Price, AvailableQuantity: quantity, VariantCount: len(product.Variants), Compliance: compliance.Compliance}
}

func (repository CatalogRepository) detail(product catalog.Product, cards []application.ProductCard, selected value.Optional[value.SKU]) (application.ProductDetail, error) {
	if len(cards) == 0 {
		return application.ProductDetail{}, value.NewError("NotFound", "ProductID")
	}

	chosen := repository.preferred(cards).SKU
	if selected.Present {
		if _, err := product.Variant(selected.Value); err != nil {
			return application.ProductDetail{}, err
		}

		chosen = selected.Value
	}

	return application.ProductDetail{Product: product, ChosenSKU: chosen, Cards: cards}, nil
}

func (store *Store) productInventory(ctx context.Context, db database, product catalog.Product) ([]sale.Inventory, error) {
	result := make([]sale.Inventory, 0, len(product.Variants))
	for _, variant := range product.Variants {
		inventory := sale.Inventory{}
		if err := newRecord("inventory", string(variant.SKU), &inventory).read(ctx, db); err != nil {
			return nil, err
		}

		result = append(result, inventory)
	}

	return result, nil
}

func (repository CatalogRepository) Snapshot(ctx context.Context) (application.AdminCatalogSnapshot, error) {
	result := application.AdminCatalogSnapshot{}
	reads := map[string]any{"product": &result.Products, "family": &result.Families, "category": &result.Categories, "discount": &result.DiscountRules, "inventory": &result.Inventory}
	if err := repository.readSnapshot(ctx, reads); err != nil {
		return application.AdminCatalogSnapshot{}, err
	}

	return result, nil
}

func (repository CatalogRepository) Category(ctx context.Context, id value.CategoryID, request value.RequestContext) (application.CategoryView, error) {
	if err := request.Validate(); err != nil {
		return application.CategoryView{}, err
	}

	snapshot := application.AdminCatalogSnapshot{}
	reads := map[string]any{"family": &snapshot.Families, "category": &snapshot.Categories}
	if err := repository.readSnapshot(ctx, reads); err != nil {
		return application.CategoryView{}, err
	}

	return repository.category(snapshot, id)
}

func (repository CatalogRepository) category(snapshot application.AdminCatalogSnapshot, id value.CategoryID) (application.CategoryView, error) {
	categories := map[value.CategoryID]catalog.Category{}
	for _, category := range snapshot.Categories {
		categories[category.ID] = category
	}

	category, exists := categories[id]
	if !exists {
		return application.CategoryView{}, value.NewError("NotFound", "CategoryID")
	}

	result := application.CategoryView{Category: category, Breadcrumb: repository.breadcrumb(categories, id)}
	for _, child := range snapshot.Categories {
		if child.ParentID.Value == id {
			result.Children = append(result.Children, child)
		}
	}

	return repository.categoryDefinitions(result, snapshot.Families), nil
}

func (repository CatalogRepository) breadcrumb(categories map[value.CategoryID]catalog.Category, id value.CategoryID) []catalog.Category {
	result := []catalog.Category{}
	for id != "" {
		category := categories[id]
		result = append([]catalog.Category{category}, result...)
		id = category.ParentID.Value
	}

	return result
}

func (repository CatalogRepository) categoryDefinitions(view application.CategoryView, families []catalog.ProductFamily) application.CategoryView {
	definitions := map[value.AttributeCode]catalog.AttributeDefinition{}
	for _, family := range families {
		for _, definition := range family.Attributes {
			definitions[definition.Code] = definition
		}
	}

	for _, code := range view.Category.FilterAttributes {
		view.FilterDefinitions = append(view.FilterDefinitions, definitions[code])
	}

	for _, code := range view.Category.TopAttributes {
		view.TopDefinitions = append(view.TopDefinitions, definitions[code])
	}

	return view
}

func (repository CatalogRepository) ProductForSKU(ctx context.Context, sku value.SKU, request value.RequestContext) (application.ProductDetail, error) {
	product, err := repository.store.productForSKU(ctx, repository.store.pool, sku)
	if err != nil {
		return application.ProductDetail{}, err
	}

	return repository.Product(ctx, product.ID, value.Optional[value.SKU]{Value: sku, Present: true}, request)
}

func (repository CatalogRepository) preferred(cards []application.ProductCard) application.ProductCard {
	chosen := cards[0]
	for _, candidate := range cards[1:] {
		if repository.prefer(candidate, chosen) {
			chosen = candidate
		}
	}

	return chosen
}

func (repository CatalogRepository) prefer(candidate, current application.ProductCard) bool {
	available := candidate.AvailableQuantity > 0
	if available != (current.AvailableQuantity > 0) {
		return available
	}

	if candidate.UnitPrice.MinorUnits != current.UnitPrice.MinorUnits {
		return candidate.UnitPrice.MinorUnits < current.UnitPrice.MinorUnits
	}

	return candidate.SKU < current.SKU
}

func (repository CatalogRepository) topAttributes(product catalog.Product, variant catalog.Variant, codes []value.AttributeCode) catalog.Attributes {
	available := product.Attributes.Clone()
	for code, attribute := range variant.Attributes.Clone() {
		available[code] = attribute
	}

	chosen := catalog.Attributes{}
	for _, code := range codes {
		if attribute, exists := available[code]; exists {
			chosen[code] = attribute
		}
	}

	return chosen
}

func (repository CatalogRepository) readSnapshot(ctx context.Context, reads map[string]any) error {
	for kind, destination := range reads {
		if err := repository.store.documents(ctx, kind, destination); err != nil {
			return err
		}
	}

	return nil
}
