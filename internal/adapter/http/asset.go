package http

import (
	"embed"
	"net/http"
	"regexp"
	"strings"
)

//go:embed asset/*
var assetFiles embed.FS

var assetFingerprint = regexp.MustCompile(`\.[a-f0-9]{12}\.`)

func (handler *Handler) asset(writer http.ResponseWriter, request *http.Request) {
	path := "asset/" + strings.TrimPrefix(request.URL.Path, "/assets/")
	content, err := assetFiles.ReadFile(path)
	if err != nil {
		handler.failure(writer, request, newHTTPError(404, "Asset nicht gefunden"))
		return
	}

	handler.assetHeaders(writer, path, content)
	// #nosec G705 -- Ausschließlich unveränderliche eingebettete Assets, keine Benutzerausgabe.
	_, _ = writer.Write(content)
}

func (handler *Handler) assetHeaders(writer http.ResponseWriter, path string, content []byte) {
	writer.Header().Set("Cache-Control", "public,max-age=3600")
	if assetFingerprint.MatchString(path) {
		writer.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
	}

	writer.Header().Set("Content-Type", http.DetectContentType(content))
	if strings.HasSuffix(path, ".css") {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
	}

	if strings.HasSuffix(path, ".js") {
		writer.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	}

	if strings.HasSuffix(path, ".svg") {
		writer.Header().Set("Content-Type", "image/svg+xml")
	}
}
