package demo_test

import (
	"bytes"
	"context"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/adapter/postgres"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/Alps86/shop/internal/runtime"
	"go.temporal.io/sdk/client"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"sync"
	"time"
)

type world struct {
	deployment   *deployment
	clients      map[string]*http.Client
	client       *http.Client
	response     response
	forms        []form
	saved        url.Values
	order        application.OrderView
	orderPath    string
	cartID       string
	runs         map[string]string
	snapshot     string
	prepared     map[string]url.Values
	attempts     []purchaseAttempt
	orders       []value.OrderID
	searchCache  searchCacheProof
	productCache searchCacheEntry
}

type purchaseAttempt struct {
	jar      string
	response response
	err      error
}

type deployment struct {
	app         *runtime.App
	store       *postgres.Store
	connection  client.Client
	config      runtime.Config
	directory   string
	public      string
	backend     string
	reserved    map[string]net.Listener
	external    bool
	cancel      context.CancelFunc
	finished    chan error
	nginx       *exec.Cmd
	varnish     *exec.Cmd
	output      engineLog
	search      *httptest.Server
	unavailable bool
	mutex       sync.RWMutex
}

type response struct {
	requested time.Time
	status    int
	header    http.Header
	body      string
}

type form struct {
	action  string
	values  url.Values
	buttons url.Values
}

type engineLog struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

type searchCacheProof struct {
	host       string
	entries    []searchCacheEntry
	baseline   application.AdminCatalogSnapshot
	seed       runtime.Seed
	product    value.ProductID
	slot       string
	checkpoint uint64
	bans       int64
}

type searchCacheEntry struct {
	path   string
	stored time.Time
	ttl    time.Duration
}

type searchProjectionReply struct {
	Source opensearch.Document `json:"_source"`
}
