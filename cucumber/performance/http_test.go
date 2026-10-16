package performance_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (current *world) probe(input probeRequest) error {
	response, err := readResponse(input.origin, input.path, input.cookie)
	if err == nil {
		err = validHTML(response.status, response.body, input.contains)
	}

	if err != nil {
		return errors.Join(err, current.captureProbe(input, response, err))
	}

	return nil
}

func readResponse(origin, path, cookie string) (probeResponse, error) {
	request, err := http.NewRequest(http.MethodGet, origin+path, nil)
	if err != nil {
		return probeResponse{}, err
	}

	request.Header.Set("Cookie", cookie)
	started := time.Now()
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return probeResponse{started: started, ended: time.Now()}, err
	}

	body, readError := io.ReadAll(response.Body)

	closeError := response.Body.Close()

	return probeResponse{status: response.StatusCode, header: response.Header, body: string(body), started: started, ended: time.Now()}, errors.Join(readError, closeError)
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
