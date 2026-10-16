package performance_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (current *world) probe(origin, path, cookie, contains string) error {
	response, err := readResponse(origin, path, cookie)
	if err != nil {
		return err
	}

	return validHTML(response.status, response.body, contains)
}

func readResponse(origin, path, cookie string) (probeResponse, error) {
	request, err := http.NewRequest(http.MethodGet, origin+path, nil)
	if err != nil {
		return probeResponse{}, err
	}

	request.Header.Set("Cookie", cookie)
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return probeResponse{}, err
	}

	body, readError := io.ReadAll(response.Body)

	return probeResponse{status: response.StatusCode, header: response.Header, body: string(body)}, errors.Join(readError, response.Body.Close())
}

func validHTML(status int, body, contains string) error {
	if status != http.StatusOK {
		return fmt.Errorf("Messroute HTTP %d", status)
	}

	for _, marker := range []string{"<esi:", "<!--#", "Daten vorübergehend nicht verfügbar"} {
		if strings.Contains(body, marker) {
			return fmt.Errorf("unvollständige oder degradierte Messseite %s", marker)
		}
	}

	if !strings.Contains(body, contains) {
		return fmt.Errorf("Messseite enthält nicht %s", contains)
	}

	return nil
}
