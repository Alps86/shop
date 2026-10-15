package opensearch

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/url"
)

func (adapter Adapter) Initialize(ctx context.Context) error {
	reply, err := adapter.send(ctx, http.MethodGet, "/_alias/"+url.PathEscape(adapter.config.Index), nil)
	if err != nil {
		return err
	}

	if reply.Status == http.StatusOK {
		return nil
	}

	if reply.Status != http.StatusNotFound {
		return value.NewError("DependencyUnavailable", "Index")
	}

	index := adapter.config.Index + "-v1"
	if err := adapter.CreateIndex(ctx, index); err != nil {
		return err
	}

	return adapter.alias(ctx, index, false)
}

func (adapter Adapter) CreateIndex(ctx context.Context, index string) error {
	reply, err := adapter.send(ctx, http.MethodPut, "/"+url.PathEscape(index), json.RawMessage(mapping))
	if err != nil {
		return err
	}

	if reply.Status == http.StatusOK {
		return nil
	}

	if reply.indexExists() {
		return nil
	}

	return value.NewError("DependencyUnavailable", "Index")
}

func (reply Reply) indexExists() bool {
	var failure FailureResponse
	err := json.Unmarshal(reply.Body, &failure)

	return err == nil && failure.Error.Type == "resource_already_exists_exception"
}

func (adapter Adapter) alias(ctx context.Context, index string, replace bool) error {
	actions := []JSON{}
	if replace {
		actions = append(actions, JSON{"remove": JSON{"index": "*", "alias": adapter.config.Index}})
	}

	actions = append(actions, JSON{"add": JSON{"index": index, "alias": adapter.config.Index}})
	reply, err := adapter.send(ctx, http.MethodPost, "/_aliases", JSON{"actions": actions})
	if err != nil {
		return err
	}

	if reply.Status != http.StatusOK {
		return value.NewError("DependencyUnavailable", "Alias")
	}

	return nil
}
