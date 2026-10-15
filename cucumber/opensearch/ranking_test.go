package opensearch_test

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) rankingSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^vergleichbare synthetische Suchkandidaten liegen vor$`, current.candidates)
	ctx.Step(`^die Suche sortiert wird$`, current.sortedSearch)
	ctx.Step(`^lautet die öffentliche Ergebnisreihenfolge$`, current.orderObserved)
	ctx.Step(`^veröffentlichte Suchbegriffe sind vorhanden$`, current.project)
	ctx.Step(`^Vorschläge angefragt werden$`, current.suggest)
	ctx.Step(`^lauten die Vorschläge$`, current.suggestionsObserved)
}

func (current *world) candidates(table *godog.Table) error {
	current.config.PurchaseCosts = map[value.SKU]value.Money{}
	current.interests = nil
	for _, data := range read(table) {
		if err := current.candidate(data); err != nil {
			return err
		}
	}

	current.adapter = opensearch.New(current.config)

	return current.publish()
}

func (current *world) candidate(data row) error {
	if err := current.addProduct(data); err != nil {
		return err
	}

	id := value.ProductID(data["ProductID"])
	product := current.products[id]
	product.PrimaryCategoryID = value.CategoryID(data["CategoryID"])
	current.products[id] = product
	current.interests = append(current.interests, sale.CategoryInterest{CategoryID: product.PrimaryCategoryID, Weight: value.Decimal(data["InterestWeight"])})
	cost, err := strconv.ParseInt(data["PurchaseCost"], 10, 64)
	if err != nil {
		return err
	}

	current.config.PurchaseCosts[value.SKU(data["SKU"])] = value.Money{MinorUnits: cost, Currency: current.config.Context.Currency}

	return nil
}

func (current *world) publish() error {
	for id, product := range current.products {
		event := application.ProjectionEvent{ProductID: id, Revision: 1, Product: value.Optional[catalog.Product]{Present: true, Value: product}, Inventory: current.inventory[id]}
		if err := current.adapter.Apply(context.Background(), event); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) sortedSearch(table *godog.Table) error {
	return current.runSearch(read(table)[0])
}

func (current *world) runSearch(data row) error {
	page, err := strconv.Atoi(data["Page"])
	if err != nil {
		return err
	}

	current.query = application.SearchQuery{Text: data["Text"], Page: page, Sort: application.Sort(data["Sort"]), Context: current.config.Context, Interests: current.interests}
	if data.filterPresent("CategoryID") {
		current.query.CategoryID = value.Optional[value.CategoryID]{Present: true, Value: value.CategoryID(data["CategoryID"])}
	}

	current.requestedText = current.query.Text
	current.result, current.err = current.adapter.Search(context.Background(), current.query)

	return nil
}

func (current *world) orderObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	ids := []string{}
	for _, card := range current.result.Cards {
		ids = append(ids, string(card.ProductID))
	}

	payload, err := json.Marshal(current.result)
	if err != nil {
		return err
	}

	return read(table)[0].matches(map[string]string{"ProductIDs": strings.Join(ids, ","), "ExposedMargin": strconv.FormatBool(strings.Contains(string(payload), "\"Margin\"")), "ExposedInterestWeight": strconv.FormatBool(strings.Contains(string(payload), "\"InterestWeight\""))})
}

func (current *world) suggest(table *godog.Table) error {
	data := read(table)[0]
	if data["Query"] == "empty" {
		data["Query"] = ""
	}

	data["Text"] = data["Query"]

	return current.runSearch(data)
}

func (current *world) suggestionsObserved(table *godog.Table) error {
	if current.err != nil {
		return current.err
	}

	names, correction := current.suggestionValues()
	expected := read(table)[0]
	if err := expected.maximum(len(current.result.Suggestions)); err != nil {
		return err
	}

	delete(expected, "MaximumCount")

	return expected.matches(map[string]string{"Names": names, "Correction": strconv.FormatBool(correction), "QueryReplaced": strconv.FormatBool(current.queryReplaced())})
}

func (current *world) suggestionValues() (string, bool) {
	names := []string{}
	correction := false
	for _, item := range current.result.Suggestions {
		names = append(names, item.Text)
		correction = correction || item.Correction
	}

	text := strings.Join(names, ",")
	if text == "" {
		text = "[]"
	}

	return text, correction
}
