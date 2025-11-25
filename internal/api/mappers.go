package api

import (
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

// MapOrderReqToDomain builds a domain.Order from the incoming OrderReqDTO.
func MapOrderReqToDomain(orderID domain.OrderID, req OrderReqDTO) (*domain.Order, error) {
	items := make([]domain.OrderItem, 0, len(req.Items))

	for i, item := range req.Items {
		if item.ProductID == "" {
			return nil, fmt.Errorf("items[%d].productId: required", i)
		}
		if item.Quantity < 1 {
			return nil, fmt.Errorf("items[%d].quantity: must be >= 1", i)
		}

		items = append(items, domain.OrderItem{
			ProductID: domain.ProductID(item.ProductID),
			Quantity:  item.Quantity,
		})
	}

	return domain.NewOrder(orderID, items, req.CouponCode)
}

// MapDomainProductToDTO converts a domain.Product to the API representation
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
