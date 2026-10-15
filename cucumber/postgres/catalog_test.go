package postgres_test

import (
	"context"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"strconv"
)

func (current *world) requestContext(row map[string]string) value.RequestContext {
	request := value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}
	if row["Locale"] != "" {
		request.Locale = value.Locale(row["Locale"])
	}

	return request
}

func (current *world) readProduct(row map[string]string) error {
	selected := value.Optional[value.SKU]{}
	if row["SKU"] != "" {
		selected = value.Optional[value.SKU]{Value: value.SKU(row["SKU"]), Present: true}
	}

	detail, err := current.store.Catalog().Product(context.Background(), value.ProductID(row["ID"]), selected, current.requestContext(row))
	current.err = err
	current.observed["ChosenSKU"] = string(detail.ChosenSKU)
	current.observed["Cards"] = strconv.Itoa(len(detail.Cards))
	current.observeChosen(detail)

	return nil
}

func (current *world) observeChosen(detail application.ProductDetail) {
	for _, card := range detail.Cards {
		if card.SKU != detail.ChosenSKU {
			continue
		}

		current.observed["URL"] = card.URL
		current.observed["ImageURL"] = card.Image.URL
		current.observed["Available"] = strconv.FormatInt(card.AvailableQuantity, 10)
		current.observed["VerificationStatus"] = card.Compliance.Context.VerificationStatus
		current.observed["Warnings"] = strconv.Itoa(len(card.Compliance.Safety.Warnings))
		current.observed["TopAttributes"] = strconv.Itoa(len(card.TopAttributes))
		current.observed["ProductFlag"] = strconv.FormatBool(card.TopAttributes["ready"].Boolean.Value)
		current.observed["VariantFlag"] = strconv.FormatBool(card.TopAttributes["wireless"].Boolean.Value)
	}
}

func (current *world) readCategory(row map[string]string) error {
	view, err := current.store.Catalog().Category(context.Background(), value.CategoryID(row["ID"]), current.requestContext(row))
	current.err = err
	current.observed["Breadcrumb"] = strconv.Itoa(len(view.Breadcrumb))
	current.observed["Children"] = strconv.Itoa(len(view.Children))
	current.observed["FilterDefinitions"] = strconv.Itoa(len(view.FilterDefinitions))
	current.observed["TopDefinitions"] = strconv.Itoa(len(view.TopDefinitions))

	return nil
}

func (current *world) readOrders(row map[string]string) error {
	page, err := strconv.Atoi(row["Page"])
	if err != nil {
		return err
	}

	result, err := current.store.AdminOrders().List(context.Background(), page)
	current.err = err
	current.observed["Total"] = strconv.FormatInt(result.Total, 10)

	return nil
}

func (current *world) readRecommendations(row map[string]string) error {
	source := current.store.Recommendations()
	query := application.RecommendationQuery{ProductID: value.ProductID(row["ID"]), SKU: value.SKU(row["SKU"]), Context: current.requestContext(row), Limit: 3}
	result, err := (application.RecommendationService{Source: source}).Recommend(context.Background(), query)
	current.err = err
	current.observed["Count"] = strconv.Itoa(len(result.Cards))
	current.observed["Degraded"] = strconv.FormatBool(result.Degraded)
	current.observed["Label"] = result.Label
	if len(result.Cards) > 0 {
		current.observed["SKU"] = string(result.Cards[0].SKU)
	}

	return nil
}

func (current *world) readOrder(row map[string]string) error {
	id := current.result.OrderID
	if row["ID"] == "missing" {
		id = "o-missing"
	}

	order, err := current.store.Orders().Get(context.Background(), id)
	current.err = err
	current.observed["OwnOrder"] = strconv.FormatBool(order.ID == current.result.OrderID)

	return nil
}

func (current *world) readAdminOrder(row map[string]string) error {
	id := current.result.OrderID
	if row["ID"] == "missing" {
		id = "o-missing"
	}

	order, err := current.store.AdminOrders().Get(context.Background(), id)
	current.err = err
	current.observed["OwnOrder"] = strconv.FormatBool(order.Purchase.ID == current.result.OrderID)

	return nil
}

func (current *world) readCartOrder(row map[string]string) error {
	id := current.cart.Cart.ID
	if row["ID"] == "missing" {
		id = "c-missing"
	}

	order, err := current.store.Orders().ForCart(context.Background(), id)
	current.err = err
	current.observed["OwnOrder"] = strconv.FormatBool(order == current.result.OrderID)

	return nil
}

func (current *world) readVariant(row map[string]string) error {
	detail, err := current.store.Catalog().ProductForSKU(context.Background(), value.SKU(row["SKU"]), current.requestContext(row))
	current.err = err
	current.observed["ChosenSKU"] = string(detail.ChosenSKU)
	current.observed["Cards"] = strconv.Itoa(len(detail.Cards))

	return nil
}
