package demo_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"
	"net/url"
	"strings"
)

func (current *world) registerSearchCorrection(ctx *godog.ScenarioContext) {
	ctx.Step(`^die echte Suche erhält Eingabe und Treffer unverändert$`, current.observeSearchCorrection)
	ctx.Step(`^der Gast folgt dem tatsächlich angebotenen Korrekturlink$`, current.followCorrection)
}

func (current *world) observeSearchCorrection(table *godog.Table) error {
	data := read(table)[0]
	if !current.hasSearchInput(data["Query"]) {
		return fmt.Errorf("ursprüngliche Sucheingabe %q fehlt im tatsächlichen Formular", data["Query"])
	}

	expected := map[string]string{"Status": "200", "Contains": data["Models"] + " Modelle"}
	if err := current.observeRow(expected); err != nil {
		return err
	}

	document, err := html.Parse(strings.NewReader(current.response.body))
	if err != nil {
		return err
	}

	canonical := findSearchLink(document, "link", "rel", "canonical")

	return checkSearchURL(canonical, data["CanonicalQuery"])
}

func (current *world) hasSearchInput(text string) bool {
	for _, form := range current.forms {
		if form.action == "/de-DE/suche" {
			if form.values.Get("q") == text {
				return true
			}
		}
	}

	return false
}

func (current *world) followCorrection(table *godog.Table) error {
	data := read(table)[0]
	document, err := html.Parse(strings.NewReader(current.response.body))
	if err != nil {
		return err
	}

	href := findCorrectionLink(document, data["LinkText"])
	if href == "" {
		return fmt.Errorf("sichtbarer Korrekturlink %q fehlt", data["LinkText"])
	}

	if err := checkSearchURL(href, data["QueryString"]); err != nil {
		return err
	}

	return current.get(href, "")
}

func checkSearchURL(href, expected string) error {
	actual, err := url.Parse(href)
	if err != nil {
		return err
	}

	if !localSearchURL(actual) {
		return fmt.Errorf("kein lokaler Suchlink: %q", href)
	}

	wanted, err := url.ParseQuery(expected)
	if err != nil {
		return err
	}

	if actual.Query().Encode() != wanted.Encode() {
		return fmt.Errorf("Suchlinkquery %q statt %q", actual.RawQuery, expected)
	}

	return nil
}

func localSearchURL(link *url.URL) bool {
	return link.Scheme == "" && link.Host == "" && link.Path == "/de-DE/suche" && link.Fragment == ""
}

func findCorrectionLink(node *html.Node, text string) string {
	if correctionLinkMatches(node, text) {
		return attribute(node, "href")
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if href := findCorrectionLink(child, text); href != "" {
			return href
		}
	}

	return ""
}

func correctionLinkMatches(node *html.Node, text string) bool {
	return node.Data == "a" && strings.TrimSpace(searchNodeText(node)) == text
}

func searchNodeText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	text := ""
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text += searchNodeText(child)
	}

	return text
}

func findSearchLink(node *html.Node, tag, name, value string) string {
	if node.Data == tag {
		if attribute(node, name) == value {
			return attribute(node, "href")
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if href := findSearchLink(child, tag, name, value); href != "" {
			return href
		}
	}

	return ""
}
