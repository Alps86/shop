package opensearch

import (
	"net/http"
	"time"
)

func newHTTPClient() *http.Client {
	transport := http.DefaultTransport
	standard, ok := transport.(*http.Transport)
	if ok {
		clone := standard.Clone()
		clone.MaxIdleConnsPerHost = 10
		transport = clone
	}

	return &http.Client{Transport: transport, Timeout: 5 * time.Second}
}
