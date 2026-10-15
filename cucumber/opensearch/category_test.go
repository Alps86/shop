package opensearch_test

import (
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
)

func (current *world) categories(table *godog.Table) {
	current.config.Categories = map[value.CategoryID]catalog.Category{}
	for _, data := range read(table) {
		category := catalog.Category{ID: value.CategoryID(data["ID"]), Name: catalog.LocalizedText{current.config.Context.Locale: data["Name"]}}
		if data["ParentID"] != "absent" {
			category.ParentID = value.Optional[value.CategoryID]{Present: true, Value: value.CategoryID(data["ParentID"])}
		}

		current.config.Categories[category.ID] = category
	}

	current.adapter = opensearch.New(current.config)
}
