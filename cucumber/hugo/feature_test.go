package hugo_test

import (
	"errors"
	"fmt"
	"github.com/Alps86/shop/internal/adapter/hugo"
	"github.com/cucumber/godog"
	"io"
	"os"
	"strconv"
	"testing"
)

func TestFeatures(t *testing.T) {
	current := &world{}
	options := &godog.Options{Format: "pretty", Paths: []string{"../../spec/features"}, Tags: "~@spec-only && @hugo-io", Strict: true, TestingT: t}
	suite := godog.TestSuite{Name: "hugo-io", Options: options, ScenarioInitializer: current.register}
	if suite.Run() != 0 {
		t.Fatal("Hugo-Ausgabevertrag fehlgeschlagen")
	}
}

func (current *world) register(ctx *godog.ScenarioContext) {
	ctx.Step(`^eine öffentliche Hugo-Seite vorbereitet wird$`, current.prepare)
	ctx.Step(`^die Hugo-Seite an einen abbrechenden Ausgabestrom geschrieben wird$`, current.write)
	ctx.Step(`^der Schreibfehler erhalten bleibt$`, current.observed)
}

func (current *world) prepare(table *godog.Table) {
	data := read(table)[0]
	current.page = hugo.Page{Title: data["title"], Description: data["description"], URL: data["url"], Layout: data["layout"], FrameKind: data["frame_kind"], Canonical: data["canonical"], Fragment: data["fragment"]}
}

func (current *world) write(table *godog.Table) error {
	data := read(table)[0]
	if data["Writer"] == "closed-file" {
		return current.writeClosedFile()
	}

	count, err := strconv.Atoi(data["SuccessfulWrites"])
	if err != nil {
		return err
	}

	writer := &interruptedWriter{successful: count}
	current.err = current.page.Write(writer)

	return nil
}

func (writer *interruptedWriter) Write(data []byte) (int, error) {
	if writer.writes == writer.successful {
		return 0, io.ErrClosedPipe
	}

	writer.writes++

	return len(data), nil
}

func (current *world) writeClosedFile() error {
	file, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	current.err = current.page.Write(file)

	return nil
}

func (current *world) observed(table *godog.Table) error {
	expected, exists := map[string]error{"io.ErrClosedPipe": io.ErrClosedPipe, "os.ErrClosed": os.ErrClosed}[read(table)[0]["Error"]]
	if !exists {
		return fmt.Errorf("unbekannter erwarteter IO-Fehler")
	}

	if !errors.Is(current.err, expected) {
		return fmt.Errorf("Schreibfehler nicht erhalten: %w", current.err)
	}

	return nil
}

func read(table *godog.Table) []map[string]string {
	result := []map[string]string{}
	for _, record := range table.Rows[1:] {
		data := map[string]string{}
		for index, cell := range record.Cells {
			data[table.Rows[0].Cells[index].Value] = cell.Value
		}

		result = append(result, data)
	}

	return result
}
