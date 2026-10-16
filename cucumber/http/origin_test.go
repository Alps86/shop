package http_test

import (
	"net/http"
	"strings"
)

func (origin compositionOrigin) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if origin.dropFragment(request) {
		origin.closeConnection(writer)
		return
	}

	origin.handler.ServeHTTP(writer, request)
}

func (origin compositionOrigin) dropFragment(request *http.Request) bool {
	if !origin.current.fragmentFault.Load() {
		return false
	}

	return strings.HasPrefix(request.URL.Path, "/fragments/")
}

func (origin compositionOrigin) closeConnection(writer http.ResponseWriter) {
	connection, _, err := writer.(http.Hijacker).Hijack()
	if err != nil {
		return
	}

	_ = connection.Close()
}
