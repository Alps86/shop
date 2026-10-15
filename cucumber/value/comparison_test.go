package value_test

import (
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) comparisonSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^zwei exakte Dezimalwerte verglichen werden$`, current.compare)
}

func (current *world) compare(table *godog.Table) {
	row := table.Rows[1].Cells
	result, err := value.Decimal(row[0].Value).Compare(value.Decimal(row[1].Value))
	current.err = err
	current.canonical = "absent"
	if err == nil {
		current.canonical = strconv.Itoa(result)
	}
}
