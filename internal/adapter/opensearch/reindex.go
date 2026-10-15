package opensearch

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (adapter Adapter) Rebuild(ctx context.Context, index string, at time.Time) error {
	if !strings.HasPrefix(index, adapter.config.Index+"-") {
		return value.NewError("InvalidInput", "Index")
	}

	if adapter.config.Source == nil {
		return value.NewError("DependencyUnavailable", "ReindexSource")
	}

	if err := adapter.beginRebuild(ctx, index); err != nil {
		return err
	}

	defer adapter.endRebuild()

	return adapter.rebuildSnapshot(ctx, index, at)
}

func (adapter Adapter) freshIndex(ctx context.Context, index string) error {
	reply, err := adapter.send(ctx, http.MethodPut, "/"+url.PathEscape(index), json.RawMessage(mapping))
	if err != nil {
		return err
	}

	if reply.Status != http.StatusOK {
		return value.NewError("Conflict", "Index")
	}

	return nil
}

func (adapter Adapter) replay(ctx context.Context, events []application.ProjectionEvent) error {
	for _, event := range events {
		if err := adapter.Apply(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (adapter Adapter) rebuildSnapshot(ctx context.Context, index string, at time.Time) (err error) {
	snapshot, err := adapter.config.Source.Snapshot(ctx, at)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, adapter.config.Source.Abort(context.WithoutCancel(ctx), snapshot.Cursor))
	}()

	return adapter.rebuildEvents(ctx, index, snapshot)
}

func (adapter Adapter) rebuildEvents(ctx context.Context, index string, snapshot event.ReindexSnapshot) error {
	projection := adapter
	projection.config.Index = index
	projection.projection = nil
	if err := projection.replay(ctx, snapshot.Events); err != nil {
		return err
	}

	if err := adapter.config.Source.Catchup(ctx, snapshot.Cursor, projection); err != nil {
		return err
	}

	return adapter.finishRebuild(ctx, index)
}

func (adapter Adapter) beginRebuild(ctx context.Context, index string) error {
	adapter.projection.mutex.Lock()
	defer adapter.projection.mutex.Unlock()
	if adapter.projection.index != "" {
		return value.NewError("Conflict", "Reindex")
	}

	if err := adapter.freshIndex(ctx, index); err != nil {
		return err
	}

	adapter.projection.index = index
	adapter.projection.err = nil

	return nil
}

func (adapter Adapter) endRebuild() {
	adapter.projection.mutex.Lock()
	defer adapter.projection.mutex.Unlock()
	adapter.projection.index = ""
}

func (adapter Adapter) finishRebuild(ctx context.Context, index string) error {
	adapter.projection.mutex.Lock()
	defer adapter.projection.mutex.Unlock()
	if adapter.projection.err != nil {
		return adapter.projection.err
	}

	return adapter.alias(ctx, index, true)
}
