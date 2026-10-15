package http

import (
	"errors"
	"github.com/Alps86/shop/internal/domain/value"
	"strings"
)

type HTTPError struct {
	Status  int
	Message string
}

func (err HTTPError) Error() string {
	return err.Message
}

func newHTTPError(status int, message string) error {
	return HTTPError{Status: status, Message: message}
}

func (mapped responseError) status() int {
	err := mapped.cause
	var requestError HTTPError
	if errors.As(err, &requestError) {
		return requestError.Status
	}

	if strings.HasPrefix(value.Code(err), "Invalid") {
		return 422
	}

	statuses := map[string]int{"NotFound": 404, "Removed": 410, "Expired": 410, "Conflict": 409, "PriceChanged": 409, "PriceConfirmationRequired": 409, "Pending": 202, "Unavailable": 409, "AlreadyOrdered": 409, "ComplianceIncomplete": 409, "UnsupportedContext": 422, "SearchUnavailable": 503, "InvalidInput": 422, "InvalidAttribute": 422, "InvalidFilter": 422, "PaymentDeclined": 422, "Declined": 422}
	if status, ok := statuses[value.Code(err)]; ok {
		return status
	}

	return 503
}

func (mapped responseError) message() string {
	err := mapped.cause
	var requestError HTTPError
	if errors.As(err, &requestError) {
		return requestError.Error()
	}

	messages := map[int]string{202: "Änderung angenommen, Status wird aktualisiert", 404: "Nicht gefunden", 410: "Nicht mehr verfügbar", 409: "Daten geändert. Bitte neu laden und bestätigen.", 422: "Eingabe prüfen", 503: "Daten vorübergehend nicht verfügbar"}

	specific := map[string]string{"Removed": "Dieses Produkt wurde dauerhaft entfernt", "Expired": "Warenkorb abgelaufen", "PaymentDeclined": "Simulierte Freigabe abgelehnt"}
	if message, exists := specific[value.Code(err)]; exists {
		return message
	}

	return messages[(responseError{err}).status()]
}
