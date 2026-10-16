package performance_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (current *world) probe(path, cookie, contains string) error {
	request, err := http.NewRequest(http.MethodGet, current.public+path, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Cookie", cookie)
	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return err
	}

	body, readError := io.ReadAll(response.Body)

	return errors.Join(readError, response.Body.Close(), validHTML(response.StatusCode, string(body), contains))
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
