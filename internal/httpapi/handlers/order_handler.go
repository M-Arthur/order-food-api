package handlers

import "github.com/M-Arthur/order-food-api/internal/service"

type OrderHandler struct {
	orderSvc   service.OrderService
	productSvc service.ProductService
}

func NewOrderHandler(orderSvc service.OrderService, productSvc service.ProductService) *OrderHandler {
	return &OrderHandler{
		orderSvc:   orderSvc,
		productSvc: productSvc,
	}
}
