package http

import (
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func (frame frameReference) path() string {
	return "/_frame/" + frame.Kind + "/" + frame.ID + "/" + strconv.FormatUint(frame.Revision, 10) + "/index.html"
}

func (frame frameReference) valid(request *http.Request) bool {
	if err := value.IdentifierSyntax(frame.ID).Validate(); err != nil {
		return false
	}

	if frame.Revision == 0 {
		return false
	}

	if frame.Kind == "category" {
		if request.URL.RawQuery != "" {
			return false
		}
	}

	return true
}

func (handler *Handler) accelerate(writer http.ResponseWriter, request *http.Request, frame frameReference) bool {
	if !handler.canAccelerate(request, frame) {
		return false
	}

	path := filepath.Join(handler.config.ContentDirectory, frame.path())
	// #nosec G703 -- Frame-ID validiert, Kind intern fest und Revision ausschließlich uint64.
	if _, err := os.Stat(path); err != nil {
		return false
	}

	writer.Header().Set("X-Accel-Redirect", frame.path()+"?"+request.URL.RawQuery)
	writer.Header().Set("Cache-Control", "public,max-age=5")
	writer.Header().Set("X-Robots-Tag", "index,follow")

	return true
}

func (handler *Handler) canAccelerate(request *http.Request, frame frameReference) bool {
	if handler.config.ContentDirectory == "" {
		return false
	}

	if request.Header.Get("HX-Request") == "true" {
		return false
	}

	if (requestView{request}).session() != "" {
		return false
	}

	return frame.valid(request)
}
