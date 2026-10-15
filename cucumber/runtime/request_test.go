package runtime_test

import (
	"context"
	"errors"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (current *world) request(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	response, err := http.Get(current.baseURL + read(table)[0]["Path"])
	if err != nil {
		return err
	}

	current.status = response.StatusCode
	body, err := io.ReadAll(response.Body)
	current.body = string(body)

	return errors.Join(err, response.Body.Close())
}

func (current *world) cleanup(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	err := errors.Join(current.stopCommand(), current.stopRuntime())
	if current.app == nil {
		return ctx, err
	}

	return ctx, errors.Join(err, current.app.Close(ctx), current.releaseDatabase(ctx))
}

func (current *world) stopRuntime() error {
	if current.server != nil {
		current.server.Close()
	}

	if current.cancel != nil {
		current.cancel()
		if err := <-current.finished; !errors.Is(err, context.Canceled) {
			return err
		}
	}

	return nil
}

func (current *world) releaseDatabase(ctx context.Context) error {
	connection, err := pgx.Connect(ctx, current.config.DSN)
	if err != nil {
		return err
	}

	_, err = connection.Exec(ctx, "SELECT pg_drop_replication_slot(slot_name) FROM pg_replication_slots WHERE slot_name=$1", current.config.Slot)

	return errors.Join(err, current.removeResources(ctx, connection), connection.Close(ctx))
}

func (current *world) removeResources(ctx context.Context, connection *pgx.Conn) error {
	_, publicationError := connection.Exec(ctx, "DROP PUBLICATION IF EXISTS "+pgx.Identifier{current.config.Publication}.Sanitize())
	_, schemaError := connection.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{current.config.Schema}.Sanitize()+" CASCADE")
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, current.config.SearchEndpoint+"/"+current.config.SearchIndex+"*", nil)
	if err != nil {
		return errors.Join(publicationError, schemaError, err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Join(publicationError, schemaError, err)
	}

	return errors.Join(publicationError, schemaError, response.Body.Close())
}

func (current *world) matches(data map[string]string) bool {
	return strconv.Itoa(current.status) == data["Status"] && strings.Contains(current.body, data["Contains"])
}

func read(table *godog.Table) []map[string]string {
	result := []map[string]string{}
	for _, record := range table.Rows[1:] {
		data := map[string]string{}
		for index, cell := range record.Cells {
			data[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, data)
	}

	return result
}
