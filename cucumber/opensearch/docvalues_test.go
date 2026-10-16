package opensearch_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

func (current *world) mappingObserved(table *godog.Table) error {
	response, err := http.Get(current.config.Endpoint + "/" + current.config.Index + "/_mapping")
	if err != nil {
		return err
	}

	var exported map[string]any
	decodeError := json.NewDecoder(response.Body).Decode(&exported)
	if err := errors.Join(decodeError, response.Body.Close()); err != nil {
		return err
	}

	if len(exported) != 1 {
		return fmt.Errorf("genau eine echte Mappinggeneration erwartet")
	}

	return mappingFields(slices.Collect(maps.Values(exported))[0], table)
}

func mappingFields(mapping any, table *godog.Table) error {
	for _, expected := range read(table) {
		if err := expected.protocolPath(mapping); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) nestedDocValues(table *godog.Table) error {
	current.msearch.mutex.Lock()
	payload := append([]byte{}, current.msearch.request...)
	current.msearch.mutex.Unlock()
	lines := strings.Split(strings.TrimSpace(string(payload)), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("Hauptabfrage fehlt")
	}

	var body docValuesQuery
	if err := json.Unmarshal([]byte(lines[1]), &body); err != nil {
		return err
	}

	if len(body.Query.Bool.Should) != 1 {
		return fmt.Errorf("genau eine gewählte Variantenabfrage erwartet")
	}

	return compareDocValues(body.Query.Bool.Should[0].Nested.InnerHits, read(table)[0])
}

func compareDocValues(inner map[string]any, expected row) error {
	for key, source := range map[string]string{"_source": "Source", "stored_fields": "StoredFields", "docvalue_fields": "DocValueFields"} {
		var wanted any
		if err := json.Unmarshal([]byte(expected[source]), &wanted); err != nil {
			return err
		}

		if !reflect.DeepEqual(inner[key], wanted) {
			return fmt.Errorf("gewählter Varianttreffer: %s weicht ab: %v", key, inner[key])
		}
	}

	return nil
}
