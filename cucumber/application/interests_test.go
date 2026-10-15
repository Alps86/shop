package application_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) registerInterests(ctx *godog.ScenarioContext) {
	ctx.Step(`^persistierte Interessen getrennt nach Gast vorliegen$`, current.persistedInterests)
	ctx.Step(`^Interessen über die Gastfähigkeit angefordert werden$`, current.requestInterests)
	ctx.Step(`^lautet die autorisierte Interessensicht$`, current.interestsObserved)
}

func (current *world) persistedInterests(input *godog.Table) {
	current.interestData = map[value.UserID][]sale.CategoryInterest{}
	current.service.Interests = interestReader{current}
	for _, row := range read(input) {
		user := value.UserID(row["UserID"])
		current.interestData[user] = append(current.interestData[user], sale.CategoryInterest{CategoryID: value.CategoryID(row["CategoryID"]), Weight: value.Decimal(row["Weight"])})
	}
}

func (current *world) requestInterests(input *godog.Table) {
	row := read(input)[0]
	current.fault = row["Fault"]
	current.interestResult, current.err = current.service.GetInterests(context.Background(), application.Session(row["Session"]))
}

func (current *world) interestsObserved(input *godog.Table) error {
	category, weight := "", ""
	if len(current.interestResult) > 0 {
		category = string(current.interestResult[0].CategoryID)
		weight = string(current.interestResult[0].Weight)
	}

	return read(input).matches(map[string]string{"Code": value.Code(current.err), "Count": strconv.Itoa(len(current.interestResult)), "CategoryID": category, "Weight": weight, "ReadUserID": string(current.interestUser)})
}

func (port interestReader) ForUser(_ context.Context, user value.UserID) ([]sale.CategoryInterest, error) {
	port.current.interestUser = user
	if port.current.fault == "interestRead" {
		return nil, value.NewError("DependencyUnavailable", "Interests")
	}

	return port.current.interestData[user], nil
}
