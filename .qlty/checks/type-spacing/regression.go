package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (runner runner) regression(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return err
	}

	return runner.verifyFixtures(fixtures)
}

func (runner runner) verifyFixtures(fixtures []fixture) (result error) {
	directory, err := os.MkdirTemp("", "type-spacing-regression-")
	if err != nil {
		return err
	}

	defer runner.cleanup(directory, &result)
	if err := runner.verifyAll(fixtures, filepath.Join(directory, "type.go")); err != nil {
		return err
	}

	fmt.Printf("%d Deklarationsabstands-Regressionen erfolgreich.\n", len(fixtures))
	return nil
}

func (fixture fixture) verify(path string) error {
	if err := os.WriteFile(path, []byte(fixture.Source), 0600); err != nil {
		return err
	}

	binary, err := os.Executable()
	if err != nil {
		return err
	}

	result, err := exec.Command(binary, path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w: %s", fixture.Name, err, result)
	}

	return fixture.assertOutput(string(result))
}

func (fixture fixture) assertOutput(result string) error {
	found := strings.Count(result, "shop-go-type-declaration-spacing")
	if found != fixture.Violations {
		return fmt.Errorf("%s: %d Befunde, erwartet %d", fixture.Name, found, fixture.Violations)
	}

	return nil
}

func (runner runner) verifyAll(fixtures []fixture, path string) error {
	var failures []error
	for _, fixture := range fixtures {
		failures = append(failures, fixture.verify(path))
	}

	return errors.Join(failures...)
}

func (runner runner) cleanup(directory string, result *error) {
	*result = errors.Join(*result, os.RemoveAll(directory))
}
