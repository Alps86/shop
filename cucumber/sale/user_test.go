package sale_test

import (
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strings"
)

func (current *scenario) bindUser(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Kategorieinteressen eines Users validiert werden$`, current.userValidate)
	ctx.Step(`^lautet das Interessenvalidierungsergebnis$`, current.userResult)
}

func (current *scenario) userValidate(table *godog.Table) {
	row := rows(table)[0]
	user := sale.User{ID: value.UserID(input(row["UserID"]))}
	weights := strings.Split(row["Weights"], ",")
	for index, category := range strings.Split(row["CategoryIDs"], ",") {
		user.CategoryInterests = append(user.CategoryInterests, sale.CategoryInterest{CategoryID: value.CategoryID(category), Weight: value.Decimal(weights[index])})
	}

	categories := []value.CategoryID{}
	for _, category := range strings.Split(row["ExistingCategories"], ",") {
		categories = append(categories, value.CategoryID(category))
	}

	current.err = user.Validate(categories)
}

func (current *scenario) userResult(table *godog.Table) error {
	return check(table, map[string]string{"Code": value.Code(current.err)})
}
