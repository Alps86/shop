package http

import (
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"strconv"
)

func (handler *Handler) cartAction(writer http.ResponseWriter, request *http.Request) {
	if !handler.authorize(writer, request) {
		return
	}

	command, err := handler.changeCommand(request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.performCartAction(writer, request, command)
}

func (handler *Handler) changeCommand(request *http.Request) (application.ChangeCommand, error) {
	revision, err := strconv.ParseUint(request.Form.Get("revision"), 10, 64)
	if err != nil {
		return application.ChangeCommand{}, newHTTPError(422, "Ungültige Revision")
	}

	quantity := int64(0)
	if request.Form.Has("quantity") {
		quantity, err = strconv.ParseInt(request.Form.Get("quantity"), 10, 64)
		if err != nil {
			return application.ChangeCommand{}, newHTTPError(422, "Ungültige Menge")
		}
	}

	form := request.Form
	address := sale.Address{Recipient: form.Get("recipient"), Street: form.Get("street"), PostalCode: form.Get("postalCode"), City: form.Get("city"), Country: value.CountryCode(form.Get("country"))}

	return application.ChangeCommand{CartID: value.CartID(form.Get("cartID")), CommandID: value.CommandID(form.Get("commandID")), ExpectedRevision: revision, Operation: form.Get("action"), SKU: value.SKU(form.Get("sku")), Quantity: quantity, ShippingMethod: sale.ShippingMethod(form.Get("shipping")), PaymentMethod: sale.PaymentMethod(form.Get("payment")), Address: address}, nil
}

func (handler *Handler) mutateCart(request *http.Request, command application.ChangeCommand) (application.CartView, error) {
	if command.Operation == "confirm" {
		confirm := application.ConfirmCommand{CartID: command.CartID, CommandID: command.CommandID, ExpectedRevision: command.ExpectedRevision, EvaluationID: value.EvaluationID(request.Form.Get("evaluationID"))}
		return handler.config.Guests.Confirm(request.Context(), (requestView{request}).session(), confirm)
	}

	if command.Operation == "evaluate" {
		return handler.config.Guests.Evaluate(request.Context(), (requestView{request}).session(), command.CartID)
	}

	return handler.config.Guests.Change(request.Context(), (requestView{request}).session(), command)
}

func (handler *Handler) startCart(writer http.ResponseWriter, request *http.Request, id value.CommandID) {
	result, err := handler.config.Guests.StartCart(request.Context(), (requestView{request}).session(), id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.afterCart(writer, request, result)
}

func (handler *Handler) afterCart(writer http.ResponseWriter, request *http.Request, cart application.CartView) {
	target := "/de-DE/warenkorb"
	if request.Form.Get("return") == "checkout" {
		target = "/de-DE/kasse"
	}

	if request.Header.Get("HX-Request") != "true" {
		http.Redirect(writer, request, target, http.StatusSeeOther)
		return
	}

	page := handler.privatePage(writer, request, "cart", "Dein Warenkorb")
	if target == "/de-DE/kasse" {
		page.Kind = "checkout"
	}

	page.useCart(cart)
	handler.cartRecommendations(request, &page)
	writer.Header().Set("HX-Push-Url", target)
	writer.Header().Set("HX-Trigger", "cartChanged")
	handler.render(writer, request, page)
}

func (handler *Handler) mutationError(writer http.ResponseWriter, request *http.Request, cart application.CartView, err error) {
	page := handler.privatePage(writer, request, "cart", "Dein Warenkorb")
	page.useCart(cart)
	if request.Form.Get("return") == "checkout" {
		page.Kind = "checkout"
	}

	handler.cartError(writer, request, page, err)
}

func (handler *Handler) place(writer http.ResponseWriter, request *http.Request, command application.ChangeCommand) {
	place := application.PlaceCommand{CartID: command.CartID, CommandID: command.CommandID, EvaluationID: value.EvaluationID(request.Form.Get("evaluationID"))}
	result, err := handler.config.Guests.Place(request.Context(), (requestView{request}).session(), place)
	if err != nil {
		handler.purchaseError(writer, request, result, err)
		return
	}

	if result.Pending {
		handler.pendingPurchase(writer, request, result, command)
		return
	}

	handler.redirectOrder(writer, request, result.OrderID)
}

func (handler *Handler) redirectOrder(writer http.ResponseWriter, request *http.Request, id value.OrderID) {
	target := "/de-DE/bestellung/" + string(id)
	if request.Header.Get("HX-Request") == "true" {
		writer.Header().Set("HX-Redirect", target)
		writer.WriteHeader(http.StatusOK)

		return
	}

	http.Redirect(writer, request, target, http.StatusSeeOther)
}

func (handler *Handler) purchaseError(writer http.ResponseWriter, request *http.Request, result application.PurchaseResult, err error) {
	page := handler.privatePage(writer, request, "checkout", "Gastcheckout")
	page.useCart(result.Cart.Value)
	page.Alternatives = result.Alternatives
	page.UnavailableSKU = result.UnavailableSKU
	handler.cartError(writer, request, page, err)
}

func (handler *Handler) performCartAction(writer http.ResponseWriter, request *http.Request, command application.ChangeCommand) {
	if command.Operation == "start" {
		handler.startCart(writer, request, command.CommandID)
		return
	}

	if command.Operation == "place" {
		handler.place(writer, request, command)
		return
	}

	result, err := handler.mutateCart(request, command)
	if err != nil {
		handler.mutationError(writer, request, result, err)
		return
	}

	handler.afterCart(writer, request, result)
}

func (handler *Handler) pendingPurchase(writer http.ResponseWriter, request *http.Request, result application.PurchaseResult, command application.ChangeCommand) {
	page := handler.privatePage(writer, request, "checkout", "Bestellung wird verarbeitet")
	page.Status = http.StatusAccepted
	page.Message = "Bestellung angenommen, Status wird aktualisiert"
	page.CommandID = command.CommandID
	page.useCart(result.Cart.Value)
	writer.Header().Set("Location", "/de-DE/kasse")
	writer.Header().Set("Retry-After", "2")
	handler.render(writer, request, page)
}

func (page *Page) useCart(cart application.CartView) {
	page.Cart = cart
	page.Guest.UserID = cart.Cart.UserID
}
