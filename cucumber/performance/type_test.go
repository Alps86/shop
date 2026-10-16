package performance_test

import (
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"go.temporal.io/sdk/client"
	"net/http"
)

type world struct {
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
	status int
	header http.Header
	body   string
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
