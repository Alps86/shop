package opensearch_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (transport *captureTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if strings.HasSuffix(request.URL.Path, "search") {
		payload, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}

		request.Body = io.NopCloser(bytes.NewReader(payload))
		transport.payload = payload
	}

	return transport.forward(request)
}

func (current *world) queryReplaced() bool {
	if current.requestedText == "" {
		return false
	}

	payload, err := json.Marshal(current.requestedText)
	if err != nil {
		return true
	}

	needle := append([]byte(`"query":`), payload...)

	return !bytes.Contains(current.capture.payload, needle)
}

func (expected row) maximum(count int) error {
	limit, err := strconv.Atoi(expected["MaximumCount"])
	if err != nil {
		return err
	}

	if count > limit {
		return fmt.Errorf("%d Vorschläge überschreiten Maximum %d", count, limit)
	}

	return nil
}

func (transport *captureTransport) forward(request *http.Request) (*http.Response, error) {
	if !transport.observeResponse {
		return http.DefaultTransport.RoundTrip(request)
	}

	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	payload, readError := io.ReadAll(response.Body)
	closeError := response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(payload))
	if strings.HasSuffix(request.URL.Path, "search") {
		transport.response = payload
	}

	return response, errors.Join(readError, closeError)
}
