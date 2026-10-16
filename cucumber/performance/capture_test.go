package performance_test

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func capturesRoute(route string) bool {
	return route == "category" || route == "search" || route == "personalized-search"
}

func allowedHeaders(input http.Header) http.Header {
	result := http.Header{}
	for _, name := range []string{"Content-Type", "Cache-Control", "Retry-After"} {
		for _, value := range input.Values(name) {
			result.Add(name, value)
		}
	}

	return result
}

func (current *world) captureProbe(input probeRequest, response probeResponse, cause error) error {
	endpoint, err := url.Parse(input.origin + input.path)
	if err != nil {
		return err
	}

	proof := originalFailure{Route: input.route, Boundary: current.profile.Boundary, Source: "original-probe", StartedUTC: response.started.UTC(), EndedUTC: response.ended.UTC(), URL: endpoint.String(), Host: endpoint.Host, Status: response.status, Headers: allowedHeaders(response.header), Error: snapshotError(cause)}

	return current.captureOriginal(proof, []byte(response.body))
}

func (current *world) captureOriginal(proof originalFailure, body []byte) error {
	if current.originalCaptured {
		return nil
	}

	if !capturesRoute(proof.Route) {
		return nil
	}

	current.originalCaptured = true

	return current.saveOriginal(proof, body)
}

func (current *world) saveOriginal(proof originalFailure, body []byte) error {
	proof.BodyBytes = len(body)
	proof.BodySHA256 = fmt.Sprintf("%x", sha256.Sum256(body))
	directory := filepath.Join(current.directory, "original-failure")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}

	saved := errors.Join(os.WriteFile(filepath.Join(directory, "body.html"), body, 0600), writeDiagnosticJSON(filepath.Join(directory, "response.json"), proof))

	return errors.Join(saved, current.captureSnapshot(directory))
}

func writeDiagnosticJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0600)
}
