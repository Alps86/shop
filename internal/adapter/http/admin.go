package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"strconv"
	"strings"
)

func (handler *Handler) admin(writer http.ResponseWriter, request *http.Request) {
	routes := map[string]http.HandlerFunc{"/backend": handler.adminIndex, "/backend/inventory": handler.adminInventory, "/backend/lieferschritt": handler.adminDelivery}
	if route, found := routes[request.URL.Path]; found {
		route(writer, request)
		return
	}

	if strings.HasPrefix(request.URL.Path, "/backend/bestellung/") {
		handler.adminOrder(writer, request)
		return
	}

	handler.failure(writer, request, newHTTPError(404, "Nicht gefunden"))
}

func (handler *Handler) adminIndex(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page, err := handler.adminPage(writer, request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.render(writer, request, page)
}

func (handler *Handler) adminOrder(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	id := value.OrderID(strings.TrimPrefix(request.URL.Path, "/backend/bestellung/"))
	order, err := handler.config.AdminOrders.Get(request.Context(), id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	page := handler.privatePage(writer, request, "adminOrder", "Lieferung simulieren")
	page.Order = order
	handler.render(writer, request, page)
}

func (handler *Handler) adminInventory(writer http.ResponseWriter, request *http.Request) {
	if !handler.authorize(writer, request) {
		return
	}

	command, err := newInventoryCommand(request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	_, err = handler.config.Inventory.Change(request.Context(), command)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	http.Redirect(writer, request, "/backend", http.StatusSeeOther)
}

func newInventoryCommand(request *http.Request) (application.InventoryCommand, error) {
	revision, err := strconv.ParseUint(request.Form.Get("revision"), 10, 64)
	if err != nil {
		return application.InventoryCommand{}, newHTTPError(422, "Ungültige Revision")
	}

	command := application.InventoryCommand{SKU: value.SKU(request.Form.Get("sku")), ExpectedRevision: revision, CommandID: value.CommandID(request.Form.Get("commandID"))}
	err = (inventoryInput{request.Form}).fill(&command)

	return command, err
}

func (handler *Handler) adminDelivery(writer http.ResponseWriter, request *http.Request) {
	if !handler.authorize(writer, request) {
		return
	}

	quantity, err := strconv.ParseInt(request.Form.Get("quantity"), 10, 64)
	if err != nil {
		handler.failure(writer, request, newHTTPError(422, "Ungültige Teilmenge"))
		return
	}

	form := request.Form
	command := application.DeliveryCommand{OrderID: value.OrderID(form.Get("orderID")), CommandID: value.CommandID(form.Get("commandID"))}
	command.Update = sale.FulfillmentUpdate{OrderID: command.OrderID, CommandID: command.CommandID, ShipmentID: value.ShipmentID(form.Get("shipmentID")), Status: sale.DeliveryStatus(form.Get("status")), Lines: []sale.FulfillmentLine{{SKU: value.SKU(form.Get("sku")), Quantity: quantity}}}
	if err := handler.config.Processes.ReportDelivery(request.Context(), command); err != nil {
		handler.failure(writer, request, err)
		return
	}

	// #nosec G710 -- Fester lokaler Routenpräfix, OrderID wird vom Applicationport validiert.
	http.Redirect(writer, request, "/backend/bestellung/"+string(command.OrderID), http.StatusSeeOther)
}

func (input inventoryInput) fill(command *application.InventoryCommand) error {
	form := input.values
	if form.Has("price") {
		price, err := strconv.ParseInt(form.Get("price"), 10, 64)
		if err != nil {
			return newHTTPError(422, "Ungültiger Preis in Cent")
		}

		command.Price = value.Optional[value.Money]{Value: value.Money{MinorUnits: price, Currency: "EUR"}, Present: true}
	}

	if form.Has("stock") {
		stock, err := strconv.ParseInt(form.Get("stock"), 10, 64)
		if err != nil {
			return newHTTPError(422, "Ungültiger Bestand")
		}

		command.Stock = value.Optional[int64]{Value: stock, Present: true}
	}

	return nil
}

func (handler *Handler) adminPage(writer http.ResponseWriter, request *http.Request) (Page, error) {
	page := handler.privatePage(writer, request, "admin", "Lokale Verwaltung")
	snapshot, err := handler.config.AdminCatalog.Snapshot(request.Context())
	if err != nil {
		return page, err
	}

	if err := handler.adminOrders(request, &page); err != nil {
		return page, err
	}

	page.Diagnostics = handler.diagnostics(request.Context())
	page.AdminCatalog = snapshot

	return page, nil
}

func (handler *Handler) adminOrders(request *http.Request, page *Page) error {
	number, err := (requestView{request}).adminPageNumber()
	if err != nil {
		return err
	}

	orders, err := handler.config.AdminOrders.List(request.Context(), number)
	if err != nil {
		return err
	}

	page.AdminOrders = orders
	page.Query.Page = orders.Page
	page.Search.TotalProducts = orders.Total
	page.pagination(request)

	return nil
}
