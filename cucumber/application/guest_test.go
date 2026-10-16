package application_test

import (
	"context"
	"strconv"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerGuest(ctx *godog.ScenarioContext) {
	ctx.Step(`^zwei getrennte Browsersessions wurden geöffnet$`, current.browsers)
	ctx.Step(`^das Profil in einem Browser bestätigt wird$`, current.profile)
	ctx.Step(`^lautet die Gastansicht$`, current.guestViews)
	ctx.Step(`^eine Gastsession mit aktivem Cart besteht$`, current.session)
	ctx.Step(`^der Einkauf erneut gelesen wird$`, current.resume)
	ctx.Step(`^lautet die Wiederaufnahme$`, current.resumed)
}

func (current *world) browsers(table *godog.Table) {
	for _, row := range read(table) {
		user := value.UserID(row["UserID"])
		current.sessions[application.Session(row["Browser"])] = application.SessionIdentity{UserID: user, IssuedAt: current.now, CartID: value.Optional[value.CartID]{Present: true, Value: value.CartID(row["CartID"])}, Profile: application.ProfileCode(row["Profile"])}
		current.profiles[user] = application.ProfileCode(row["Profile"])
	}
}

func (current *world) profile(table *godog.Table) {
	row := read(table)[0]
	current.result, current.err = current.service.SelectProfile(context.Background(), application.Session(row["Browser"]), application.ProfileCode(row["Profile"]))
}

func (current *world) guestViews(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	for _, row := range read(table) {
		view, err := current.service.Open(context.Background(), application.Session(row["Browser"]))
		if err != nil {
			return err
		}

		if err := (rows{row}).matches(map[string]string{"Browser": row["Browser"], "UserID": string(view.UserID), "CartID": string(view.CartID.Value), "Profile": string(view.Profile)}); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) session(table *godog.Table) error {
	row := read(table)[0]
	issued, err := time.Parse(time.RFC3339, row["IssuedAt"])
	if err != nil {
		return err
	}

	active, err := time.Parse(time.RFC3339, row["LastActivityAt"])
	if err != nil {
		return err
	}

	current.sessions["valid"] = application.SessionIdentity{UserID: "u-a", IssuedAt: issued}
	current.carts["c-a"] = sale.Cart{ID: "c-a", UserID: "u-a", LastActivityAt: active, State: "open"}

	return nil
}

func (current *world) resume(table *godog.Table) error {
	row := read(table)[0]
	now, err := time.Parse(time.RFC3339, row["At"])
	if err != nil {
		return err
	}

	current.now = now
	current.cartView, current.err = current.service.GetCart(context.Background(), application.Session(row["Cookie"]), "c-a")

	return nil
}

func (current *world) resumed(table *godog.Table) error {
	if err := current.code(table); err != nil {
		return err
	}

	cart := current.carts["c-a"]
	_, exists := current.sessions[application.Session("valid")]
	state := string(cart.State)
	if cart.IsExpired(current.now) {
		state = "expired"
	}

	if value.Code(current.err) == "NotFound" {
		exists = false
		state = "inaccessible"
	}

	return read(table).matches(map[string]string{"Code": value.Code(current.err), "ExistingUser": strconv.FormatBool(exists), "CartState": state, "LastActivityAt": cart.LastActivityAt.Format(time.RFC3339)})
}

func (current *world) Resolve(_ context.Context, session application.Session) (application.SessionIdentity, error) {
	current.interestResolves++
	if current.fault == "guestResolve" {
		return application.SessionIdentity{}, value.NewError("DependencyUnavailable", "Guest")
	}

	identity, exists := current.sessions[session]
	if !exists {
		return application.SessionIdentity{}, value.NewError("NotFound", "Session")
	}

	identity.Profile = current.profiles[identity.UserID]

	return identity, nil
}

func (current *world) Open(_ context.Context) (application.GuestView, error) {
	if current.fault == "guestOpen" {
		return application.GuestView{}, value.NewError("DependencyUnavailable", "Guest")
	}

	current.sessions["new-session"] = application.SessionIdentity{UserID: "u-new", IssuedAt: current.now, Profile: "neutral"}
	current.profiles["u-new"] = "neutral"

	return application.GuestView{UserID: "u-new", Profile: "neutral", Session: "new-session"}, nil
}

func (current *world) Profile(_ context.Context, user value.UserID, profile application.ProfileCode) (application.GuestView, error) {
	if current.fault == "guestProfile" {
		return application.GuestView{}, value.NewError("DependencyUnavailable", "Guest")
	}

	current.profiles[user] = profile

	return application.GuestView{UserID: user, Profile: profile}, nil
}

func (current *world) Get(_ context.Context, id value.CartID) (sale.Cart, error) {
	if current.fault == "cartRead" {
		return sale.Cart{}, value.NewError("DependencyUnavailable", "Cart")
	}

	cart, exists := current.carts[id]
	if !exists {
		return sale.Cart{}, value.NewError("NotFound", "CartID")
	}

	return cart, nil
}

func (store orderStore) Get(_ context.Context, id value.OrderID) (sale.Order, error) {
	if store.current.fault == "orderRead" {
		return sale.Order{}, value.NewError("DependencyUnavailable", "Order")
	}

	order, exists := store.current.orders[id]
	if !exists {
		return sale.Order{}, value.NewError("NotFound", "OrderID")
	}

	return order, nil
}

func (store orderStore) ForCart(_ context.Context, id value.CartID) (value.OrderID, error) {
	if store.current.fault == "orderForCart" {
		return "", value.NewError("DependencyUnavailable", "Order")
	}

	for orderID, order := range store.current.orders {
		if order.CartID == id {
			return orderID, nil
		}
	}

	return "", value.NewError("NotFound", "Order")
}
