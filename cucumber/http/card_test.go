package http_test

import (
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"strconv"
	"strings"
)

func (current *world) prepareCards(table *godog.Table) error {
	current.fixture.Search.Cards = nil
	for _, fields := range read(table) {
		cards, err := cardSeries(fields)
		if err != nil {
			return err
		}

		current.fixture.Search.Cards = append(current.fixture.Search.Cards, cards...)
	}

	current.fixture.Search.TotalProducts = int64(len(current.fixture.Search.Cards))

	return nil
}

func cardSeries(fields map[string]string) ([]application.ProductCard, error) {
	var card application.ProductCard
	if err := json.Unmarshal([]byte(fields["Card"]), &card); err != nil {
		return nil, err
	}

	cards := []application.ProductCard{}
	for _, suffix := range strings.Split(fields["Suffixes"], ",") {
		copy := card
		copy.ProductID += value.ProductID(suffix)
		copy.SKU += value.SKU(suffix)
		copy.Name += suffix
		copy.URL += suffix
		cards = append(cards, copy)
	}

	return cards, nil
}

func (current *world) expectCardOrder(table *godog.Table) error {
	fields := read(table)
	expected, err := strconv.Atoi(fields[0]["Count"])
	if err != nil {
		return err
	}

	body := current.response.Body.String()
	if strings.Count(body, `<article class="card">`) != expected {
		return fmt.Errorf("Kartenanzahl entspricht nicht %d", expected)
	}

	return fields.orderedHTML(body)
}

func (expected rows) orderedHTML(body string) error {
	offset := 0
	for _, fields := range expected {
		part := fields["HTML"]
		index := strings.Index(body[offset:], part)
		if index < 0 {
			return fmt.Errorf("geordneter Kartenteil %q fehlt", part)
		}

		offset += index + len(part)
	}

	return nil
}
