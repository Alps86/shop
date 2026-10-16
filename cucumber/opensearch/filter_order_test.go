package opensearch_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/cucumber/godog"
	"reflect"
	"strconv"
)

func (current *world) filterOrderSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Filterabfrage identische vollständige Suchantworten und Drahtbytes behält$`, current.stableFilterWire)
}

func (current *world) stableFilterWire(table *godog.Table) error {
	data := read(table)[0]
	count, err := strconv.Atoi(data["Attempts"])
	if err != nil {
		return err
	}

	baseline := current.result
	payload := bytes.Clone(current.capture.payload)
	for range count {
		if err := current.repeatStableFilter(baseline, payload); err != nil {
			return err
		}
	}

	return current.canonicalFilterWire(data)
}

func (current *world) repeatStableFilter(baseline application.SearchResult, payload []byte) error {
	result, err := current.adapter.Search(context.Background(), current.query)
	if err != nil {
		return err
	}

	if !bytes.Equal(current.capture.payload, payload) {
		return fmt.Errorf("identische Filter erzeugen abweichende Drahtbytes")
	}

	if !reflect.DeepEqual(result, baseline) {
		return fmt.Errorf("identische Filter ändern vollständige Suchantwort")
	}

	return nil
}

func (current *world) canonicalFilterWire(data row) error {
	var body any
	if err := json.Unmarshal(current.capture.payload, &body); err != nil {
		return err
	}

	return (row{"Path": "post_filter", "JSON": data["PostFilter"]}).protocolPath(body)
}
