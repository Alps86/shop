package application_test

import (
	"context"
	"errors"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"time"
)

func (current *world) registerInterests(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Interessenfähigkeit folgendes Alter hat$`, current.interestAge)
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

	var failure value.Error
	field := ""
	if errors.As(current.err, &failure) {
		field = failure.Field
	}

	return read(input).matches(map[string]string{"Field": field, "Code": value.Code(current.err), "Count": strconv.Itoa(len(current.interestResult)), "CategoryID": category, "Weight": weight, "ReadUserID": string(current.interestUser), "ResolveCalls": strconv.Itoa(current.interestResolves), "ReadCalls": strconv.Itoa(current.interestReads)})
}

func (port interestReader) ForSession(_ context.Context, session application.Session) (application.SessionInterests, error) {
	port.current.interestReads++
	identity, exists := port.current.sessions[session]
	if !exists {
		return application.SessionInterests{}, value.NewError("NotFound", "Session")
	}

	port.current.interestUser = identity.UserID
	if port.current.fault == "interestRead" {
		return application.SessionInterests{}, value.NewError("DependencyUnavailable", "Interests")
	}

	return port.interestSnapshot(identity), nil
}

func (port interestReader) interestSnapshot(identity application.SessionIdentity) application.SessionInterests {
	snapshot := application.SessionInterests{IssuedAt: identity.IssuedAt, Interests: port.current.interestData[identity.UserID]}
	if port.current.fault == "payload" {
		snapshot.PayloadError = value.NewError("DependencyUnavailable", "PostgreSQL")
	}

	return snapshot
}

func (current *world) interestAge(input *godog.Table) error {
	row := read(input)[0]
	age, err := time.ParseDuration(row["Age"])
	if err != nil {
		return err
	}

	identity := current.sessions[application.Session(row["Session"])]
	identity.IssuedAt = current.now.Add(-age)
	current.sessions[application.Session(row["Session"])] = identity

	return nil
}
