package postgres_test

import (
	"context"
	"strconv"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) purchaseSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein Gast einen Warenkorb mit einer Position dauerhaft anlegt$`, current.newCart)
	ctx.Step(`^der Gast den vollständigen Checkout bestätigt$`, current.checkout)
	ctx.Step(`^der Gast den bestätigten Kauf wiederholt abschließt$`, current.purchase)
	ctx.Step(`^lautet die dauerhafte Kaufbeobachtung$`, current.purchaseObserved)
}

func (current *world) Now() time.Time {
	return time.Now().UTC()
}

func (current *world) newCart(input *godog.Table) error {
	current.service = application.Service{Guests: current.store.Guests(), Carts: current.store.Carts(), Orders: current.store.Orders(), Fulfillment: current.store.Fulfillment(), Writer: current.store.Carts(), Starter: current.store.Carts(), Evaluator: current.store.Evaluator(), Purchase: current.store.Purchases(), Clock: current, Payment: current, Commands: current.store.Carts()}
	guest, err := current.service.Open(context.Background(), "")
	if err != nil {
		return err
	}

	current.guest = guest
	cart, err := current.service.StartCart(context.Background(), guest.Session, "start-a")
	if err != nil {
		return err
	}

	quantity, err := strconv.ParseInt(table(input)["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	current.cart, current.err = current.service.ApplyChange(context.Background(), application.ChangeCommand{UserID: guest.UserID, CartID: cart.Cart.ID, CommandID: "add-a", ExpectedRevision: cart.Cart.Revision, Operation: "add", SKU: value.SKU(table(input)["SKU"]), Quantity: quantity})

	return current.err
}

func (current *world) checkout(input *godog.Table) error {
	row := table(input)
	cart := current.cart.Cart
	address := sale.Address{Recipient: row["Recipient"], Street: row["Street"], PostalCode: row["PostalCode"], City: row["City"], Country: value.CountryCode(row["Country"])}
	view, err := current.service.ApplyChange(context.Background(), application.ChangeCommand{UserID: current.guest.UserID, CartID: cart.ID, CommandID: "address-a", ExpectedRevision: cart.Revision, Operation: "address", Address: address})
	if err != nil {
		return err
	}

	view, err = current.payment(view, row)
	if err != nil {
		return err
	}

	current.cart, err = current.service.ApplyConfirm(context.Background(), application.ConfirmCommand{UserID: current.guest.UserID, CartID: cart.ID, CommandID: "confirm-a", EvaluationID: view.Cart.Evaluation.Value.ID, ExpectedRevision: view.Cart.Revision})

	return err
}

func (current *world) purchase(input *godog.Table) error {
	row := table(input)
	attempts, err := strconv.Atoi(row["Attempts"])
	if err != nil {
		return err
	}

	command := application.PlaceCommand{UserID: current.guest.UserID, CartID: current.cart.Cart.ID, CommandID: value.CommandID(row["CommandID"]), EvaluationID: current.cart.Cart.Evaluation.Value.ID}
	for range attempts {
		current.result, current.err = current.service.ApplyPlace(context.Background(), command)
	}

	return nil
}

func (current *world) purchaseObserved(input *godog.Table) error {
	page, err := current.store.AdminOrders().List(context.Background(), 1)
	if err != nil {
		return err
	}

	snapshot, err := current.store.Catalog().Snapshot(context.Background())
	if err != nil {
		return err
	}

	return check(table(input), map[string]string{"Code": value.Code(current.err), "OrderCount": strconv.FormatInt(page.Total, 10), "Reserved": strconv.FormatInt(snapshot.Inventory[0].ReservedQuantity(), 10), "Total": strconv.FormatInt(page.Orders[0].Purchase.Total.MinorUnits, 10), "CartState": string(current.cartState())})
}

func (current *world) cartState() sale.CartState {
	cart, err := current.store.Carts().Get(context.Background(), current.cart.Cart.ID)
	if err != nil {
		return "unknown"
	}

	return cart.State
}
