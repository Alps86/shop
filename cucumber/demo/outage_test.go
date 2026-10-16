package demo_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"os/exec"
)

func (current *world) searchState(table *godog.Table) error {
	state := read(table)[0]["State"]
	if !validSearchState(state) {
		return fmt.Errorf("unbekannter Suchzustand %q", state)
	}

	deployment := current.deployment
	deployment.mutex.Lock()
	deployment.unavailable = state == "unavailable"
	deployment.mutex.Unlock()
	if deployment.external {
		return deployment.composeSearch(state)
	}

	return nil
}

func (deployment *deployment) restoreSearch() error {
	deployment.mutex.Lock()
	unavailable := deployment.unavailable
	deployment.unavailable = false
	deployment.mutex.Unlock()
	if !unavailable {
		return nil
	}

	if deployment.external {
		return deployment.composeSearch("available")
	}

	return nil
}

func validSearchState(state string) bool {
	return state == "available" || state == "unavailable"
}

func (deployment *deployment) composeSearch(state string) error {
	if deployment.systemd {
		return deployment.systemdSearch(state)
	}

	command := exec.Command("docker", "compose", "stop", "opensearch")
	if state == "available" {
		command = exec.Command("docker", "compose", "up", "--detach", "--wait", "--wait-timeout", "180", "opensearch")
	}

	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Suchdienst %s: %w: %s", state, err, output)
	}

	return nil
}
