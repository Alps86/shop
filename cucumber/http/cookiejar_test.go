package http_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

func (current *world) prepareClients(table *godog.Table) error {
	current.engineClients = map[string]*http.Client{}
	for _, fields := range read(table) {
		if err := current.prepareClient(fields); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) prepareClient(fields map[string]string) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	endpoint, err := url.Parse(current.engine.url)
	if err != nil {
		return err
	}

	// #nosec G124 -- Synthetische Gastcookies für die echte lokale HTTP-Testengine ohne TLS.
	jar.SetCookies(endpoint, []*http.Cookie{{SameSite: http.SameSiteLaxMode, Name: "shop_session", Value: fields["Session"], Path: "/", HttpOnly: true}})
	current.engineClients[fields["Jar"]] = &http.Client{Jar: jar, Timeout: 5 * time.Second}

	return nil
}

func (current *world) engineClient(name string) *http.Client {
	if client, exists := current.engineClients[name]; exists {
		return client
	}

	return http.DefaultClient
}

func (current *world) compareFrame(table *godog.Table) error {
	if read(table)[0]["Mode"] == "remember" {
		current.sharedFrame = current.response.Body.String()
		return nil
	}

	if current.sharedFrame != current.response.Body.String() {
		return fmt.Errorf("öffentliche Rahmenantworten unterscheiden sich zwischen Gastcookiejars")
	}

	return nil
}
