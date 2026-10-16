package opensearch

import (
	"bytes"
	"context"
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"io"
	"net/http"
	"strings"
)

func (adapter Adapter) send(ctx context.Context, method, path string, payload any) (Reply, error) {
	body, err := encodePayload(payload)
	if err != nil {
		return Reply{}, value.NewError("InvalidInput", "Projection")
	}

	return adapter.request(ctx, method, path, body)
}

func (adapter Adapter) request(ctx context.Context, method, path string, body []byte) (Reply, error) {
	request, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(adapter.config.Endpoint, "/")+path, bytes.NewReader(body))
	if err != nil {
		return Reply{}, value.NewError("DependencyUnavailable", "OpenSearch")
	}

	request.Header.Set("Content-Type", requestContentType(path))
	response, err := adapter.config.HTTP.Do(request)
	if err != nil {
		return Reply{}, value.NewError("SearchUnavailable", "OpenSearch")
	}

	return adapter.receive(response)
}

func (adapter Adapter) receive(response *http.Response) (Reply, error) {
	data, readErr := io.ReadAll(io.LimitReader(response.Body, 16<<20))
	closeErr := response.Body.Close()
	if errors.Join(readErr, closeErr) != nil {
		return Reply{}, value.NewError("SearchUnavailable", "Response")
	}

	return Reply{Status: response.StatusCode, Body: data}, nil
}

func requestContentType(path string) string {
	if strings.HasSuffix(path, "/_msearch") {
		return "application/x-ndjson"
	}

	return "application/json"
}
