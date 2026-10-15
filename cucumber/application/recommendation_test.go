package application_test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) registerRecommendations(ctx *godog.ScenarioContext) {
	ctx.Step(`^ein internes Empfehlungsgewicht geprüft wird$`, current.invalidInterest)
	ctx.Step(`^die Kandidatenquelle technisch ausfällt$`, current.candidatesFailure)
	ctx.Step(`^die vollständige Paarstatistik validiert wird$`, current.validStatistic)
	ctx.Step(`^ausdrücklich synthetische abgeschlossene Demowarenkörbe liefern Paarstatistik$`, current.pairs)
	ctx.Step(`^Empfehlungen angefragt werden$`, current.recommend)
	ctx.Step(`^lautet die öffentliche Angebotsreihenfolge$`, current.recommended)
	ctx.Step(`^die Empfehlungsquelle verändert und danach abgefragt wird$`, current.modifyRecommendations)
	ctx.Step(`^lautet das Empfehlungsergebnis$`, current.recommended)
	ctx.Step(`^Kandidaten besitzen gleiche Paarhäufigkeit$`, current.equalPairs)
	ctx.Step(`^die Statistik validiert oder ausgewertet wird$`, current.statisticOperation)
	ctx.Step(`^lautet das Statistikergebnis$`, current.recommended)
}

func (current *world) pairs(table *godog.Table) error {
	for _, row := range read(table) {
		if err := current.addPair(row); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) addPair(row map[string]string) error {
	count, err := strconv.ParseInt(row["BasketCount"], 10, 64)
	if err != nil {
		return err
	}

	start, err := time.Parse(time.RFC3339, row["WindowStart"])
	if err != nil {
		return err
	}

	end, err := time.Parse(time.RFC3339, row["WindowEnd"])
	if err != nil {
		return err
	}

	current.source.statistics = append(current.source.statistics, application.PairStatistic{SourceProductID: value.ProductID(row["SourceProductID"]), TargetProductID: value.ProductID(row["TargetProductID"]), BasketCount: count, WindowStart: start, WindowEnd: end, DatasetVersion: row["DatasetVersion"]})
	current.source.candidates = append(current.source.candidates, application.RecommendationCandidate{Card: application.ProductCard{ProductID: value.ProductID(row["TargetProductID"]), SKU: value.SKU(row["TargetProductID"]), UnitPrice: value.Money{MinorUnits: 2000, Currency: "EUR"}, AvailableQuantity: 1}, CategoryID: value.CategoryID(row["CategoryID"]), Compatible: row["Compatible"] == "true"})

	return nil
}

func (current *world) recommend(table *godog.Table) error {
	row := read(table)[0]
	weight, err := value.ParseDecimal(row["InterestWeight"])
	if err != nil {
		return err
	}

	limit, err := strconv.Atoi(row["Limit"])
	if err != nil {
		return err
	}

	current.query = application.RecommendationQuery{ProductID: value.ProductID(row["ProductID"]), SKU: value.SKU(row["SKU"]), Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}, Limit: limit, Interests: []sale.CategoryInterest{{CategoryID: value.CategoryID(row["Profile"]), Weight: weight}}}
	current.fetchRecommendations()

	return nil
}

func (current *world) fetchRecommendations() {
	current.recommendations, current.err = (application.RecommendationService{Source: current.source}).Recommend(context.Background(), current.query)
}

func (current *world) recommended(table *godog.Table) error {
	ids, skus := []string{}, []string{}
	for _, card := range current.recommendations.Cards {
		ids = append(ids, string(card.ProductID))
		skus = append(skus, string(card.SKU))
	}

	label := current.recommendations.Label
	if label == "" {
		label = "absent"
	}

	actual := map[string]string{"Code": value.Code(current.err), "ProductIDs": list(ids), "SelectedSKUs": list(skus), "Label": label, "Degraded": strconv.FormatBool(current.recommendations.Degraded), "ExposedMargin": "false", "ExposedInterestWeight": "false"}

	return read(table).matches(actual)
}

func list(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	return strings.Join(values, ",")
}

func (current *world) modifyRecommendations(table *godog.Table) error {
	row := read(table)[0]
	current.query = application.RecommendationQuery{ProductID: "p-phone", SKU: "PHONE", Limit: 3, Interests: []sale.CategoryInterest{{CategoryID: value.CategoryID(row["Profile"]), Weight: "1"}}}
	actions := map[string]func(){"chargerUnavailable": current.unavailable, "chargerIncompatible": current.incompatible, "noStatistics": current.noStatistics, "statisticsFailure": current.statisticsFailure, "noAccessories": current.noAccessories}
	action, exists := actions[row["Case"]]
	if !exists {
		return fmt.Errorf("unbekannter Fall: %s", row["Case"])
	}

	action()
	current.fetchRecommendations()

	return nil
}

