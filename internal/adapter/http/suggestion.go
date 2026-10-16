package http

import "net/url"

func suggestionURL(text string, form map[string][]string) string {
	values := url.Values{}
	for key, entries := range form {
		values[key] = entries
	}

	values.Set("q", text)
	values.Del("page")

	return "/de-DE/suche?" + values.Encode()
}
