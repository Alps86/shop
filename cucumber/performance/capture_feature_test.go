package performance_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCapture(t *testing.T) {
	t.Setenv("SHOP_PERFORMANCE_OPENSEARCH_URL", "")
	t.Setenv("SHOP_PERFORMANCE_SEARCH_INDEX", "")
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @performance-capture", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "performance-capture", Options: options, ScenarioInitializer: registerCapture}
	if suite.Run() != 0 {
		t.Fatal("Technische Zuordnung der ursprünglichen Messantwort fehlgeschlagen")
	}
}

func registerCapture(ctx *godog.ScenarioContext) {
	ctx.Step("^kalte curl-Antworten unabhängig von ihrer Abschlussreihenfolge zugeordnet werden$", verifyCaptureMapping)
	ctx.Step("^ursprüngliche kalte Fehlerantworten vor dem Cleanup dauerhaft erhalten bleiben$", verifyOriginalBodies)
	ctx.Step("^Diagnoseheader ausschließlich erlaubte öffentliche Werte enthalten$", verifyCaptureHeaders)
}

func verifyCaptureMapping(table *godog.Table) error {
	batch := coldBatch{route: "category", path: "/de-DE/kategorie/test", directory: "/tmp/capture-contract", count: 2}
	batch.config("http://127.0.0.1:8080")
	for _, row := range read(table) {
		if err := verifyCaptureRow(batch, row); err != nil {
			return err
		}
	}

	return nil
}

func verifyCaptureRow(batch coldBatch, row map[string]string) error {
	var observed curlResult
	if err := json.Unmarshal([]byte(row["JSON"]), &observed); err != nil {
		return err
	}

	transfer, err := batch.transfer(observed)
	if err != nil {
		return err
	}

	if filepath.Base(transfer.fileName) != row["File"] {
		return fmt.Errorf("falsche Antwortdatei für urlnum %d", observed.URLNum)
	}

	return verifyCaptureStatus(observed, row["Error"])
}

func verifyCaptureStatus(observed curlResult, expected string) error {
	err := validHTML(observed.HTTPCode, "<html>marker</html>", "marker")
	actual := fmt.Sprint(err)
	if actual != expected {
		return fmt.Errorf("Statusprüfung: %s statt %s", actual, expected)
	}

	return nil
}

func verifyCaptureHeaders(table *godog.Table) error {
	input := http.Header{}
	for _, row := range read(table) {
		input.Set(row["Name"], row["Value"])
	}

	safe := allowedHeaders(input)
	if len(safe) != 3 {
		return fmt.Errorf("unerwartete Diagnoseheader: %v", safe)
	}

	for _, name := range []string{"Cookie", "Authorization", "Set-Cookie"} {
		if safe.Get(name) != "" {
			return fmt.Errorf("privater Header wurde übernommen: %s", name)
		}
	}

	return nil
}

func verifyOriginalBodies(table *godog.Table) error {
	for _, row := range read(table) {
		if err := verifyOriginalBody(row); err != nil {
			return err
		}
	}

	return nil
}

func verifyOriginalBody(row map[string]string) (err error) {
	directory, err := os.MkdirTemp("", "capture-contract-")
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, os.RemoveAll(directory))
	}()
	current := &world{directory: directory}
	batch := coldBatch{route: "category", directory: filepath.Join(directory, "temporary"), count: 1}
	batch.config("http://127.0.0.1:8080")
	if err := os.MkdirAll(batch.directory, 0700); err != nil {
		return err
	}

	return verifyOriginalFiles(current, batch, row)
}

func verifyOriginalFiles(current *world, batch coldBatch, row map[string]string) error {
	transfer := batch.transfers[0]
	if err := os.WriteFile(transfer.fileName, []byte(row["Body"]), 0600); err != nil {
		return err
	}

	status, err := strconv.Atoi(row["Status"])
	if err != nil {
		return err
	}

	observed := curlResult{URLNum: 0, FileName: transfer.fileName, URL: transfer.url, HTTPCode: status}
	if current.checkColdTransfer(batch, observed, "marker") == nil {
		return errors.New("ursprünglicher Messfehler wurde verborgen")
	}

	return verifySavedBody(current, batch, row["Body"])
}

func verifySavedBody(current *world, batch coldBatch, expected string) error {
	if err := os.RemoveAll(batch.directory); err != nil {
		return err
	}

	if err := current.captureOriginal(originalFailure{Route: "category"}, []byte("spätere Antwort")); err != nil {
		return err
	}

	body, err := readBody(filepath.Join(current.directory, "original-failure"), "body.html")
	if err != nil {
		return err
	}

	if string(body) != expected {
		return errors.New("ursprünglicher Body fehlt nach Cleanup oder wurde überschrieben")
	}

	return verifySavedMetadata(current, expected)
}

func verifySavedMetadata(current *world, expected string) error {
	payload, err := readBody(filepath.Join(current.directory, "original-failure"), "response.json")
	if err != nil {
		return err
	}

	var proof originalFailure
	if err := json.Unmarshal(payload, &proof); err != nil {
		return err
	}

	return verifyOriginalMetadata(proof, expected)
}

func verifyOriginalMetadata(proof originalFailure, expected string) error {
	if proof.BodyBytes != len(expected) {
		return errors.New("Originalmetadaten passen nicht zum erhaltenen Body")
	}

	if proof.HeaderCaptureError == "" {
		return errors.New("fehlende Headerdatei ist nicht ausdrücklich als unbelegt erfasst")
	}

	return nil
}
