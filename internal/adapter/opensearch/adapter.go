package opensearch

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"time"
)

func New(config Config) Adapter {
	if config.HTTP == nil {
		config.HTTP = &http.Client{Timeout: 5 * time.Second}
	}

	if config.Clock == nil {
		config.Clock = Clock{}
	}

	return Adapter{config: config, projection: &ProjectionState{}}
}

func (adapter Adapter) Search(ctx context.Context, query application.SearchQuery) (application.SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()

	body, err := adapter.searchBody(query)
	if err != nil {
		return application.SearchResult{}, err
	}

	if err := query.Validate(adapter.config.Schema); err != nil {
		return application.SearchResult{}, err
	}

	return adapter.performSearch(ctx, query, body)
}

func (adapter Adapter) performSearch(ctx context.Context, query application.SearchQuery, body JSON) (application.SearchResult, error) {
	reply, err := adapter.send(ctx, http.MethodPost, "/"+adapter.config.Index+"/_search?request_cache=true", body)
	if err != nil {
		return application.SearchResult{}, err
	}

	return adapter.searchResult(reply, query)
}

func (adapter Adapter) searchResult(reply Reply, query application.SearchQuery) (application.SearchResult, error) {
	if reply.Status != http.StatusOK {
		return application.SearchResult{}, value.NewError("SearchUnavailable", "Search")
	}

	var response SearchResponse
	if err := json.Unmarshal(reply.Body, &response); err != nil {
		return application.SearchResult{}, value.NewError("SearchUnavailable", "Response")
	}

	return response.result(query)
}

func (clock Clock) Now() time.Time {
	return time.Now()
}
