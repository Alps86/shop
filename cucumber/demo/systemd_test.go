package demo_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (deployment *deployment) restartSystemd() error {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	if err := systemdAction(ctx, "stop", "shop-ref-demo-shop.service"); err != nil {
		return err
	}

	if err := systemdAction(ctx, "restart", "shop-ref-demo-postgres.service", "shop-ref-demo-temporal.service", "shop-ref-demo-opensearch.service"); err != nil {
		return err
	}

	if err := deployment.waitSystemd(ctx, true); err != nil {
		return err
	}

	if err := systemdAction(ctx, "start", "shop-ref-demo-shop.service"); err != nil {
		return err
	}

	return deployment.ready(30 * time.Second)
}

func (deployment *deployment) systemdSearch(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	action := "stop"
	if state == "available" {
		action = "start"
	}

	if err := systemdAction(ctx, action, "shop-ref-demo-opensearch.service"); err != nil {
		return err
	}

	if state == "available" {
		return deployment.waitSystemd(ctx, false)
	}

	return nil
}

func systemdAction(ctx context.Context, arguments ...string) error {
	// #nosec G204 -- Nur feste eigene Referenzunits und feste Aktionen, keine Shell.
	command := exec.CommandContext(ctx, "systemctl", append([]string{"--user"}, arguments...)...)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("Referenzdienstaktion %v: %w: %s", arguments, err, output)
	}

	return nil
}

func systemdCacheCounter(name string) (int64, error) {
	program := os.Getenv("SHOP_SYSTEMD_VARNISHSTAT")
	if program == "" {
		return 0, fmt.Errorf("SHOP_SYSTEMD_VARNISHSTAT für Referenzzähler fehlt")
	}

	// #nosec G204 G702 -- Bewusst konfigurierte eigene Werkzeugdatei; feste Zählerargumente, keine Shell oder HTTPeingaben.
	output, err := exec.Command(program, "-1", "-f", name).CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("Referenzcachezähler: %w", err)
	}

	return parseCounter(string(output), name)
}
