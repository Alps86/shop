package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) readSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine neue echte Gastfähigkeit geöffnet wird$`, current.openGuest)
	ctx.Step(`^kuratierte Produkt- und Variantenattribute vorbereitet werden$`, current.curated)
	ctx.Step(`^Modellbilder und eine Unterkategorie vorbereitet werden$`, current.productPresentation)
	ctx.Step(`^ein persistenter Lesevertrag aufgerufen wird$`, current.readContract)
	ctx.Step(`^lautet das öffentliche SQL-Leseergebnis$`, current.readObserved)
	ctx.Step(`^das persistierte Gastprofil gewählt wird$`, current.selectProfile)
	ctx.Step(`^die Gastfähigkeit um Tage gealtert wird$`, current.ageSession)
}

func (current *world) openGuest() error {
	current.guest, current.err = current.store.Guests().Open(context.Background())
	return current.err
}

func (current *world) readContract(input *godog.Table) error {
	row := table(input)
	current.observed = map[string]string{}
	operations := map[string]func(map[string]string) error{"diagnostics": current.readDiagnostics, "resolve": current.resolveGuest, "interests": current.readInterests, "product": current.readProduct, "variant": current.readVariant, "category": current.readCategory, "adminOrders": current.readOrders, "recommendations": current.readRecommendations, "order": current.readOrder, "adminOrder": current.readAdminOrder, "cartOrder": current.readCartOrder}

	return operations[row["Operation"]](row)
}

func (current *world) readObserved(input *godog.Table) error {
	current.observed["Code"] = value.Code(current.err)
	return check(table(input), current.observed)
}

func (current *world) resolveGuest(row map[string]string) error {
	session := current.guest.Session
	if row["Session"] == "unknown" {
		session = "unknown"
	}

	identity, err := current.store.Guests().Resolve(context.Background(), session)
	current.err = err
	current.observed["OwnUser"] = strconv.FormatBool(identity.UserID == current.guest.UserID)
	current.observed["Profile"] = string(identity.Profile)
	current.observed["CartPresent"] = strconv.FormatBool(identity.CartID.Present)
	var rawStored bool
	if err := current.admin.QueryRow(context.Background(), "SELECT COALESCE(bool_or(position($1 in row_to_json(sessions)::text)>0),false) FROM "+current.schema+".sessions", session).Scan(&rawStored); err != nil {
		return err
	}

	current.observed["RawStored"] = strconv.FormatBool(rawStored)
	current.observed["TokenBytes"] = strconv.Itoa(len(current.guest.Session) / 2)

	return nil
}

func (current *world) readInterests(row map[string]string) error {
	id := current.guest.UserID
	if row["User"] == "unknown" {
		id = "u-unknown"
	}

	interests, err := current.store.Guests().ForUser(context.Background(), id)
	current.err = err
	current.observed["Count"] = strconv.Itoa(len(interests))
	if len(interests) > 0 {
		current.observed["CategoryID"] = string(interests[0].CategoryID)
		current.observed["Weight"] = string(interests[0].Weight)
	}

	return nil
}

func (current *world) selectProfile(input *godog.Table) {
	row := table(input)
	id := current.guest.UserID
	if row["User"] == "unknown" {
		id = "u-unknown"
	}

	_, current.err = current.store.Guests().Profile(context.Background(), id, application.ProfileCode(row["Profile"]))
	current.observed = map[string]string{}
}

func (current *world) ageSession(input *godog.Table) error {
	days, err := strconv.Atoi(table(input)["AgeDays"])
	if err != nil {
		return err
	}

	_, err = current.admin.Exec(context.Background(), "UPDATE "+current.schema+".sessions SET issued_at=now()-($1*interval '1 day')", days)

	return err
}
