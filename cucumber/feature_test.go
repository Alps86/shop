package cucumber

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gherkin "github.com/cucumber/gherkin/go/v26"
	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages/go/v21"
)

func TestFeatureSyntax(t *testing.T) {
	if err := (suite{}).parse(); err != nil {
		t.Fatal(err)
	}
}

func TestFeatures(t *testing.T) {
	current := suite{}
	if err := current.run(); err != nil {
		t.Fatal(err)
	}
}

func (current suite) parse() error {
	return filepath.WalkDir("../spec", current.parseEntry)
}

func (current suite) parseEntry(path string, _ os.DirEntry, err error) error {
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if filepath.Ext(path) != ".feature" {
		return nil
	}

	return current.parseFile(path)
}

func (current suite) parseFile(path string) error {
	data, err := fs.ReadFile(os.DirFS("../spec"), strings.TrimPrefix(path, "../spec/"))
	if err != nil {
		return err
	}

	identifier := &messages.Incrementing{}
	_, err = gherkin.ParseGherkinDocument(bytes.NewReader(data), identifier.NewId)

	return err
}

func (current suite) run() error {
	options := &godog.Options{
		Format: "pretty", Paths: []string{"../spec/features"},
		Tags: "~@spec-only", Strict: true,
	}
	status := (godog.TestSuite{Name: "shop", Options: options}).Run()
	if status != 0 {
		return fmt.Errorf("cucumber exit status: %d", status)
	}

	return nil
}
