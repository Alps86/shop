package opensearch_test

import (
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http/httptest"
	"time"
)

type row map[string]string

type captureTransport struct {
	payload []byte
}

type world struct {
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
}
