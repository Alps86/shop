package performance_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func (current *world) cacheCounters() (counters, error) {
	delay, err := strconv.Atoi(current.data["CounterFlushMilliseconds"])
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Duration(delay) * time.Millisecond)

	return current.readCounters()
}

func (current *world) readCounters() (counters, error) {
	command := exec.Command("varnishstat", "-1", "-j")
	configureTool(command, "SHOP_VARNISHSTAT")
	command.Args = append(command.Args, "-n", os.Getenv("SHOP_PERFORMANCE_VARNISH_NAME"))
	if project := os.Getenv("SHOP_PERFORMANCE_COMPOSE_PROJECT"); project != "" {
		command = exec.Command("docker", "compose")
		command.Args = append(command.Args, "--project-name", project, "exec", "-T", "varnish", "varnishstat", "-1", "-j")
	}

	output, err := command.Output()
	if err != nil {
		return nil, require(err, "isolierter Standard-varnishstat-Beleg")
	}

	var report varnishReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, err
	}

	return report.values()
}

func (report varnishReport) values() (counters, error) {
	result := counters{}
	for _, key := range []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.busy_sleep"} {
		value, exists := report.Counters[key]
		if !exists {
			return nil, fmt.Errorf("Varnishzähler fehlt: %s", key)
		}

		result[key] = value.Value
	}

	return result, nil
}

func (after counters) subtract(before counters) counters {
	result := counters{}
	for key, value := range after {
		result[key] = value - before[key]
	}

	return result
}
