package runtime

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"time"
)

func (config Config) Run(ctx context.Context) (err error) {
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	app, err := Open(startup, config)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, app.Close(context.WithoutCancel(ctx)))
	}()
	err = app.Run(ctx)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func (app *App) Run(ctx context.Context) error {
	worker := app.Processes.Worker()
	if err := worker.Start(); err != nil {
		return err
	}

	defer worker.Stop()
	group, running := errgroup.WithContext(ctx)
	for _, task := range []func(context.Context) error{app.project, app.dispatch, app.serve, app.serveBackend} {
		group.Go(func() error {
			return task(running)
		})
	}

	return group.Wait()
}
