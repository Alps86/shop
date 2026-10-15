package http

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/Alps86/shop/internal/application"
	"github.com/Alps86/shop/internal/domain/value"
	"net/http"
	"time"
)

func (view requestView) session() application.Session {
	request := view.request
	cookie, err := request.Cookie("shop_session")
	if err != nil {
		return ""
	}

	return application.Session(cookie.Value)
}

func (handler *Handler) cookie(writer http.ResponseWriter, name, token string) {
	// #nosec G124 -- Die lokale HTTPdemo setzt Secure bewusst konfigurierbar; HTTPS aktiviert es.
	http.SetCookie(writer, &http.Cookie{Name: name, Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: handler.config.SecureCookie, MaxAge: int((30 * 24 * time.Hour).Seconds())})
}

func newRandomID() string {
	bytes := make([]byte, 32)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes)
}

func (handler *Handler) csrf(writer http.ResponseWriter, request *http.Request) string {
	identity := string((requestView{request}).session())
	if identity == "" {
		cookie, err := request.Cookie("shop_start")
		identity = newRandomID()
		if err == nil {
			identity = cookie.Value
		}

		handler.cookie(writer, "shop_start", identity)
	}

	return handler.csrfToken(identity)
}

func (handler *Handler) csrfToken(identity string) string {
	mac := hmac.New(sha256.New, handler.config.CSRFKey)
	_, _ = mac.Write([]byte(identity))

	return hex.EncodeToString(mac.Sum(nil))
}

func (handler *Handler) authorize(writer http.ResponseWriter, request *http.Request) bool {
	handler.privateHeaders(writer)
	if request.Method != "POST" {
		handler.failure(writer, request, newHTTPError(405, "Änderung benötigt POST"))
		return false
	}

	request.Body = http.MaxBytesReader(writer, request.Body, 64<<10)
	if err := request.ParseForm(); err != nil {
		handler.failure(writer, request, newHTTPError(422, "Formular ungültig"))
		return false
	}

	expected := handler.csrf(writer, request)
	if !hmac.Equal([]byte(expected), []byte(request.Form.Get("csrf"))) {
		handler.failure(writer, request, newHTTPError(403, "Formular abgelaufen. Bitte neu laden."))
		return false
	}

	return true
}

func (handler *Handler) privatePage(writer http.ResponseWriter, request *http.Request, kind, title string) Page {
	handler.privateHeaders(writer)
	page := newPage(kind, title)
	page.CacheControl = "private,no-store"
	page.Robots = "noindex,nofollow"
	page.CSRF = handler.csrf(writer, request)
	page.CommandID = value.CommandID(newRandomID())

	return page
}

func (handler *Handler) guest(request *http.Request) (application.GuestView, error) {
	if (requestView{request}).session() == "" {
		return application.GuestView{}, nil
	}

	return handler.config.Guests.Open(request.Context(), (requestView{request}).session())
}

func (handler *Handler) privateHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "private,no-store")
	writer.Header().Set("X-Robots-Tag", "noindex,nofollow")
}
