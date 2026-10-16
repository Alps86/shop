package opensearch_test

import (
	"bytes"
	"encoding/json"
	"github.com/Alps86/shop/internal/adapter/event"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
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
	transportLog   *transportLog
	previousLogger *slog.Logger
	codec          *cardCodecProbe
	msearch        *msearchProbe
	pool           *poolProbe
	cacheBefore    int64
	capture        *captureTransport
	requestedText  string
	adapter        opensearch.Adapter
	config         opensearch.Config
	result         application.SearchResult
	query          application.SearchQuery
	interests      []sale.CategoryInterest
	products       map[value.ProductID]catalog.Product
	inventory      map[value.ProductID][]sale.Inventory
	template       catalog.Product
	now            time.Time
	err            error
	writerErr      error
	sourceFault    string
	sourceCode     string
	abortCount     int
	snapshot       event.ReindexSnapshot
	lateEvents     []application.ProjectionEvent
	faultServer    *httptest.Server
	cleanupConfig  opensearch.Config
	elapsed        time.Duration
}

type sourceResponse struct {
	Hits struct {
		Hits []map[string]json.RawMessage `json:"hits"`
	} `json:"hits"`
}

type poolProbe struct {
	original  http.RoundTripper
	transport *http.Transport
	server    *httptest.Server
	sentinel  *poolSentinel
	barrier   *poolBarrier
	payload   []byte
	expected  application.SearchResult
	results   []application.SearchResult
	failures  []error
	reused    atomic.Int64
	parallel  int
}

type poolBarrier struct {
	arrived  atomic.Int64
	release  chan struct{}
	parallel int
}

type poolSentinel struct {
	next  http.RoundTripper
	calls atomic.Int64
}

type poolBatch struct {
	probe   *poolProbe
	adapter opensearch.Adapter
	query   application.SearchQuery
	wait    sync.WaitGroup
}

type msearchProbe struct {
	mutex         sync.Mutex
	server        *httptest.Server
	response      []byte
	request       []byte
	path          string
	mediaType     string
	calls         int
	status        int
	delay         time.Duration
	phase         string
	prefix        []byte
	firstPartSent bool
}

type sourceEnvelope struct {
	Responses []sourceResponse `json:"responses"`
}

type protocolEnvelope struct {
	Responses []json.RawMessage `json:"responses"`
}

type cardCodecProbe struct {
	original application.ProductCard
	document cardCodecDocument
}

type cardCodecDocument struct {
	Variants []cardCodecVariant `json:"variants"`
}

type cardCodecVariant struct {
	Card        application.ProductCard `json:"card"`
	CardPayload []byte                  `json:"card_payload"`
}

type transportLog struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}
