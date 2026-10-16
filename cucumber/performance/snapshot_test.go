package performance_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/domain/value"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"time"
)

func (current *world) captureSnapshot(directory string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	proof := failureSnapshot{StartedUTC: time.Now().UTC(), Note: "After original failure; sequential non-atomic reads, not the original failure state"}
	endpoint := os.Getenv("SHOP_PERFORMANCE_OPENSEARCH_URL")
	index := os.Getenv("SHOP_PERFORMANCE_SEARCH_INDEX")
	proof.Alias = readSnapshotHTTP(snapshotRequest{ctx: ctx, url: endpoint + "/_alias/" + url.PathEscape(index)})
	proof.Document = readSnapshotDocument(ctx, endpoint, proof.Alias)
	proof.Product, proof.PostgreSQLError = current.readSnapshotProduct(ctx)
	proof.Pool = current.readSnapshotPool(ctx)
	proof.EndedUTC = time.Now().UTC()

	return writeDiagnosticJSON(filepath.Join(directory, "after-failure.json"), proof)
}

func readSnapshotHTTP(input snapshotRequest) snapshotHTTP {
	request, err := http.NewRequestWithContext(input.ctx, http.MethodGet, input.url, nil)
	if err != nil {
		return snapshotHTTP{Error: snapshotError(err)}
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return snapshotHTTP{Error: snapshotError(err)}
	}

	return receiveSnapshot(response)
}

func receiveSnapshot(response *http.Response) snapshotHTTP {
	payload, readError := io.ReadAll(io.LimitReader(response.Body, 1024*1024))
	if err := errors.Join(readError, response.Body.Close()); err != nil {
		return snapshotHTTP{Status: response.StatusCode, Error: snapshotError(err)}
	}

	if !json.Valid(payload) {
		return snapshotHTTP{Status: response.StatusCode, Error: "Response is not complete valid JSON"}
	}

	return snapshotHTTP{Status: response.StatusCode, Body: payload}
}

func readSnapshotDocument(ctx context.Context, endpoint string, alias snapshotHTTP) snapshotHTTP {
	var indices map[string]json.RawMessage
	if err := json.Unmarshal(alias.Body, &indices); err != nil {
		return snapshotHTTP{Error: snapshotError(err)}
	}

	if len(indices) != 1 {
		return snapshotHTTP{Error: "Alias does not resolve to exactly one index"}
	}

	index := slices.Sorted(maps.Keys(indices))[0]
	fields := "product_id,revision,deleted,published,name,brand,category_ids,visibility,variants.sku,variants.price,variants.available,variants.card.Compliance.Context,variants.card.Compliance.ApplicableRules"
	target := endpoint + "/" + url.PathEscape(index) + "/_doc/bench-product-01?_source_includes=" + url.QueryEscape(fields)

	return readSnapshotHTTP(snapshotRequest{ctx: ctx, url: target})
}

func snapshotError(err error) string {
	if err == nil {
		return ""
	}

	return fmt.Sprintf("%T: %s", err, value.Code(err))
}
