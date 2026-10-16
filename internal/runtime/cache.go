package runtime

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
)

func (projection cacheProjection) Apply(ctx context.Context, event application.ProjectionEvent) error {
	if err := projection.Adapter.Apply(ctx, event); err != nil {
		return err
	}

	return projection.cache.Invalidate(ctx, projection.skus)
}

func catalogSKUs(snapshot application.AdminCatalogSnapshot) []value.SKU {
	skus := []value.SKU{}
	for _, product := range snapshot.Products {
		for _, variant := range product.Variants {
			skus = append(skus, variant.SKU)
		}
	}

	return skus
}
