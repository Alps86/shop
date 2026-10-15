package http

import (
	"github.com/Alps86/shop/internal/application"
	"net/url"
)

func (view requestView) profileSelection(guest application.GuestView) (application.ProfileCode, error) {
	selected := application.ProfileCode(view.request.URL.Query().Get("profile"))
	if selected == "" {
		return guest.Profile, nil
	}

	allowed := map[application.ProfileCode]bool{"neutral": true, "mobile": true, "home": true, "coffee": true}
	if !allowed[selected] {
		return "", newHTTPError(422, "Unbekanntes Interessenprofil")
	}

	return selected, nil
}

func (view requestView) guestTarget() string {
	if view.request.Form.Get("return") == "profile" {
		return "/de-DE/profil?profile=" + url.QueryEscape(view.request.Form.Get("profile"))
	}

	target := "/de-DE/warenkorb"
	if sku := view.request.Form.Get("add"); sku != "" {
		target += "?add=" + url.QueryEscape(sku)
	}

	return target
}
