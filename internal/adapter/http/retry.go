package http

import "net/http"

func (page *Page) retry(request *http.Request) {
	if request.Method != "POST" {
		return
	}

	page.RetryAction = request.URL.Path
	labels := map[string]string{"quantity": "Menge", "price": "Preis in Cent", "stock": "Bestand", "recipient": "Name", "street": "Straße", "postalCode": "Postleitzahl", "city": "Ort", "shipping": "Versand", "payment": "Zahlung", "shipmentID": "Lieferkennung", "status": "Lieferfortschritt"}
	for name, values := range request.Form {
		for _, input := range values {
			page.RetryFields = append(page.RetryFields, InputField{Name: name, Value: input, Label: labels[name], Hidden: labels[name] == ""})
		}
	}
}
