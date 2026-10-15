package runtime

import (
	"context"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
	"time"
)

func (app *App) project(ctx context.Context) error {
	app.projectionState("initializing")
	err := app.initializeProjection(ctx)
	if err == nil {
		app.projectionState("initialized")
		err = app.Consumer.Run(ctx)
	}

	app.projectionState(value.Code(err))
	<-ctx.Done()

	return ctx.Err()
}

func (app *App) initializeProjection(ctx context.Context) error {
	for {
		err := app.bootstrap(ctx)
		if err == nil {
			return nil
		}

		app.projectionState(value.Code(err))
		if value.Code(err) == "ReindexRequired" {
			return err
		}

		if err := app.pause(ctx); err != nil {
			return err
		}
	}
}

func (app *App) bootstrap(ctx context.Context) error {
	if err := app.Consumer.Prepare(ctx); err != nil {
		return err
	}

	if err := app.Search.Initialize(ctx); err != nil {
		return err
	}

	at := time.Now().UTC()

	return app.Search.Rebuild(ctx, app.Config.SearchIndex+"-"+strconv.FormatInt(at.UnixNano(), 10), at)
}

func (app *App) projectionState(status string) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.projectionStatus = status
}

func (app *App) pause(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
