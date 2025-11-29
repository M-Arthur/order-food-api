package api

import (
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

// ValidationError is an API-level validation error, suitable for mapping to 422.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// OrderPayload is a helper struct used between adapter and service layers.
type OrderPayload struct {
	Items      []domain.OrderItem
	CouponCode *string
}

// MapOrderReqToPayload validates the request and returns domain-friendly data.
func MapOrderReqToPayload(req OrderReqDTO) (*OrderPayload, error) {
	items := make([]domain.OrderItem, 0, len(req.Items))

	for i, item := range req.Items {
		if item.ProductID == "" {
			return nil, &ValidationError{
				Field:   fmt.Sprintf("items[%d].productId", i),
				Message: "required",
			}
		}
		if item.Quantity < 1 {
			return nil, &ValidationError{
				Field:   fmt.Sprintf("items[%d].quantity", i),
				Message: "must be >= 1",
			}
		}

		items = append(items, domain.OrderItem{
			ProductID: domain.ProductID(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return &OrderPayload{
		Items:      items,
		CouponCode: req.CouponCode,
	}, nil
}

// MapDomainProductToDTO converts a domain.Product to the API representation.
func MapDomainProductToDTO(p domain.Product) ProductDTO {
	return ProductDTO{
		ID:       string(p.ID),
		Name:     p.Name,
		Price:    p.Price.ToFloat(),
		Category: p.Category,
	}
}

func MapDomainProductsToDTO(products []domain.Product) []ProductDTO {
	out := make([]ProductDTO, 0, len(products))
	for _, p := range products {
		out = append(out, MapDomainProductToDTO(p))
	}
	return out
}

// MapDomainOrderToDTO converts a domain.Order plus a set of products
// into the OpenAPI Order shape.
func MapDomainOrderToDTO(order *domain.Order, products []domain.Product) OrderDTO {
	itemDTOs := make([]OrderItemDTO, 0, len(order.Items))
	for _, item := range order.Items {
		itemDTOs = append(itemDTOs, OrderItemDTO{
			ProductID: string(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return OrderDTO{
		ID:       string(order.ID),
		Items:    itemDTOs,
		Products: MapDomainProductsToDTO(products),
	}
}
