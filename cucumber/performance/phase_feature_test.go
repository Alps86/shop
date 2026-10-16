package performance_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"time"
)

func verifyPoolPhases(table *godog.Table) error {
	for _, row := range read(table) {
		if err := verifyPoolPhase(row); err != nil {
			return err
		}
	}

	return nil
}

func verifyPoolPhase(row map[string]string) (err error) {
	fixture, err := newPhaseFixture(row)
	if err != nil {
		return err
	}

	server := httptest.NewServer(http.HandlerFunc(fixture.serve))
	defer server.Close()
	previous := os.Getenv("SHOP_PERFORMANCE_INTERNAL_URL")
	defer func() {
		err = errors.Join(err, os.Setenv("SHOP_PERFORMANCE_INTERNAL_URL", previous))
	}()
	if err := os.Setenv("SHOP_PERFORMANCE_INTERNAL_URL", server.URL); err != nil {
		return err
	}

	return fixture.observe(row)
}

func newPhaseFixture(row map[string]string) (*phaseFixture, error) {
	beforeStatus, err := strconv.Atoi(row["BeforeStatus"])
	if err != nil {
		return nil, err
	}

	afterStatus, err := strconv.Atoi(row["AfterStatus"])
	if err != nil {
		return nil, err
	}

	return &phaseFixture{before: row["Before"], after: row["After"], beforeStatus: beforeStatus, afterStatus: afterStatus}, nil
}

func (fixture *phaseFixture) serve(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/ready" {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	call := fixture.calls.Add(1)
	status, body := fixture.beforeStatus, fixture.before
	if call > 1 {
		status, body = fixture.afterStatus, fixture.after
	}

	writer.WriteHeader(status)
	if _, err := fmt.Fprint(writer, body); err != nil {
		fixture.writeFailed.Store(true)
	}
}

func (fixture *phaseFixture) observe(row map[string]string) error {
	if err := fixture.prepare(row); err != nil {
		return err
	}

	var cause error
	if row["ActionError"] != "" {
		cause = errors.New(row["ActionError"])
	}

	output, observed := fixture.current.phaseOutput(poolPhaseInput{route: row["Route"], cache: row["Cache"]}, func() ([]byte, error) {
		return []byte(row["Payload"]), cause
	})

	return fixture.verify(row, string(output), errors.Is(observed, cause))
}

func (fixture *phaseFixture) prepare(row map[string]string) error {
	if err := fixture.current.initializeInternal(); err != nil {
		return err
	}

	fixture.current.profile.Boundary = row["Boundary"]

	return os.Unsetenv("SHOP_PERFORMANCE_INTERNAL_URL")
}

func (fixture *phaseFixture) verify(row map[string]string, output string, sameError bool) error {
	if !sameError {
		return errors.New("ursprünglicher Lastfehler wurde verändert")
	}

	if output != row["Payload"] {
		return errors.New("ursprüngliche Werkzeugausgabe wurde verändert")
	}

	calls, err := strconv.ParseInt(row["Calls"], 10, 64)
	if err != nil {
		return err
	}

	if fixture.calls.Load() != calls {
		return fmt.Errorf("Poolaufrufe %d statt %d", fixture.calls.Load(), calls)
	}

	return fixture.verifyReport(row)
}

func (fixture *phaseFixture) verifyReport(row map[string]string) error {
	if fixture.writeFailed.Load() {
		return errors.New("technischer Origin konnte Diagnose nicht übertragen")
	}

	if row["Calls"] == "0" {
		return fixture.verifyNoPhase()
	}

	return fixture.verifyPhase(row)
}

func (fixture *phaseFixture) verifyNoPhase() error {
	if len(fixture.current.poolPhases) != 0 {
		return errors.New("nicht ausgewählte Phase wurde trotzdem berichtet")
	}

	return nil
}

func (fixture *phaseFixture) verifyPhase(row map[string]string) error {
	if len(fixture.current.poolPhases) != 1 {
		return errors.New("erreichte Phase fehlt oder wurde mehrfach berichtet")
	}

	phase := fixture.current.poolPhases[0]
	if phase.Comparison != row["Comparison"] {
		return fmt.Errorf("Vergleichbarkeit %s statt %s", phase.Comparison, row["Comparison"])
	}

	return verifyPoolDeltas(phase, row)
}

func verifyPoolDeltas(phase poolPhase, row map[string]string) error {
	var expected map[string]uint64
	if err := json.Unmarshal([]byte(row["Deltas"]), &expected); err != nil {
		return err
	}

	if !maps.Equal(phase.ConditionalDeltas, expected) {
		return fmt.Errorf("bedingte Differenzen %v statt %v", phase.ConditionalDeltas, expected)
	}

	if phase.LifetimeEvidence != "unproven" {
		return errors.New("Poollebensdauer wurde ohne Nachweis behauptet")
	}

	return verifyPoolBoundaries(phase)
}

func verifyPoolBoundaries(phase poolPhase) error {
	for _, instant := range []time.Time{phase.Start.StartedUTC, phase.End.EndedUTC} {
		if instant.IsZero() {
			return errors.New("UTC-Grenzen der Poolbeobachtung fehlen")
		}
	}

	if phase.End.StartedUTC.Before(phase.Start.EndedUTC) {
		return errors.New("Poolbeobachtungen überlappen")
	}

	return nil
}

func verifyEarlyInternalOrigin(table *godog.Table) error {
	for _, row := range read(table) {
		current := &world{internal: row["Origin"]}
		err := current.prepareMeasurements()
		if fmt.Sprint(err) != row["Error"] {
			return fmt.Errorf("frühe Originprüfung: %w statt %s", err, row["Error"])
		}
	}

	return nil
}
