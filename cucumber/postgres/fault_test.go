package postgres_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v5"
	"strings"
)

func (current *world) faultSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein echter SQL-Schreibfehler gezielt vorbereitet ist$`, current.writeFault)
}

func (current *world) writeFault(input *godog.Table) error {
	row := table(input)
	schema := pgx.Identifier{current.schema}.Sanitize()
	function := schema + ".contract_fault"
	sql := "CREATE FUNCTION " + function + "() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'gezielter SQL-Speicherfehler' USING ERRCODE='" + strings.ReplaceAll(row["SQLState"], "'", "''") + "'; END $$"
	if _, err := current.admin.Exec(context.Background(), sql); err != nil {
		return err
	}

	condition := "true"
	if row["MatchField"] != "" {
		condition = fmt.Sprintf("to_jsonb(NEW)->>'%s'='%s'", strings.ReplaceAll(row["MatchField"], "'", "''"), strings.ReplaceAll(row["MatchValue"], "'", "''"))
	}

	table := pgx.Identifier{current.schema, row["Table"]}.Sanitize()
	sql = "CREATE TRIGGER contract_fault BEFORE " + row["Operation"] + " ON " + table + " FOR EACH ROW WHEN (" + condition + ") EXECUTE FUNCTION " + function + "()"
	_, err := current.admin.Exec(context.Background(), sql)

	return err
}
