package opensearch

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/url"
	"strconv"
)

func (adapter Adapter) Apply(ctx context.Context, event application.ProjectionEvent) error {
	document, err := adapter.document(event)
	if err != nil {
		return err
	}

	if adapter.projection == nil {
		return adapter.write(ctx, document, adapter.config.Index)
	}

	return adapter.projectLive(ctx, document)
}

func (adapter Adapter) projectLive(ctx context.Context, document Document) error {
	adapter.projection.mutex.Lock()
	defer adapter.projection.mutex.Unlock()
	if err := adapter.write(ctx, document, adapter.config.Index); err != nil {
		return err
	}

	if adapter.projection.index != "" {
		err := adapter.write(ctx, document, adapter.projection.index)
		adapter.projection.err = errors.Join(adapter.projection.err, err)

		return err
	}

	return nil
}

func (adapter Adapter) write(ctx context.Context, document Document, index string) error {
	path := "/" + url.PathEscape(index) + "/_doc/" + url.PathEscape(string(document.ID)) + "?version=" + strconv.FormatUint(document.Revision, 10) + "&version_type=external&refresh=true"
	reply, err := adapter.send(ctx, http.MethodPut, path, document)
	if err != nil {
		return err
	}

	if reply.projectionAccepted() {
		return nil
	}

	return value.NewError("DependencyUnavailable", "Projection")
}

func (reply Reply) projectionAccepted() bool {
	return reply.Status == http.StatusOK || reply.Status == http.StatusCreated || reply.Status == http.StatusConflict
}

func (adapter Adapter) document(event application.ProjectionEvent) (Document, error) {
	if err := value.IdentifierSyntax(string(event.ProductID)).Validate(); err != nil {
		return Document{}, err
	}

	if event.Revision == 0 {
		return Document{}, value.NewError("InvalidInput", "ProjectionVersion")
	}

	if err := event.Product.Validate(); err != nil {
		return Document{}, err
	}

	document := Document{ID: event.ProductID, Revision: event.Revision, Deleted: !event.Product.Present}
	if !event.Product.Present {
		return document, nil
	}

	return document.project(event, adapter.config)
}

func (document Document) project(event application.ProjectionEvent, config Config) (Document, error) {
	product := event.Product.Value
	if product.ID != document.ID {
		return Document{}, value.NewError("InvalidInput", "ProductID")
	}

	document = document.metadata(product, config)
	inventory := map[value.SKU]sale.Inventory{}
	for _, item := range event.Inventory {
		inventory[item.SKU] = item
	}

	return document.withVariants(product, inventory, config)
}

func (document Document) metadata(product catalog.Product, config Config) Document {
	document.Name, _ = product.Name.Resolve(config.Context.Locale)
	document.Description, _ = product.Description.Resolve(config.Context.Locale)
	document.Brand = product.Brand
	document.Published = product.Published
	document.Visibility = "private"
	if product.Published {
		document.Visibility = "public"
	}

	document.CategoryIDs = config.categoryClosure(append([]value.CategoryID{product.PrimaryCategoryID}, product.AdditionalCategoryIDs...))
	document.Suggest = config.suggestionNames(document.Name, document.CategoryIDs)

	return document
}

func (config Config) categoryClosure(pending []value.CategoryID) []value.CategoryID {
	result := []value.CategoryID{}
	seen := map[value.CategoryID]bool{}
	for len(pending) > 0 {
		id := pending[0]
		pending = pending[1:]
		if seen[id] {
			continue
		}

		seen[id] = true
		result = append(result, id)
		if parent := config.Categories[id].ParentID; parent.Present {
			pending = append(pending, parent.Value)
		}
	}

	return result
}

func (config Config) suggestionNames(product string, categories []value.CategoryID) []string {
	result := []string{product}
	for _, id := range categories {
		name, _ := config.Categories[id].Name.Resolve(config.Context.Locale)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}

func (document Document) withVariants(product catalog.Product, inventory map[value.SKU]sale.Inventory, config Config) (Document, error) {
	for _, variant := range product.Variants {
		projected, err := NewVariant(product, variant, inventory[variant.SKU], config)
		if err != nil {
			return Document{}, err
		}

		document.Variants = append(document.Variants, projected)
	}

	return document, nil
}
