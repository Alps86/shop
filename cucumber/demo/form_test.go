package demo_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
)

func parseForms(source string) []form {
	document, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return nil
	}

	return collectForms(document)
}

func collectForms(node *html.Node) []form {
	result := []form{}
	if node.Data == "form" {
		entry := form{action: attribute(node, "action"), values: url.Values{}, buttons: url.Values{}}
		entry.controls(node)

		return append(result, entry)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result = append(result, collectForms(child)...)
	}

	return result
}

func attribute(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}

	return ""
}

func (form *form) controls(node *html.Node) {
	name := attribute(node, "name")
	if node.Data == "input" {
		form.values.Set(name, attribute(node, "value"))
	}

	if node.Data == "button" {
		form.buttons.Add(name, attribute(node, "value"))
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		form.controls(child)
	}
}

func (form form) matches(data map[string]string) bool {
	if form.action != data["FormAction"] {
		return false
	}

	if action := data["action"]; action != "" {
		if !strings.Contains("|"+strings.Join(form.buttons["action"], "|")+"|", "|"+action+"|") {
			return false
		}
	}

	return form.matchesSKU(data["sku"])
}

func (form form) matchesSKU(sku string) bool {
	if form.values.Get("sku") == "" {
		return true
	}

	return form.values.Get("sku") == sku
}

func (current *world) selectedForm(data map[string]string) (url.Values, error) {
	for _, form := range current.forms {
		if form.matches(data) {
			return form.fill(data), nil
		}
	}

	return nil, fmt.Errorf("sichtbares Formular %v fehlt: %s", data, current.response.diagnostic())
}

func (form form) fill(data map[string]string) url.Values {
	values := cloneValues(form.values)
	for key, value := range data {
		if ignoredFormKey(key) {
			continue
		}

		values.Set(key, value)
	}

	return values
}

func ignoredFormKey(key string) bool {
	return key == "FormAction" || key == "Status" || key == "HX" || key == "Contains"
}

func (current *world) submit(table *godog.Table) error {
	data := read(table)[0]
	values, err := current.selectedForm(data)
	if err != nil {
		return err
	}

	return current.post(data["FormAction"], data["HX"], values)
}

func (current *world) post(path, hx string, values url.Values) error {
	target := current.deployment.public
	if strings.HasPrefix(path, "/backend") {
		target = current.deployment.backend
	}

	request, err := http.NewRequest(http.MethodPost, target+path, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", hx)
	current.cartID = values.Get("cartID")

	return current.send(request)
}

func (current *world) rememberForm(table *godog.Table) error {
	var err error
	current.saved, err = current.selectedForm(read(table)[0])

	return err
}

func (current *world) replay(table *godog.Table) error {
	data := read(table)[0]
	return current.post(data["FormAction"], data["HX"], cloneValues(current.saved))
}
