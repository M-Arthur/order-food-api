package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// ProductHandler is the HTTP adapter for product-related endpoints.
type ProductHandler struct {
	productSvc service.ProductService
}

func NewProductHandler(productSvc service.ProductService) *ProductHandler {
	return &ProductHandler{
		productSvc: productSvc,
	}
}

// ListProducts handles GET /product.
//
// @Summary List products
// @Description Get all products available for order
// @Tags product
// @Produce json
// @Success 200 {array} api.ProductDTO
// @Router /product [get]
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

// GetProductByID handles GET /product/{productId}
//
// @Summary Find product by ID
// @Description Returns a single product
// @Tags product
// @Produce json
// @Param productId path int true "ID of product to return"
// @Success 200 {object} api.ProductDTO
// @Failure 400 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /product/{productId} [get]
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	idStr := chi.URLParam(r, "productId")
	if idStr == "" {
		shared.WriteJSONError(w, r, http.StatusBadRequest, "Invalid ID supplied")
		return
	}

	// per OpenAPI: productId is int64
	if _, err := strconv.ParseInt(idStr, 10, 64); err != nil {
		logger.Warn().Str("productId", idStr).Err(err).Msg("invalid product ID format")
		shared.WriteJSONError(w, r, http.StatusBadRequest, "invalid ID supplied")
		return
	}

	product, err := h.productSvc.GetProduct(ctx, domain.ProductID(idStr))
	if errors.Is(err, domain.ErrProductNotFound) {
		logger.Warn().Str("productId", idStr).Msg("product not found")
		shared.WriteJSONError(w, r, http.StatusNotFound, "Product not found")
		return
	}
	if err != nil {
		logger.Error().Str("productId", idStr).Err(err).Msg("product not found")
		// It should return 500 HTTP status code in PROD
		shared.WriteJSONError(w, r, http.StatusNotFound, "Product not found")
		return
	}

	dto := api.MapDomainProductToDTO(*product)
	shared.WriteJSON(w, r, http.StatusOK, dto)
}
