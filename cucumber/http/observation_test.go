package http_test

import (
	"github.com/Alps86/shop/internal/application"
	"slices"
)

func (current *world) recordSearchQuery(query application.SearchQuery) {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	current.lastQuery = query
}

func (current *world) snapshotSearch() searchObservation {
	current.mutex.Lock()
	defer current.mutex.Unlock()

	return searchObservation{query: current.lastQuery, calls: slices.Clone(current.calls)}
}
