package http

import (
	"context"
	"strconv"
)

func (handler *Handler) diagnostics(ctx context.Context) map[string]string {
	if handler.config.Diagnostics == nil {
		return nil
	}

	snapshot, err := handler.config.Diagnostics.Snapshot(ctx)
	if err != nil {
		return map[string]string{"Status": "Laufzeitdiagnose vorübergehend nicht verfügbar"}
	}

	return snapshot
}

func (view requestView) adminPageNumber() (int, error) {
	input := view.request.URL.Query().Get("page")
	if input == "" {
		return 1, nil
	}

	number, err := strconv.Atoi(input)
	if err != nil {
		return 0, newHTTPError(422, "Ungültige Bestellseite")
	}

	if number < 1 {
		return 0, newHTTPError(422, "Ungültige Bestellseite")
	}

	return number, nil
}
