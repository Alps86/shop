package http_test

import (
	"fmt"
	"github.com/cucumber/godog"
)

func (current *world) resetSearchCalls(table *godog.Table) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	current.calls = nil
	current.fault = read(table)[0]["Fault"]

	return nil
}

func (current *world) searchCalls(table *godog.Table) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	count := 0
	for _, call := range current.calls {
		if call == "Search" {
			count++
		}
	}

	return read(table).matches(map[string]string{"SearchCalls": fmt.Sprint(count)})
}