func (current *world) unavailable() {
	current.source.candidates[1].Card.AvailableQuantity = 0
}

func (current *world) incompatible() {
	current.source.candidates[1].Compatible = false
}

func (current *world) noStatistics() {
	current.source.statistics = nil
}

func (current *world) statisticsFailure() {
	current.source.statistics = nil
	current.source.fault = "Statistics"
}

func (current *world) noAccessories() {
	current.source.candidates = nil
}

func (source recommendationSource) Candidates(_ context.Context, _ application.RecommendationQuery) ([]application.RecommendationCandidate, error) {
	if source.fault == "Candidates" {
		return nil, value.NewError("DependencyUnavailable", "Candidates")
	}

	return source.candidates, nil
}

func (source recommendationSource) Statistics(_ context.Context, _ value.ProductID) (application.PairStatistics, error) {
	if source.fault == "Statistics" {
		return nil, value.NewError("DependencyUnavailable", "Statistics")
	}

	return source.statistics, nil
}

func (current *world) equalPairs(table *godog.Table) error {
	current.source = recommendationSource{}
	for _, row := range read(table) {
		price, err := strconv.ParseInt(row["Price"], 10, 64)
		if err != nil {
			return err
		}

		current.source.candidates = append(current.source.candidates, application.RecommendationCandidate{Card: application.ProductCard{ProductID: value.ProductID(row["ProductID"]), SKU: value.SKU(row["SKU"]), UnitPrice: value.Money{MinorUnits: price, Currency: "EUR"}, AvailableQuantity: 1}, Compatible: row["Compatible"] == "true"})
	}

	current.source.statistics = application.PairStatistics{{SourceProductID: "p-phone", TargetProductID: "p-a", BasketCount: 5, WindowStart: current.now, WindowEnd: current.now.Add(90 * 24 * time.Hour), DatasetVersion: "demo-v1"}, {SourceProductID: "p-phone", TargetProductID: "p-b", BasketCount: 5, WindowStart: current.now, WindowEnd: current.now.Add(90 * 24 * time.Hour), DatasetVersion: "demo-v1"}}

	return nil
}

func (current *world) statisticOperation(table *godog.Table) error {
	row := read(table)[0]
	if row["Operation"] == "recommend" {
		current.query = application.RecommendationQuery{ProductID: "p-phone", SKU: "PHONE", Limit: 3}
		current.fetchRecommendations()

		return nil
	}

	changes := map[string]func(){"negativeCount": current.negativeStatistic, "duplicateUnorderedPair": current.duplicateStatistic, "inconsistentWindow": current.inconsistentStatistic, "unknownProduct": current.unknownStatistic}
	change, exists := changes[row["Input"]]
	if !exists {
		return fmt.Errorf("Unbekannter Statistikfall: %s", row["Input"])
	}

	change()
	current.err = current.source.statistics.Validate([]value.ProductID{"p-phone", "p-a", "p-b"})

	return nil
}

func (current *world) negativeStatistic() {
	current.source.statistics[0].BasketCount = -1
}

func (current *world) duplicateStatistic() {
	entry := current.source.statistics[0]
	entry.SourceProductID, entry.TargetProductID = entry.TargetProductID, entry.SourceProductID
	current.source.statistics = append(current.source.statistics, entry)
}

func (current *world) inconsistentStatistic() {
	current.source.statistics[1].WindowStart = current.now.Add(time.Hour)
	current.source.statistics[1].WindowEnd = current.source.statistics[1].WindowStart.Add(90 * 24 * time.Hour)
}

func (current *world) unknownStatistic() {
	current.source.statistics[0].TargetProductID = "unknown"
}

func (current *world) invalidInterest(table *godog.Table) {
	current.query = application.RecommendationQuery{ProductID: "p-phone", SKU: "PHONE", Limit: 3, Interests: []sale.CategoryInterest{{CategoryID: "mobile", Weight: value.Decimal(read(table)[0]["Weight"])}}}
	current.fetchRecommendations()
}

func (current *world) candidatesFailure(_ *godog.Table) {
	current.source.fault = "Candidates"
	current.query = application.RecommendationQuery{ProductID: "p-phone", SKU: "PHONE", Limit: 3}
	current.fetchRecommendations()
}

func (current *world) validStatistic(table *godog.Table) {
	products := []value.ProductID{}
	for _, id := range strings.Split(read(table)[0]["ProductIDs"], ",") {
		products = append(products, value.ProductID(id))
	}

	current.err = current.source.statistics.Validate(products)
}
