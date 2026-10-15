package opensearch_test

import (
	"context"
	"github.com/Alps86/shop/internal/adapter/opensearch"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/catalog"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"github.com/cucumber/godog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

func (current *world) failureSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^die Suchdienstverbindung eine explizite Protokollstörung liefert$`, current.fault)
	ctx.Step(`^der Suchadapter die Verwaltungsoperation ausführt$`, current.operation)
	ctx.Step(`^eine externe Produktprojektion mit expliziten Metadaten übernommen wird$`, current.externalProjection)
}

func (current *world) externalProjection(table *godog.Table) error {
	data := read(table)[0]
	revision, err := strconv.ParseUint(data["Revision"], 10, 64)
	if err != nil {
		return err
	}

	inventory, err := current.projectionInventory(data)
	if err != nil {
		return err
	}

	product := current.products[value.ProductID(data["PayloadProductID"])]
	product.ID = value.ProductID(data["RootID"])
	current.err = current.adapter.Apply(context.Background(), application.ProjectionEvent{ProductID: value.ProductID(data["EventProductID"]), Revision: revision, Product: value.Optional[catalog.Product]{Present: data["Present"] == "true", Value: product}, Inventory: inventory})

	return nil
}

func (current *world) projectionInventory(data row) ([]sale.Inventory, error) {
	stock, err := strconv.ParseInt(data["Stock"], 10, 64)
	if err != nil {
		return nil, err
	}

	inventory := append([]sale.Inventory{}, current.inventory[value.ProductID(data["PayloadProductID"])]...)
	for index := range inventory {
		inventory[index].StockQuantity = stock
		inventory[index].Price.Currency = value.Currency(data["Currency"])
	}

	return inventory, nil
}

func (current *world) fault(table *godog.Table) error {
	data := read(table)[0]
	if _, err := strconv.Atoi(data["Status"]); err != nil {
		return err
	}

	current.cleanupConfig = current.config
	current.faultServer = httptest.NewServer(data)
	current.config.Endpoint = current.faultServer.URL
	current.configureFault(data)
	current.adapter = opensearch.New(current.config)

	return nil
}

func (current *world) configureFault(data row) {
	if data["Connection"] == "closed" {
		current.faultServer.Close()
	}

	if data["Connection"] == "invalid" {
		current.config.Endpoint = "://invalid"
	}

	if data["Clock"] == "system" {
		current.config.Clock = nil
	}
}

func (current *world) operation(table *godog.Table) {
	data := read(table)[0]
	if data["Operation"] == "initialize" {
		current.err = current.adapter.Initialize(context.Background())
		return
	}

	if data["Operation"] == "create" {
		current.err = current.adapter.CreateIndex(context.Background(), current.config.Index+"-candidate")
		return
	}

	current.projectAgain(data)
}

func (current *world) projectAgain(data row) {
	id := value.ProductID(data["ProductID"])
	product := current.products[id]
	current.err = current.adapter.Apply(context.Background(), application.ProjectionEvent{ProductID: id, Revision: 2, Product: value.Optional[catalog.Product]{Present: true, Value: product}, Inventory: current.inventory[id]})
}

func (data row) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	status, err := strconv.Atoi(data["Status"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if data["FailPath"] != "" {
		status = data.responseStatus(request.URL.Path, status)
	}

	if data["Connection"] == "truncated" {
		data.truncateResponse(writer, request.URL.Path)
	}

	writer.WriteHeader(status)
	if _, err := writer.Write([]byte(data["Body"])); err != nil {
		return
	}
}

func (data row) responseStatus(path string, status int) int {
	if strings.Contains(path, data["FailPath"]) {
		return status
	}

	return http.StatusOK
}

func (data row) truncateResponse(writer http.ResponseWriter, path string) {
	if strings.Contains(path, data["FailPath"]) {
		writer.Header().Set("Content-Length", "1000000")
	}
}
