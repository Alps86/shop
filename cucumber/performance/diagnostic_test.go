package performance_test

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

func (current *world) diagnoseFailure(route, boundary string) {
	origin, path, cookie := current.diagnosticRequest(route, boundary)
	response, err := readResponse(origin, path, cookie)
	observed := failureProbe{Route: route, Boundary: boundary, Note: "Separate fresh probe after failure; not the original hey/curl failed response", Status: response.status, BodyBytes: len(response.body), BodySHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(response.body)))}
	if err != nil {
		observed.TransportError = err.Error()
	}

	observed.DegradedHTML = strings.Contains(response.body, "Daten vorübergehend nicht verfügbar")
	observed.RemainingESI = strings.Contains(response.body, "<esi:")
	observed.RemainingSSI = strings.Contains(response.body, "<!--#")
	observed.CacheControl = response.header.Get("Cache-Control")
	observed.ContentType = response.header.Get("Content-Type")
	observed.RetryAfter = response.header.Get("Retry-After")
	current.diagnostics = append(current.diagnostics, observed)
}

func (current *world) diagnosticRequest(route, boundary string) (string, string, string) {
	path, cookie, _ := current.route(route)
	origin := current.public
	if boundary != "requestStart-to-last-composed-HTML-byte" {
		origin = current.internal
	}

	if boundary == "requestStart-to-last-Go-fragment-byte" {
		path = "/fragments" + path
		cookie = ""
	}

	return origin, path, cookie
}
