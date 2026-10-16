package performance_test

import (
	"context"
	"encoding/json"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/client"
	"net/http"
	"time"
)

type world struct {
	originalCaptured bool
	store            *postgres.Store
	temporal         client.Client
	service          application.Service
	guests           []application.GuestView
	order            value.OrderID
	cookie           string
	public           string
	internal         string
	backendReports   []measurement
	fragmentReports  []measurement
	directory        string
	data             map[string]string
	reports          []measurement
	expectedRequests int
	profile          measurementProfile
	failure          string
	diagnostics      []failureProbe
}

type measurement struct {
	measurementProfile
	Route                 string
	Cache                 string
	Requests              int
	P50Milliseconds       float64
	P95Milliseconds       float64
	P99Milliseconds       float64
	ErrorRate             float64
	StatusCounts          map[int]int
	NewConnections        int
	CacheHits             uint64
	CacheMisses           uint64
	Coalesced             uint64
	CacheApplicability    string
	BrowserLCPMeasured    bool
	TargetP95Milliseconds float64
	Goal                  string
	Representation        string
	ComposedHTMLMeasured  bool
}

type sample struct {
	seconds     float64
	status      int
	connections int
}

type counters map[string]uint64

type varnishReport struct {
	Counters map[string]varnishCounter `json:"counters"`
}

type varnishCounter struct {
	Value uint64 `json:"value"`
}

type curlResult struct {
	TimeTotal   float64 `json:"time_total"`
	HTTPCode    int     `json:"http_code"`
	NumConnects int     `json:"num_connects"`
	FileName    string  `json:"filename_effective"`
	URLNum      int     `json:"urlnum"`
	URL         string  `json:"url_effective"`
}

type fixtureState struct {
	Schema      string
	PublicURL   string
	Guests      []application.GuestView
	Order       value.OrderID
	OrderCookie string
	CartCookie  string
}

type measurementProfile struct {
	Boundary                 string
	WarmupPerRoute           int
	Concurrency              int
	MutatingCommandsIncluded bool
}

type probeResponse struct {
	started time.Time
	ended   time.Time
	status  int
	header  http.Header
	body    string
}

type directRequest struct {
	cookie         string
	data           map[string]string
	representation string
}

type failureProbe struct {
	Route          string
	Boundary       string
	Note           string
	Status         int
	BodyBytes      int
	BodySHA256     string
	DegradedHTML   bool
	RemainingESI   bool
	RemainingSSI   bool
	TransportError string
	CacheControl   string
	ContentType    string
	RetryAfter     string
}

type probeRequest struct {
	route    string
	origin   string
	path     string
	cookie   string
	contains string
}

type originalFailure struct {
	HeaderCaptureError   string
	CurlTimeTotalSeconds float64
	Route                string
	Boundary             string
	Source               string
	StartedUTC           time.Time
	EndedUTC             time.Time
	UTCNote              string
	Status               int
	URL                  string
	Host                 string
	URLNum               int
	FileName             string
	Headers              http.Header
	BodyBytes            int
	BodySHA256           string
	Error                string
}

type coldTransfer struct {
	host     string
	url      string
	fileName string
}

type coldBatch struct {
	route     string
	path      string
	cookie    string
	directory string
	count     int
	started   time.Time
	ended     time.Time
	transfers map[int]coldTransfer
}

type failureSnapshot struct {
	Note            string
	StartedUTC      time.Time
	EndedUTC        time.Time
	Alias           snapshotHTTP
	Document        snapshotHTTP
	Product         snapshotProduct
	PostgreSQLError string
	Pool            snapshotPool
}

type snapshotHTTP struct {
	Status int
	Body   json.RawMessage
	Error  string
}

type snapshotProduct struct {
	ID                    value.ProductID
	Revision              uint64
	Published             bool
	Name                  catalog.LocalizedText
	PrimaryCategoryID     value.CategoryID
	AdditionalCategoryIDs []value.CategoryID
	Context               catalog.ComplianceContext
	ApplicableRules       []catalog.RuleReference
	Variants              []snapshotVariant
	Inventory             []snapshotInventory
}

type snapshotVariant struct {
	SKU        value.SKU
	Compliance value.Optional[snapshotValidity]
}

type snapshotValidity struct {
	Context         catalog.ComplianceContext
	ApplicableRules []catalog.RuleReference
}

type snapshotInventory struct {
	SKU              value.SKU
	Revision         uint64
	Price            value.Money
	StockQuantity    int64
	ReservedQuantity int64
}

type snapshotPool struct {
	StartedUTC time.Time
	EndedUTC   time.Time
	Status     int
	Values     map[string]uint64
	Error      string
	Note       string
}

type snapshotRequest struct {
	ctx context.Context
	url string
}
