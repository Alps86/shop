package opensearch_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "progress", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @opensearch", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "opensearch", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("OpenSearchverträge fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Before(current.before)
	ctx.After(current.cleanup)
	ctx.Step(`^eine echte OpenSearchverbindung besteht$`, current.connect)
	current.projectionSteps(ctx)
	current.searchSteps(ctx)
	current.rankingSteps(ctx)
	current.reindexSteps(ctx)
	current.failureSteps(ctx)
	current.budgetSteps(ctx)
	current.cacheSteps(ctx)
	current.attributeSteps(ctx)
	current.poolSteps(ctx)
	current.msearchSteps(ctx)
	ctx.Step(`^der veröffentlichte Kategorienbaum steht bereit$`, current.categories)
	ctx.Step(`^der Suchindex als leere echte Fixture angelegt wird$`, current.emptyIndex)
	ctx.Step(`^die echte öffentliche Suche abgefragt wird$`, current.search)
	ctx.Step(`^liefert die Suchverbindung$`, current.observed)
}

func (current *world) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	*current = world{}

	return ctx, nil
}

func (current *world) connect(table *godog.Table) {
	data := read(table)[0]
	data["Index"] += "-" + strings.ToLower(rand.Text()[:10])
	current.config = opensearch.Config{Endpoint: os.Getenv("SHOP_TEST_OPENSEARCH_URL"), Index: data["Index"], Context: value.RequestContext{Locale: value.Locale(data["Locale"]), Currency: value.Currency(data["Currency"]), Market: value.CountryCode(data["Market"])}, Clock: current}
	current.adapter = opensearch.New(current.config)
}

func (current *world) search(table *godog.Table) {
	data := read(table)[0]
	page, err := strconv.Atoi(data["Page"])
	if err != nil {
		current.err = err
		return
	}

	current.query = application.SearchQuery{Text: data["Text"], Page: page, Sort: application.Sort(data["Sort"]), Context: current.config.Context}
	current.measuredSearch()
}

func (current *world) observed(table *godog.Table) error {
	actual := map[string]string{"Code": value.Code(current.err), "TotalProducts": strconv.FormatInt(current.result.TotalProducts, 10)}
	for index, heading := range table.Rows[0].Cells {
		expected := table.Rows[1].Cells[index].Value
		if actual[heading.Value] != expected {
			return fmt.Errorf("%s: %q statt %q", heading.Value, actual[heading.Value], expected)
		}
	}

	return nil
}

func (current *world) emptyIndex() error {
	return current.adapter.Initialize(context.Background())
}

func read(table *godog.Table) []row {
	result := []row{}
	for _, record := range table.Rows[1:] {
		data := row{}
		for index, cell := range record.Cells {
			data[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, data)
	}

	return result
}
