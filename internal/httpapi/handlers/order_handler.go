package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/M-Arthur/order-food-api/internal/api"
	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/M-Arthur/order-food-api/internal/httpapi/shared"
	"github.com/M-Arthur/order-food-api/internal/service"
	"github.com/rs/zerolog"
)

type OrderHandler struct {
	orderSvc service.OrderService
}

func NewOrderHandler(orderSvc service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderSvc: orderSvc,
	}
}

// PlaceOrder handles POST /order.
//
// Promo code validation will be added later; for now, we accept couponCode
// but only validate shape / items.
func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	var req api.OrderReqDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("invalid JSON for order request")
		shared.WriteJSONError(w, r, http.StatusBadRequest, "Invalid input")
		return
	}

	payload, err := api.MapOrderReqToPayload(req)
	if err != nil {
		var ve *api.ValidationError
		if errors.As(err, &ve) {
			logger.Warn().Err(err).Msg("order validation error (adptor)")
			shared.WriteJSONError(w, r, http.StatusUnprocessableEntity, ve.Error())
			return
		}

		// fallback
		logger.Error().Err(err).Msg("internal server eror")
		shared.WriteJSONError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}

	orders, products, err := h.orderSvc.CreateOrder(ctx, payload.Items, payload.CouponCode)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			logger.Warn().Err(err).Msg("unknown product in order items")
			shared.WriteJSONError(w, r, http.StatusBadRequest, "invalid product in items")
			return
		}

		logger.Error().Err(err).Msg("internal server eror")
		shared.WriteJSONError(w, r, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := api.MapDomainOrderToDTO(orders, products)
	shared.WriteJSON(w, r, http.StatusOK, resp)
}
