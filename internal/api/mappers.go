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

// MapOrderReqToDomain builds a domain.Order from the incoming OrderReqDTO.
//
// It performs adapter-level validation with field-specific messages, then
// delegates to domain.NewOrder for invariant checking.
func MapOrderReqToDomain(orderID domain.OrderID, req OrderReqDTO) (*domain.Order, error) {
	items := make([]domain.OrderItem, 0, len(req.Items))

	for i, item := range req.Items {
		fieldProductID := fmt.Sprintf("items[%d].productId", i)
		fieldQuantity := fmt.Sprintf("items[%d].quantity", i)

		if item.ProductID == "" {
			return nil, &ValidationError{
				Field:   fieldProductID,
				Message: "required",
			}
		}
		if item.Quantity < 1 {
			return nil, &ValidationError{
				Field:   fieldQuantity,
				Message: "must be >= 1",
			}
		}

		items = append(items, domain.OrderItem{
			ProductID: domain.ProductID(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return domain.NewOrder(orderID, items, req.CouponCode)
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
