package runtime

import (
	"github.com/Alps86/shop/internal/adapter/cdc"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	shoptemporal "github.com/Alps86/shop/internal/adapter/temporal"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/value"
)

func (app *App) bind(snapshot application.AdminCatalogSnapshot) {
	app.Service = app.applicationService()
	app.Processes = shoptemporal.New(app.Client, app.Service, app.Store.Carts())
	app.Processes.Queue = app.Config.TaskQueue
	app.Service.Processes = app.Processes
	config := app.projectionConfig()
	app.Search = opensearch.New(app.searchConfig(snapshot, config))
	config.Target = app.Search
	app.Consumer = cdc.New(config)
	app.Handler = app.healthHandler()
}

func (app *App) applicationService() application.Service {
	return application.Service{
		Guests:      app.Store.Guests(),
		Interests:   app.Store.Guests(),
		Carts:       app.Store.Carts(),
		Commands:    app.Store.Carts(),
		Orders:      app.Store.Orders(),
		Clock:       opensearch.Clock{},
		Purchase:    app.Store.Purchases(),
		Payment:     shoptemporal.NewPaymentSimulation(true, 0),
		Writer:      app.Store.Carts(),
		Evaluator:   app.Store.Evaluator(),
		Fulfillment: app.Store.Fulfillment(),
		Starter:     app.Store.Carts(),
	}
}

func (app *App) projectionConfig() cdc.Config {
	return cdc.Config{DSN: app.Config.DSN, Schema: app.Config.Schema, Publication: app.Config.Publication, Slot: app.Config.Slot, Store: app.Store}
}

func (app *App) searchConfig(snapshot application.AdminCatalogSnapshot, replication cdc.Config) opensearch.Config {
	schema := application.SearchSchema{Attributes: map[value.AttributeCode]catalog.AttributeDefinition{}}
	for _, family := range snapshot.Families {
		for _, attribute := range family.Attributes {
			schema.Attributes[attribute.Code] = attribute
		}
	}

	categories := map[value.CategoryID]catalog.Category{}
	for _, category := range snapshot.Categories {
		categories[category.ID] = category
	}

	return opensearch.Config{Endpoint: app.Config.SearchEndpoint, Index: app.Config.SearchIndex, Context: app.Config.Context, Schema: schema, Categories: categories, PurchaseCosts: app.Seed.PurchaseCosts, Source: cdc.NewReindexSource(replication)}
}
