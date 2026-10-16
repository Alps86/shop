package http_test

import (
	"context"
	"errors"
	"fmt"
	shophttp "github.com/Alps86/shop/internal/adapter/http"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func (current *world) prepareCache(table *godog.Table) error {
	fields := read(table)[0]
	endpoint := fields["Endpoint"]
	if endpoint == "test-server" {
		current.cacheStatus = readCount(fields["Status"])
		current.cacheGlobalStatus = readCount(fields["GlobalStatus"])
		current.cacheServer = httptest.NewServer(cacheOrigin{current})
		endpoint = current.cacheServer.URL
	}

	current.cache, current.err = shophttp.NewCacheInvalidator(endpoint, fields["Key"])

	return nil
}

func (current *world) invalidateCache(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if read(table)[0]["Context"] == "cancelled" {
		cancel()
	}

	current.err = current.cache.Invalidate(ctx, skuList(read(table)[0]["SKUs"]))

	return nil
}

func (origin cacheOrigin) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	origin.current.mutex.Lock()
	defer origin.current.mutex.Unlock()
	origin.current.cacheRequests = append(origin.current.cacheRequests, cacheRequest{PublicSearch: request.Header.Get("X-Cache-Public-Search"), Method: request.Method, SKU: request.Header.Get("X-Cache-SKU"), Key: request.Header.Get("X-Cache-Key")})
	writer.WriteHeader(origin.responseStatus(request))
}

func (current *world) cacheResult(table *godog.Table) error {
	skus := []string{}
	methods := []string{}
	keys := []string{}
	globals := []string{}
	for _, request := range current.cacheRequests {
		skus = append(skus, request.SKU)
		methods = append(methods, request.Method)
		keys = append(keys, request.Key)
		globals = append(globals, request.PublicSearch)
	}

	actual := map[string]string{"PublicSearch": strings.Join(globals, ","), "Outcome": current.cacheOutcome(), "Calls": strconv.Itoa(len(current.cacheRequests)), "SKUs": strings.Join(skus, ","), "Methods": strings.Join(methods, ","), "Keys": strings.Join(keys, ",")}

	return read(table).matches(actual)
}

func skuList(input string) []value.SKU {
	result := []value.SKU{}
	if input == "" {
		return result
	}

	for _, sku := range strings.Split(input, ",") {
		result = append(result, value.SKU(sku))
	}

	return result
}

func (current *world) stopCache(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	if current.cacheServer != nil {
		current.cacheServer.Close()
	}

	return ctx, nil
}

func (current *world) invalidateComposition(table *godog.Table) error {
	fields := read(table)[0]
	cache, err := shophttp.NewCacheInvalidator(current.engine.url, fields["Key"])
	if err != nil {
		return err
	}

	if err := cache.Invalidate(context.Background(), skuList(fields["SKUs"])); err != nil {
		return fmt.Errorf("Varnishinvalidierung: %w", err)
	}

	return nil
}

func (current *world) cacheOutcome() string {
	if errors.Is(current.err, context.Canceled) {
		return context.Canceled.Error()
	}

	if current.err != nil {
		return current.err.Error()
	}

	return "success"
}

func (origin cacheOrigin) responseStatus(request *http.Request) int {
	if request.Header.Get("X-Cache-Public-Search") == "1" {
		if origin.current.cacheGlobalStatus != 0 {
			return origin.current.cacheGlobalStatus
		}
	}

	return origin.current.cacheStatus
}
