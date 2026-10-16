package http

import (
	"github.com/Alps86/shop/internal/domain/sale"
	"net/http"
)

func (handler *Handler) requestInterests(request *http.Request) ([]sale.CategoryInterest, error) {
	if (requestView{request}).session() == "" {
		return nil, nil
	}

	if handler.config.Interests == nil {
		return nil, nil
	}

	return handler.config.Interests.GetInterests(request.Context(), (requestView{request}).session())
}
