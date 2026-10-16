package opensearch_test

import (
	"encoding/json"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/http/httptest"
	"time"
)

type row map[string]string

type delayedProxy struct {
	next  http.Handler
	delay time.Duration
}

type captureTransport struct {
	observeResponse bool
	payload         []byte
	response        []byte
}

type cacheStatistics struct {
	All struct {
		Primaries struct {
			RequestCache struct {
				Hits int64 `json:"hit_count"`
			} `json:"request_cache"`
		} `json:"primaries"`
	} `json:"_all"`
}

type world struct {
	cacheBefore   int64
	capture       *captureTransport
	requestedText string
	adapter       opensearch.Adapter
	config        opensearch.Config
	result        application.SearchResult
	query         application.SearchQuery
	interests     []sale.CategoryInterest
	products      map[value.ProductID]catalog.Product
	inventory     map[value.ProductID][]sale.Inventory
	template      catalog.Product
	now           time.Time
	err           error
	writerErr     error
	sourceFault   string
	sourceCode    string
	abortCount    int
	snapshot      event.ReindexSnapshot
	lateEvents    []application.ProjectionEvent
	faultServer   *httptest.Server
	cleanupConfig opensearch.Config
	elapsed       time.Duration
}

type sourceResponse struct {
	Hits struct {
		Hits []map[string]json.RawMessage `json:"hits"`
	} `json:"hits"`
}
