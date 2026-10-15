package projection_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cucumber/godog"
	"net/http"
	"strconv"
	"strings"
)

func (current *world) aliasObserved(table *godog.Table) error {
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, current.Endpoint+"/_alias/"+current.Alias, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	aliases := map[string]json.RawMessage{}
	decodeErr := json.NewDecoder(response.Body).Decode(&aliases)
	if err := errors.Join(decodeErr, response.Body.Close()); err != nil {
		return err
	}

	return current.aliasNames(table, aliases)
}

func (current *world) aliasNames(table *godog.Table, aliases map[string]json.RawMessage) error {
	actual := map[string]string{"IndexCount": strconv.Itoa(len(aliases))}
	for name := range aliases {
		actual["IndexSuffix"] = strings.TrimPrefix(name, current.Alias+"-")
	}

	return current.matches(actual, current.rows(table)[0])
}
