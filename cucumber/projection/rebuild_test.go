package projection_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/application"
	"github.com/cucumber/godog"
	"time"
)

func (current *world) registerRebuild(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein echter Rebuild an einer kontrollierten WAL-Grenze wartet$`, current.beginRebuild)
	ctx.Step(`^Snapshot und WAL-Nachlauf vor dem Aliaswechsel freigegeben werden$`, current.finishRebuild)
}

func (source *gatedSource) Snapshot(ctx context.Context, at time.Time) (event.ReindexSnapshot, error) {
	snapshot, err := source.Source.Snapshot(ctx, at)
	if err != nil {
		return snapshot, err
	}

	if err := source.wait(ctx, source.SnapshotReady, source.SnapshotResume); err != nil {
		return snapshot, errors.Join(err, source.Source.Abort(context.WithoutCancel(ctx), snapshot.Cursor))
	}

	return snapshot, nil
}

func (source *gatedSource) Catchup(ctx context.Context, cursor string, target application.ProjectionPort) error {
	if err := source.Source.Catchup(ctx, cursor, target); err != nil {
		return err
	}

	return source.wait(ctx, source.CatchupReady, source.CatchupResume)
}

func (source *gatedSource) Abort(ctx context.Context, cursor string) error {
	return source.Source.Abort(ctx, cursor)
}

func (source *gatedSource) wait(ctx context.Context, ready, resume chan struct{}) error {
	if ready == nil {
		return nil
	}

	close(ready)
	select {
	case <-resume:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (current *world) beginRebuild(table *godog.Table) error {
	row := current.rows(table)[0]
	current.TargetIndex = current.Alias + "-" + row["IndexSuffix"]
	ready, resume := make(chan struct{}), make(chan struct{})
	gates := map[string]func(){"snapshot": func() {
		current.Source.SnapshotReady, current.Source.SnapshotResume = ready, resume
	}, "afterCatchup": func() {
		current.Source.CatchupReady, current.Source.CatchupResume = ready, resume
	}}
	gates[row["Boundary"]]()
	current.Rebuilt = make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		current.Rebuilt <- current.Search.Rebuild(ctx, current.TargetIndex, time.Now())
	}()

	return current.waitBoundary(ready)
}

func (current *world) waitBoundary(ready chan struct{}) error {
	select {
	case <-ready:
		return nil
	case err := <-current.Rebuilt:
		current.RebuildComplete = true
		return fmt.Errorf("Rebuild endete vor der Prüfgrenze: %w", err)
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}

func (current *world) finishRebuild() error {
	current.Source.release()

	select {
	case err := <-current.Rebuilt:
		current.RebuildComplete = true
		return err
	case <-time.After(10 * time.Second):
		return context.DeadlineExceeded
	}
}

func (source *gatedSource) release() {
	source.releaseGate(source.SnapshotResume)
	source.releaseGate(source.CatchupResume)
}

func (source *gatedSource) releaseGate(gate chan struct{}) {
	if gate == nil {
		return
	}

	select {
	case <-gate:
	default:
		close(gate)
	}
}

func (current *world) abortRebuild() error {
	if current.Rebuilt == nil {
		return nil
	}

	if current.RebuildComplete {
		return nil
	}

	return current.finishRebuild()
}
