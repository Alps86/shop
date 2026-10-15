package postgres_test

import (
	"context"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (current *world) readFaultSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^folgende SQL-Fixtureänderung nur im eigenen Schema erfolgt$`, current.sqlFixture)
}

func (current *world) sqlFixture(input *godog.Table) error {
	for _, row := range input.Rows[1:] {
		statement := strings.ReplaceAll(row.Cells[0].Value, "$schema", pgx.Identifier{current.schema}.Sanitize())
		if _, err := current.admin.Exec(context.Background(), statement); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) readDiagnostics(_ map[string]string) error {
	current.observed, current.err = current.store.Diagnostics(context.Background())
	if current.observed == nil {
		current.observed = map[string]string{}
	}

	return nil
}
