package http

import (
	"context"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"time"
)

func NewCacheInvalidator(endpoint, key string) (CacheInvalidator, error) {
	request, err := http.NewRequest("PURGE", endpoint, nil)
	if err != nil {
		return CacheInvalidator{}, fmt.Errorf("cacheendpoint: %w", err)
	}

	if request.URL.Scheme != "http" {
		return CacheInvalidator{}, fmt.Errorf("cacheendpoint benötigt internes HTTP")
	}

	if request.URL.Host == "" {
		return CacheInvalidator{}, fmt.Errorf("cacheendpoint benötigt internen Host")
	}

	if len(key) < 32 {
		return CacheInvalidator{}, fmt.Errorf("cacheschlüssel benötigt mindestens 32 Bytes")
	}

	return CacheInvalidator{request: request, key: key, client: &http.Client{Timeout: 150 * time.Millisecond}}, nil
}

func (cache CacheInvalidator) Invalidate(ctx context.Context, skus []value.SKU) error {
	if err := validateCacheSKUs(skus); err != nil {
		return err
	}

	if err := cache.purge(ctx, "X-Cache-Public-Search", "1"); err != nil {
		return err
	}

	for _, sku := range skus {
		if err := cache.purge(ctx, "X-Cache-SKU", string(sku)); err != nil {
			return err
		}
	}

	return nil
}

func validateCacheSKUs(skus []value.SKU) error {
	for _, sku := range skus {
		if err := value.IdentifierSyntax(sku).Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (cache CacheInvalidator) purge(ctx context.Context, header, key string) error {
	request := cache.request.Clone(ctx)

	request.Header.Set("X-Cache-Key", cache.key)
	request.Header.Set(header, key)
	response, err := cache.client.Do(request)
	if err != nil {
		return fmt.Errorf("cacheinvalidierung: %w", err)
	}

	closeError := response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("cacheinvalidierung: HTTP %d", response.StatusCode)
	}

	return closeError
}
