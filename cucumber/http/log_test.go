package http_test

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (current *world) inspectLogs(table *godog.Table) error {
	path, err := exec.LookPath(binary("SHOP_VARNISH", "varnishd"))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// #nosec G204 -- Lokales Standardbinary aus der expliziten Engineinstallation; keine Shell.
	command := exec.CommandContext(ctx, filepath.Join(filepath.Dir(path), "../bin/varnishlog"), "-n", current.engine.directory+"/varnish", "-d")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Standardenginelog: %w: %s", err, output)
	}

	if !strings.Contains(string(output), "ReqURL") {
		return fmt.Errorf("tatsächliche Requestlogeinträge fehlen")
	}

	return current.forbiddenLogs(string(output), table)
}

func (current *world) forbiddenLogs(output string, table *godog.Table) error {
	for _, fields := range read(table) {
		if strings.Contains(output, fields["Forbidden"]) {
			return fmt.Errorf("Fähigkeit im Standardenginelog: %s", fields["Forbidden"])
		}
	}

	return nil
}
