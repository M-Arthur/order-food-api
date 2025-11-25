package handlers

import (
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/rs/zerolog"
)

// ProductHandler is the HTTP adapter for product-related endpoints.
type ProductHandler struct {
	productSvc service.ProductService
	logger     zerolog.Logger
}

func NewProductHandler(productSvc service.ProductService, l zerolog.Logger) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
		logger:     l,
	}
}

// ListProducts handles GET /product.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	products, err := h.productSvc.ListProducts(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("failed to list products")
		shared.WriteJSONError(w, r, http.StatusInternalServerError, "intrnal server error")
		return
	}

	dto := api.MapDomainProductsToDTO(products)
	shared.WriteJSON(w, r, http.StatusOK, dto)
}
