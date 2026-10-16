package demo_test

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

func (deployment *deployment) close() error {
	restoreError := deployment.restoreSearch()
	if deployment.external {
		if deployment.connection != nil {
			deployment.connection.Close()
		}

		if deployment.store != nil {
			deployment.store.Close()
		}

		return restoreError
	}

	return errors.Join(restoreError, deployment.closeNative())
}

func (deployment *deployment) closeNative() error {
	err := deployment.stopApp()
	err = errors.Join(err, stopEngine(deployment.varnish), stopEngine(deployment.nginx))
	if deployment.app != nil {
		err = errors.Join(err, deployment.app.Close(context.Background()), deployment.removeResources())
	}

	if deployment.search != nil {
		deployment.search.Close()
	}

	return errors.Join(err, os.RemoveAll(deployment.directory))
}

func (deployment *deployment) stopApp() error {
	if deployment.cancel == nil {
		return nil
	}

	deployment.cancel()
	err := <-deployment.finished
	deployment.cancel = nil
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func stopEngine(command *exec.Cmd) error {
	if command == nil {
		return nil
	}

	if command.Process == nil {
		return nil
	}

	if err := syscall.Kill(-command.Process.Pid, syscall.SIGTERM); err != nil {
		return err
	}

	_ = command.Wait()

	return nil
}

func (deployment *deployment) removeResources() error {
	ctx := context.Background()
	connection, err := pgx.Connect(ctx, deployment.config.DSN)
	if err != nil {
		return err
	}

	_, slot := connection.Exec(ctx, "SELECT pg_drop_replication_slot(slot_name) FROM pg_replication_slots WHERE slot_name=$1", deployment.config.Slot)
	_, publication := connection.Exec(ctx, "DROP PUBLICATION IF EXISTS "+pgx.Identifier{deployment.config.Publication}.Sanitize())
	_, schema := connection.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{deployment.config.Schema}.Sanitize()+" CASCADE")

	return errors.Join(slot, publication, schema, connection.Close(ctx), deployment.removeIndex())
}

func (deployment *deployment) removeIndex() error {
	request, err := http.NewRequest(http.MethodDelete, deployment.config.SearchEndpoint+"/"+deployment.config.SearchIndex+"*", nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	return response.Body.Close()
}
