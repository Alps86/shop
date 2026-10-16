package opensearch_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/cucumber/godog"
	"net/http"
	"reflect"
	"strconv"
)

func (current *world) cacheSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^der echte Suchantwortkörper zur Transportprüfung beobachtet wird$`, current.captureResponse)
	ctx.Step(`^der echte Suchindex seinen Requestcachezähler vormerkt$`, current.rememberCache)
	ctx.Step(`^die identische öffentliche Suche mehrfach ausgeführt wird$`, current.repeatSearch)
	ctx.Step(`^der Suchindex echte Requestcachetreffer meldet$`, current.cacheObserved)
	ctx.Step(`^die echte Suchantwort nur benötigte Rootdaten und die vollständige Variantkarte enthält$`, current.sourceObserved)
}

func (current *world) rememberCache() error {
	count, err := current.cacheHits()
	current.cacheBefore = count

	return err
}

func (current *world) cacheHits() (int64, error) {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, current.config.Endpoint+"/"+current.config.Index+"/_stats/request_cache", nil)
	if err != nil {
		return 0, err
	}

	response, err := current.config.HTTP.Do(request)
	if err != nil {
		return 0, err
	}

	counts := cacheStatistics{}
	err = json.NewDecoder(response.Body).Decode(&counts)

	return counts.All.Primaries.RequestCache.Hits, errors.Join(err, response.Body.Close())
}

func (current *world) repeatSearch(table *godog.Table) error {
	count, err := strconv.Atoi(read(table)[0]["Attempts"])
	if err != nil {
		return err
	}

	for range count {
		current.result, current.err = current.adapter.Search(context.Background(), current.query)
		if current.err != nil {
			return current.err
		}
	}

	return nil
}

func (current *world) cacheObserved(table *godog.Table) error {
	minimum, err := strconv.ParseInt(read(table)[0]["MinimumHitDelta"], 10, 64)
	if err != nil {
		return err
	}

	count, err := current.cacheHits()
	if err != nil {
		return err
	}

	if count-current.cacheBefore < minimum {
		return fmt.Errorf("Requestcachetrefferdelta %d statt mindestens %d", count-current.cacheBefore, minimum)
	}

	return nil
}

func (current *world) sourceObserved(table *godog.Table) error {
	var envelope sourceEnvelope
	if err := json.Unmarshal(current.capture.response, &envelope); err != nil {
		return err
	}

	if len(envelope.Responses) != 2 {
		return fmt.Errorf("zwei echte Teilantworten benötigt")
	}

	if len(envelope.Responses[1].Hits.Hits) != 0 {
		return fmt.Errorf("Completion darf keine Roottreffer laden")
	}

	response := envelope.Responses[0]

	fmt.Printf("Echte Suchantwort: %d Bytes\n", len(current.capture.response))

	return current.sourceHits(response.Hits.Hits, read(table)[0])
}

func (current *world) sourceHits(hits []map[string]json.RawMessage, expected row) error {
	for _, hit := range hits {
		if err := fetchMetadataAbsent(hit); err != nil {
			return err
		}

		if err := expected.docValueHit(hit); err != nil {
			return err
		}
	}

	return current.sameCard()
}

func (expected row) docValueHit(hit map[string]json.RawMessage) error {
	if err := expected.docValueRoot(hit["fields"]); err != nil {
		return err
	}

	inner := map[string]sourceResponse{}
	if err := json.Unmarshal(hit["inner_hits"], &inner); err != nil {
		return err
	}

	return docValueChosen(inner["chosen"].Hits.Hits)
}

func docValueChosen(hits []map[string]json.RawMessage) error {
	if len(hits) != 1 {
		return fmt.Errorf("genau ein DocValues-Varianttreffer benötigt")
	}

	if err := fetchMetadataAbsent(hits[0]); err != nil {
		return err
	}

	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(hits[0]["fields"], &fields); err != nil {
		return err
	}

	if len(fields) != 2 {
		return fmt.Errorf("genau SKU und vollständige Payload als DocValues-Variantfelder benötigt")
	}

	return nil
}

func (expected row) docValueRoot(payload []byte) error {
	root := map[string][]string{}
	if err := json.Unmarshal(payload, &root); err != nil {
		return err
	}

	ids := root["product_id"]
	if len(ids) != 1 {
		return fmt.Errorf("genau eine ProductID aus DocValues benötigt")
	}

	return expected.matches(row{"RootFields": fmt.Sprint(len(root)), "ProductID": ids[0]})
}

func (current *world) sameCard() error {
	if len(current.result.Cards) != 1 {
		return fmt.Errorf("eine echte Modellkarte benötigt")
	}

	product := current.products[current.result.Cards[0].ProductID]
	card := opensearch.NewCard(product, product.Variants[0], current.inventory[product.ID][0], current.config)

	if !reflect.DeepEqual(card, current.result.Cards[0]) {
		return fmt.Errorf("gewählte Suchkarte entspricht nicht vollständiger öffentlicher Katalogkarte")
	}

	return nil
}

func (current *world) captureResponse(table *godog.Table) error {
	capture, err := strconv.ParseBool(read(table)[0]["CaptureResponse"])
	current.capture.observeResponse = capture

	return err
}

func fetchMetadataAbsent(hit map[string]json.RawMessage) error {
	for _, key := range []string{"_source", "_id"} {
		if _, exists := hit[key]; exists {
			return fmt.Errorf("DocValues-Treffer darf %s nicht laden", key)
		}
	}

	return nil
}
