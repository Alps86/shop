package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"unicode/utf8"
)

func newSearchQuery(context value.RequestContext) application.SearchQuery {
	return application.SearchQuery{Context: context, Page: 1, Sort: "relevance", Filters: application.FilterSet{Options: map[value.AttributeCode][]value.OptionCode{}, Ranges: map[value.AttributeCode]application.NumericRange{}}}
}

func (handler *Handler) query(request *http.Request, schema application.SearchSchema) (application.SearchQuery, error) {
	query, values, err := handler.queryInput(request)
	if err != nil {
		return query, err
	}

	if err := (queryInput{values}).queryValues(&query, schema); err != nil {
		return query, err
	}

	if err := query.Validate(schema); err != nil {
		return query, err
	}

	(queryInput{values}).canonical()
	request.URL.RawQuery = values.Encode()

	return query, nil
}

func (input queryInput) normalize() error {
	values := input.values
	for key, selected := range values {
		if querySelection(selected).empty() {
			values.Del(key)
			continue
		}

		slices.Sort(selected)
		values[key] = slices.Compact(selected)
		if err := input.validateSelection(key, values[key]); err != nil {
			return err
		}
	}

	return nil
}

func (input queryInput) queryValues(query *application.SearchQuery, schema application.SearchSchema) error {
	values := input.values
	query.Filters.Brands = values["brand"]
	if values.Get("sort") != "" {
		query.Sort = application.Sort(values.Get("sort"))
	}

	if values.Get("page") != "" {
		page, err := strconv.Atoi(values.Get("page"))
		if err != nil {
			return newHTTPError(422, "Ungültige Seite")
		}

		query.Page = page
	}

	if err := input.priceFilter(query); err != nil {
		return err
	}

	return input.queryFilters(query, schema)
}

func (input queryInput) queryFilters(query *application.SearchQuery, schema application.SearchSchema) error {
	values := input.values
	if input.attributeCount() > 4 {
		return newHTTPError(422, "Zu viele Attributfilter")
	}

	for key, selected := range values {
		if input.standardKey(key) {
			continue
		}

		if err := input.filterQuery(query, key, selected, schema); err != nil {
			return err
		}
	}

	return input.validView()
}

func (input queryInput) validView() error {
	view := input.values.Get("view")
	allowed := map[string]bool{"": true, "grid": true, "list": true}
	if !allowed[view] {
		return newHTTPError(422, "Unbekannte Darstellung")
	}

	return nil
}

func (handler *Handler) queryInput(request *http.Request) (application.SearchQuery, url.Values, error) {
	query := newSearchQuery(newRequestContext())
	values, parseError := url.ParseQuery(request.URL.RawQuery)
	if parseError != nil {
		return query, nil, newHTTPError(422, "Ungültige Querykodierung")
	}

	query.Text = values.Get("q")
	if utf8.RuneCountInString(query.Text) > 100 {
		return query, nil, newHTTPError(422, "Suchbegriff zu lang")
	}

	if err := (queryInput{values}).normalize(); err != nil {
		return query, nil, err
	}

	return query, values, nil
}

func (selection querySelection) empty() bool {
	return selection.matches("")
}

func (selection querySelection) matches(value string) bool {
	if len(selection) != 1 {
		return false
	}

	return selection[0] == value
}

func (input queryInput) validateSelection(key string, values []string) error {
	if len(values) > 10 {
		return newHTTPError(422, "Zu viele Filterwerte")
	}

	scalar := map[string]bool{"q": true, "page": true, "sort": true, "view": true, "priceMin": true, "priceMax": true}
	if scalar[key] {
		if len(values) > 1 {
			return newHTTPError(422, "Mehrdeutiger Queryparameter")
		}
	}

	return nil
}
