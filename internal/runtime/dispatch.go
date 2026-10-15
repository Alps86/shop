package runtime

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"time"
)

func (app *App) dispatch(ctx context.Context) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			app.dispatchPending(ctx)
		}
	}
}

func (app *App) dispatchPending(ctx context.Context) {
	operation, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := (postgres.ProcessDispatcher{Store: app.Store, Processes: app.Processes}).Dispatch(operation)
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.dispatchError = err
	app.dispatchStarted = true
}
