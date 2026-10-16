package opensearch

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"time"
)

func New(config Config) Adapter {
	if config.HTTP == nil {
		config.HTTP = newHTTPClient()
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
	request := adapter.searchRequest(body)
	reply, err := adapter.send(ctx, http.MethodPost, request.Path, request.Payload)
	if err != nil {
		return application.SearchResult{}, err
	}

	return adapter.searchResult(reply, query, request.Multi)
}

func (adapter Adapter) searchResult(reply Reply, query application.SearchQuery, multi bool) (application.SearchResult, error) {
	if reply.Status != http.StatusOK {
		return application.SearchResult{}, value.NewError("SearchUnavailable", "Search")
	}

	response, err := decodeSearchResponse(reply.Body, multi)
	if err != nil {
		return application.SearchResult{}, value.NewError("SearchUnavailable", "Response")
	}

	return response.searchResult(query, multi)
}

func (clock Clock) Now() time.Time {
	return time.Now()
}
