package runtime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

func (app *App) serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", app.Config.Listen)
	if err != nil {
		return err
	}

	return app.Serve(ctx, listener)
}

func (app *App) Serve(ctx context.Context, listener net.Listener) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	server := &http.Server{Handler: app.Handler, ReadHeaderTimeout: 5 * time.Second}
	stopped := make(chan error, 1)
	go func() {
		<-ctx.Done()
		stopped <- app.shutdown(server)
	}()
	err := server.Serve(listener)
	cancel()
	shutdownError := <-stopped
	if !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(err, shutdownError)
	}

	return shutdownError
}

func (app *App) shutdown(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
