package demo_test

import (
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (current *world) cacheHit(table *godog.Table) error {
	data := read(table)[0]
	if err := current.rememberProductCache(data); err != nil {
		return err
	}

	before, err := current.deployment.cacheHits()
	if err != nil {
		return err
	}

	if err := current.get(data["Path"], ""); err != nil {
		return err
	}

	return current.awaitHits(before, data)
}

func (current *world) awaitHits(before int64, data map[string]string) error {
	minimum, err := strconv.ParseInt(data["MinimumHits"], 10, 64)
	seconds, parseError := strconv.Atoi(data["TimeoutSeconds"])
	if err := errors.Join(err, parseError); err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(seconds) * time.Second)

	return current.pollHits(before, minimum, deadline)
}

func (current *world) pollHits(before, minimum int64, deadline time.Time) error {
	for time.Now().Before(deadline) {
		after, err := current.deployment.cacheHits()
		if err != nil {
			return err
		}

		if after-before >= minimum {
			return nil
		}

		time.Sleep(20 * time.Millisecond)
	}

	return fmt.Errorf("Produkt-ESIcache: weniger als %d Hits", minimum)
}

func (deployment *deployment) cacheHits() (int64, error) {
	return deployment.cacheCounter("MAIN.cache_hit")
}

func (deployment *deployment) cacheCounter(name string) (int64, error) {
	if deployment.systemd {
		return systemdCacheCounter(name)
	}

	varnishstat := filepath.Join(filepath.Dir(binary("SHOP_VARNISH", "/usr/sbin/varnishd")), "../bin/varnishstat")
	// #nosec G204 -- Eigenes Standardwerkzeug, selbst erzeugte Instanz ohne Shell.
	command := exec.Command(varnishstat, "-n", deployment.directory+"/varnish", "-1", "-f", name)
	if deployment.external {
		// #nosec G204 -- Festes Composewerkzeug, technischer Zählername ohne Shell.
		command = exec.Command("docker", "compose", "exec", "-T", "varnish", "varnishstat", "-1", "-f", name)
		command.Dir = "../.."
	}

	output, err := command.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("Varnishcachezähler: %w: %s", err, output)
	}

	return parseCounter(string(output), name)
}

func parseCounter(output, name string) (int64, error) {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		if fields[0] == name {
			return strconv.ParseInt(fields[1], 10, 64)
		}
	}

	return 0, fmt.Errorf("Varnishcachezähler fehlt: %s", output)
}

func (current *world) rememberProductCache(data map[string]string) error {
	seconds, err := strconv.Atoi(data["TTLSeconds"])
	if err != nil {
		return err
	}

	current.productCache = searchCacheEntry{path: data["Path"], stored: current.response.requested, ttl: time.Duration(seconds) * time.Second}

	return nil
}

func (current *world) productCacheFresh() error {
	if err := current.productCache.fresh(); err != nil {
		return fmt.Errorf("Produktinvalidierung nicht vor ursprünglichem Ablauf beobachtet: %w", err)
	}

	return nil
}
