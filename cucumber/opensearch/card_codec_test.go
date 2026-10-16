package opensearch_test

import (
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
)

func (current *world) cardCodecSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine vollständige technische Suchkarte vorbereitet ist$`, current.prepareCardCodec)
	ctx.Step(`^der vollständige Kartenpayload im expliziten Base64format vorliegt$`, current.cardWire)
	ctx.Step(`^die gespeicherte Karte über den öffentlichen Suchdecoder abgefragt wird$`, current.searchWireCard)
	ctx.Step(`^die Suchmappingfelder die vereinbarten nativen DocValues bereitstellen$`, current.mappingObserved)
	ctx.Step(`^der gewählte Varianttreffer ausschließlich native DocValues anfordert$`, current.nestedDocValues)
	ctx.Step(`^das Projektionsdokument seinen gespeicherten Kartenpayload erzeugt$`, current.encodeCardDocument)
	ctx.Step(`^der Kartenpayload ausschließlich echte Nullwerte auslässt$`, current.cardFieldsObserved)
	ctx.Step(`^der bestehende Suchdecoder die vollständige Karte unverändert rekonstruiert$`, current.cardRoundtripObserved)
}

func (current *world) prepareCardCodec(table *godog.Table) error {
	data := read(table)[0]
	payload, err := os.ReadFile("../../" + data["Fixture"])
	if err != nil {
		return err
	}

	result := application.SearchResult{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return err
	}

	current.codec = &cardCodecProbe{original: result.Cards[0]}
	current.prepareCodecOrigin()
	current.codec.original.Compliance = catalog.ProductCompliance{}

	return json.Unmarshal([]byte(data["Compliance"]), &current.codec.original.Compliance)
}

func (current *world) encodeCardDocument() error {
	document := opensearch.Document{ID: current.codec.original.ProductID, Variants: []opensearch.Variant{{SKU: current.codec.original.SKU, Card: current.codec.original}}}
	payload, err := json.Marshal(document)
	if err != nil {
		return err
	}

	return json.Unmarshal(payload, &current.codec.document)
}

func (current *world) cardFieldsObserved(table *godog.Table) error {
	fields := map[string]json.RawMessage{}
	payload := map[string]json.RawMessage{}
	if err := json.Unmarshal(current.codec.document.Variants[0].CardPayload, &payload); err != nil {
		return err
	}

	if err := json.Unmarshal(payload["Compliance"], &fields); err != nil {
		return err
	}

	for _, data := range read(table) {
		if _, exists := fields[data["Field"]]; fmt.Sprint(exists) != data["Present"] {
			return fmt.Errorf("gespeichertes Compliancefeld %s: vorhanden=%t statt %s", data["Field"], exists, data["Present"])
		}
	}

	return nil
}

func (current *world) cardRoundtripObserved(table *godog.Table) error {
	if !reflect.DeepEqual(current.codec.original, current.codec.document.Variants[0].Card) {
		return fmt.Errorf("ursprüngliche Sourcekarte verändert")
	}

	current.msearch.response = current.codec.response()
	current.adapter = opensearch.New(current.config)
	if err := current.runSearch(read(table)[0]); err != nil {
		return err
	}

	if !reflect.DeepEqual(current.result.Cards, []application.ProductCard{current.codec.original}) {
		return fmt.Errorf("vollständige rekonstruierte Karte weicht ab")
	}

	return nil
}

func (probe cardCodecProbe) response() []byte {
	payload, _ := json.Marshal(probe.document.Variants[0].CardPayload)
	if probe.wirePayload.Present {
		payload, _ = json.Marshal(probe.wirePayload.Value)
	}

	id, _ := json.Marshal(probe.original.ProductID)
	sku, _ := json.Marshal(probe.original.SKU)
	response := `{"hits":{"hits":[{"fields":{"product_id":[ID]},"inner_hits":{"chosen":{"hits":{"hits":[{"fields":{"variants.sku":[SKU],"variants.card_payload":[PAYLOAD]}}]}}}}]}}`

	return []byte(strings.NewReplacer("ID", string(id), "SKU", string(sku), "PAYLOAD", string(payload)).Replace(response))
}

func (current *world) prepareCodecOrigin() {
	current.msearch = &msearchProbe{status: 200}
	current.msearch.server = httptest.NewServer(http.HandlerFunc(current.msearch.serve))
	current.config = opensearch.Config{Endpoint: current.msearch.server.URL, Index: "codec", Context: value.RequestContext{Locale: "de-DE", Currency: "EUR", Market: "DE"}}
}
