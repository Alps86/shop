package demo_test

import (
	"fmt"
	"github.com/cucumber/godog"
	"net/url"
)

func (current *world) preparePurchases(table *godog.Table) error {
	for _, data := range read(table) {
		if err := current.preparePurchase(data); err != nil {
			return err
		}
	}

	return nil
}

func (current *world) preparePurchase(data map[string]string) error {
	if err := current.selectName(data["Jar"]); err != nil {
		return err
	}

	path := "/de-DE/warenkorb?add=" + url.QueryEscape(data["sku"])
	if err := current.get(path, ""); err != nil {
		return err
	}

	if err := current.submitData(map[string]string{"FormAction": "/de-DE/gast"}); err != nil {
		return err
	}

	if err := current.get(path, ""); err != nil {
		return err
	}

	return current.configurePurchase(data)
}

func (current *world) configurePurchase(data map[string]string) error {
	if err := current.submitCart(map[string]string{"action": "add", "sku": data["sku"], "quantity": data["quantity"]}); err != nil {
		return err
	}

	if err := current.get("/de-DE/kasse", ""); err != nil {
		return err
	}

	address := map[string]string{"action": "address", "recipient": data["recipient"], "street": data["street"], "postalCode": data["postalCode"], "city": data["city"], "country": data["country"]}
	if err := current.submitCart(address); err != nil {
		return err
	}

	return current.configurePayment(data)
}

func (current *world) configurePayment(data map[string]string) error {
	for _, form := range []map[string]string{{"action": "shipping", "shipping": data["shipping"]}, {"action": "payment", "payment": data["payment"]}, {"action": "confirm"}} {
		if err := current.submitCart(form); err != nil {
			return err
		}
	}

	values, err := current.selectedForm(map[string]string{"FormAction": "/de-DE/warenkorb/aktion", "action": "place", "return": "checkout"})
	current.prepared[data["Jar"]] = values

	return err
}

func (current *world) submitCart(data map[string]string) error {
	data["FormAction"] = "/de-DE/warenkorb/aktion"
	data["return"], data["HX"] = "checkout", "true"
	if err := current.submitData(data); err != nil {
		return err
	}

	if current.response.status != 200 {
		return fmt.Errorf("Gastvorbereitung: HTTP %d %s", current.response.status, current.response.diagnostic())
	}

	return nil
}

func (current *world) submitData(data map[string]string) error {
	values, err := current.selectedForm(data)
	if err != nil {
		return err
	}

	return current.post(data["FormAction"], data["HX"], values)
}

func (current *world) submitPrepared(table *godog.Table) error {
	jar := read(table)[0]["Jar"]
	if err := current.selectName(jar); err != nil {
		return err
	}

	current.saved = cloneValues(current.prepared[jar])

	return current.post("/de-DE/warenkorb/aktion", "", current.saved)
}
