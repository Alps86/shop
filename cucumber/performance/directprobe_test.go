package performance_test

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func validateInternalOrigin(origin string) error {
	endpoint, err := url.Parse(origin)
	if err != nil {
		return err
	}

	address := net.ParseIP(endpoint.Hostname())
	if address == nil {
		return fmt.Errorf("SHOP_PERFORMANCE_INTERNAL_URL benötigt eine explizite Loopback-IP")
	}

	if !address.IsLoopback() {
		return fmt.Errorf("direkter Go-Origin muss Loopback sein")
	}

	return nil
}

func (current *world) probeDirect(path, cookie, contains string) (string, error) {
	response, err := readResponse(current.internal, path, cookie)
	if err != nil {
		return "", err
	}

	delegation := response.header.Get("X-Accel-Redirect")
	if delegation != "" {
		return "X-Accel-Redirect; empty body delegates Hugo rendering", current.verifyDelegation(path, response)
	}

	if err := validHTML(response.status, response.body, contains); err != nil {
		return "", err
	}

	return "Go-rendered HTML; no proxy includes or browser assets", nil
}

func (current *world) verifyDelegation(path string, response probeResponse) error {
	expected := map[string]string{current.data["CategoryPath"]: current.data["CategoryDelegation"], current.data["ProductPath"]: current.data["ProductDelegation"]}[path]
	if expected == "" {
		return fmt.Errorf("unerwartete Origin-Delegation")
	}

	if response.status != 200 {
		return fmt.Errorf("Delegation HTTP %d statt 200", response.status)
	}

	if response.body != "" {
		return fmt.Errorf("X-Accel-Delegation muss leeren Body haben")
	}

	if strings.Compare(response.header.Get("X-Accel-Redirect"), expected) != 0 {
		return fmt.Errorf("abweichender versionsgebundener Delegationspfad")
	}

	return nil
}
