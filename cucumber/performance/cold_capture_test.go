package performance_test

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (batch *coldBatch) config(origin string) string {
	var config strings.Builder
	batch.transfers = map[int]coldTransfer{}
	run := strings.ToLower(rand.Text())
	for index := range batch.count {
		transfer := coldTransfer{host: "cold-" + run + "-" + strconv.Itoa(index) + ".localhost", url: origin + batch.path, fileName: filepath.Join(batch.directory, strconv.Itoa(index))}
		batch.transfers[index] = transfer
		batch.writeTransfer(&config, transfer)
		if index+1 < batch.count {
			config.WriteString("next\n")
		}
	}

	return config.String()
}

func (batch coldBatch) writeTransfer(config *strings.Builder, transfer coldTransfer) {
	fmt.Fprintf(config, "url = %q\nheader = %q\noutput = %q\nwrite-out = \"%%{json}\\n\"\nmax-time = 15\n", transfer.url, "Host: "+transfer.host, transfer.fileName)
	if capturesRoute(batch.route) {
		fmt.Fprintf(config, "dump-header = %q\n", transfer.fileName+".headers")
	}

	if batch.cookie != "" {
		fmt.Fprintf(config, "header = %q\n", "Cookie: "+batch.cookie)
	}
}

func (batch coldBatch) transfer(observed curlResult) (coldTransfer, error) {
	transfer, exists := batch.transfers[observed.URLNum]
	if !exists {
		return coldTransfer{}, errors.New("curl urlnum gehört nicht zur erzeugten Konfiguration")
	}

	if observed.FileName != transfer.fileName {
		return coldTransfer{}, errors.New("curl filename_effective passt nicht zu urlnum")
	}

	if observed.URL != transfer.url {
		return coldTransfer{}, errors.New("curl url_effective passt nicht zur erzeugten URL")
	}

	return transfer, nil
}

func (current *world) checkColdTransfer(batch coldBatch, observed curlResult, contains string) error {
	transfer, err := batch.transfer(observed)
	if err != nil {
		return err
	}

	body, readError := readBody(batch.directory, filepath.Base(transfer.fileName))
	cause := errors.Join(readError, validHTML(observed.HTTPCode, string(body), contains))
	if cause == nil {
		return nil
	}

	proof := originalFailure{Route: batch.route, Boundary: current.profile.Boundary, Source: "original-curl", StartedUTC: batch.started.UTC(), EndedUTC: batch.ended.UTC(), UTCNote: "UTC-Grenzen des gesamten parallelen curl-Aufrufs, keine Einzelrequestspanne", Status: observed.HTTPCode, URL: transfer.url, Host: transfer.host, URLNum: observed.URLNum, FileName: filepath.Base(transfer.fileName), CurlTimeTotalSeconds: observed.TimeTotal, Error: cause.Error()}

	proof.captureHeaders(transfer.fileName + ".headers")

	return errors.Join(cause, current.captureOriginal(proof, body))
}

func (proof *originalFailure) captureHeaders(path string) {
	var err error
	proof.Headers, err = coldHeaders(path)
	if err != nil {
		proof.HeaderCaptureError = snapshotError(err)
	}
}

func coldHeaders(path string) (http.Header, error) {
	file, err := os.OpenInRoot(filepath.Dir(path), filepath.Base(path))
	if err != nil {
		return nil, err
	}

	response, readError := http.ReadResponse(bufio.NewReader(file), nil)
	closeError := file.Close()
	if err := errors.Join(readError, closeError); err != nil {
		return nil, err
	}

	return allowedHeaders(response.Header), nil
}

func (current *world) saveColdMetrics(route string, output []byte) error {
	var sanitized strings.Builder
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var entry map[string]json.RawMessage
		if err := decoder.Decode(&entry); err != nil {
			return err
		}

		payload, err := coldMetricJSON(entry)
		if err != nil {
			return err
		}

		sanitized.Write(payload)
		sanitized.WriteByte('\n')
	}

	return os.WriteFile(current.reportPath(route, "cold", ".jsonl"), []byte(sanitized.String()), 0600)
}

func coldMetricJSON(entry map[string]json.RawMessage) ([]byte, error) {
	safe := map[string]json.RawMessage{}
	for _, key := range []string{"time_total", "http_code", "num_connects", "urlnum", "filename_effective", "exitcode", "time_starttransfer", "time_connect", "time_appconnect", "time_namelookup", "time_pretransfer", "size_download", "speed_download"} {
		if value, exists := entry[key]; exists {
			safe[key] = value
		}
	}

	return json.Marshal(safe)
}
