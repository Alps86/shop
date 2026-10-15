package commerce_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
)

func (current *world) immediateChange(table *godog.Table) error {
	data := read(table)[0]
	if err := current.startCart(data); err != nil {
		return err
	}

	quantity, err := strconv.ParseInt(data["Quantity"], 10, 64)
	if err != nil {
		return err
	}

	current.view, current.err = current.app.Service.Change(context.Background(), current.guest.Session, application.ChangeCommand{CartID: current.view.Cart.ID, ExpectedRevision: current.view.Cart.Revision, CommandID: value.CommandID(data["ChangeCommandID"]), Operation: data["Operation"], SKU: value.SKU(data["SKU"]), Quantity: quantity})
	current.codes["Change"] = value.Code(current.err)

	return nil
}

func (current *world) startCart(data row) error {
	var err error
	current.guest, err = current.app.Service.Open(context.Background(), "")
	if err != nil {
		return err
	}

	current.view, err = current.app.Service.StartCart(context.Background(), current.guest.Session, value.CommandID(data["StartCommandID"]))

	current.cartID = string(current.view.Cart.ID)

	return err
}

func (current *world) port(ctx context.Context, data row) error {
	if data["Operation"] == "SearchPort.Search" {
		result, err := current.app.Search.Search(ctx, application.SearchQuery{Page: 1, Sort: "relevance", Context: current.config.Context})
		current.codes[data["Operation"]] = value.Code(err)
		current.answers[data["Operation"]] = map[string]string{"SearchTotal": strconv.FormatInt(result.TotalProducts, 10), "Cards": strconv.Itoa(len(result.Cards))}

		return nil
	}

	if data["Operation"] == "CatalogReader.Product" {
		result, err := current.app.Store.Catalog().Product(ctx, value.ProductID(data["ProductID"]), value.Optional[value.SKU]{Present: true, Value: value.SKU(data["SKU"])}, current.config.Context)
		current.codes[data["Operation"]] = value.Code(err)
		current.answers[data["Operation"]] = map[string]string{"ProductID": string(result.Product.ID)}

		return nil
	}

	return current.checkout(data)
}

func (current *world) missingCart(table *godog.Table) error {
	data := read(table)[0]
	_, err := current.app.Processes.EvaluateCart(context.Background(), application.EvaluateCommand{CartID: value.CartID(data["CartID"]), UserID: value.UserID(data["UserID"])})
	current.codes["Change"] = value.Code(err)

	return nil
}
