package http

import (
	"fmt"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/sale"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"strings"
)

func (handler *Handler) openGuest(writer http.ResponseWriter, request *http.Request) {
	if !handler.authorize(writer, request) {
		return
	}

	guest, err := handler.config.Guests.Open(request.Context(), (requestView{request}).session())
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.startGuestCart(writer, request, guest)
}

func (handler *Handler) profile(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		handler.selectProfile(writer, request)
		return
	}

	if !handler.readOnly(writer, request) {
		return
	}

	page, err := handler.profilePage(writer, request)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	handler.render(writer, request, page)
}

func (handler *Handler) selectProfile(writer http.ResponseWriter, request *http.Request) {
	if !handler.authorize(writer, request) {
		return
	}

	_, err := handler.config.Guests.SelectProfile(request.Context(), (requestView{request}).session(), application.ProfileCode(request.Form.Get("profile")))
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	http.Redirect(writer, request, "/de-DE/profil", http.StatusSeeOther)
}

func (handler *Handler) cart(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page := handler.privatePage(writer, request, "cart", "Dein Warenkorb")
	if err := handler.loadCart(request, &page); err != nil {
		handler.cartError(writer, request, page, err)
		return
	}

	handler.cartRecommendations(request, &page)
	if err := handler.cartPreview(request, &page); err != nil {
		handler.cartError(writer, request, page, err)
		return
	}

	handler.render(writer, request, page)
}

func (handler *Handler) loadCart(request *http.Request, page *Page) error {
	guest, err := handler.guest(request)
	if err != nil {
		return err
	}

	page.Guest = guest
	if !guest.CartID.Present {
		return nil
	}

	page.Cart, err = handler.config.Guests.GetCart(request.Context(), guest.Session, guest.CartID.Value)

	return err
}

func (handler *Handler) cartError(writer http.ResponseWriter, request *http.Request, page Page, err error) {
	page.Status = (responseError{err}).status()
	page.Message = (responseError{err}).message()
	page.retry(request)
	if page.Status == 202 {
		writer.Header().Set("Retry-After", "2")
		writer.Header().Set("Location", "/de-DE/warenkorb")
	}

	handler.render(writer, request, page)
}

func (handler *Handler) checkout(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page := handler.privatePage(writer, request, "checkout", "Gastcheckout")
	if err := handler.loadCart(request, &page); err != nil {
		handler.cartError(writer, request, page, err)
		return
	}

	if handler.checkoutRedirect(writer, request, page) {
		return
	}

	handler.render(writer, request, page)
}

func (handler *Handler) order(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	id := value.OrderID(strings.TrimPrefix(request.URL.Path, "/de-DE/bestellung/"))
	order, err := handler.config.Guests.GetOrder(request.Context(), (requestView{request}).session(), id)
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	page := handler.privatePage(writer, request, "order", "Deine Demo-Bestellung")
	page.Order = order
	handler.render(writer, request, page)
}

func (handler *Handler) badge(writer http.ResponseWriter, request *http.Request) {
	if !handler.readOnly(writer, request) {
		return
	}

	page := handler.privatePage(writer, request, "badge", "Warenkorb")
	if err := handler.loadCart(request, &page); err != nil {
		handler.failure(writer, request, err)
		return
	}

	writer.Header().Set("Cache-Control", page.CacheControl)
	writer.Header().Set("X-Robots-Tag", page.Robots)
	quantity := int64(0)
	for _, item := range page.Cart.Cart.Items {
		quantity += item.Quantity
	}

	_, _ = fmt.Fprintf(writer, "Warenkorb (%d)", quantity)
}

func (handler *Handler) startGuestCart(writer http.ResponseWriter, request *http.Request, guest application.GuestView) {
	handler.cookie(writer, "shop_session", string(guest.Session))
	_, err := handler.config.Guests.StartCart(request.Context(), guest.Session, value.CommandID(request.Form.Get("commandID")))
	if err != nil {
		handler.failure(writer, request, err)
		return
	}

	http.Redirect(writer, request, (requestView{request}).guestTarget(), http.StatusSeeOther)
}

func (handler *Handler) checkoutRedirect(writer http.ResponseWriter, request *http.Request, page Page) bool {
	if page.Cart.OrderID.Present {
		http.Redirect(writer, request, "/de-DE/bestellung/"+string(page.Cart.OrderID.Value), http.StatusSeeOther)
		return true
	}

	if len(page.Cart.Cart.Items) == 0 {
		http.Redirect(writer, request, "/de-DE/warenkorb", http.StatusSeeOther)
		return true
	}

	return false
}

func (handler *Handler) cartRecommendations(request *http.Request, page *Page) {
	if handler.config.Variants == nil {
		return
	}

	for _, item := range page.Cart.Cart.Items {
		handler.cartItem(request, page, item)
	}
}

func (handler *Handler) cartItem(request *http.Request, page *Page, item sale.CartItem) {
	detail, err := handler.config.Variants.ProductForSKU(request.Context(), item.SKU, page.Context)
	if err != nil {
		return
	}

	page.Products = append(page.Products, detail)
	query := application.RecommendationQuery{ProductID: detail.Product.ID, SKU: item.SKU, Context: page.Context, Limit: 4}
	recommendations := handler.recommendations(request, query)
	page.Recommendations.Cards = append(page.Recommendations.Cards, recommendations.Cards...)
	page.Recommendations.Label = recommendations.Label
	page.Recommendations.Degraded = page.Recommendations.Degraded || recommendations.Degraded
}

func (handler *Handler) cartPreview(request *http.Request, page *Page) error {
	sku := value.SKU(request.URL.Query().Get("vorschau"))
	if sku == "" {
		return nil
	}

	if handler.config.Preview == nil {
		return newHTTPError(503, "Vorschau vorübergehend nicht verfügbar")
	}

	evaluation, err := handler.config.Preview.PreviewAccessory(request.Context(), (requestView{request}).session(), page.Cart.Cart.ID, sku)
	if err != nil {
		return err
	}

	page.Preview = value.Optional[sale.CartEvaluation]{Present: true, Value: evaluation}
	page.PreviewSKU = sku

	return nil
}

func (handler *Handler) profilePage(writer http.ResponseWriter, request *http.Request) (Page, error) {
	page := handler.privatePage(writer, request, "profile", "Deine Interessen")
	guest, err := handler.guest(request)
	if err != nil {
		return page, err
	}

	page.Guest = guest
	page.ProfileChoice, err = (requestView{request}).profileSelection(guest)

	return page, err
}
