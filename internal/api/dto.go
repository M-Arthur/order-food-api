package api

// These types are 1:1 with the given OpenAPI schemas

// ProductDTO matches components.schemas.Product exactly.
type ProductDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// OrderItemDTO is the inline object used in Order and OrderReq
type OrderItemDTO struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderReqDTO matches components.schema.OrderReq
// Used for POST /order request bodies
type OrderReqDTO struct {
	CouponCode *string        `json:"couponCode,omitempty"`
	Items      []OrderItemDTO `json:"items"`
}

// OrderDTO matches components.schemas.Order
// used for responses from POST /order (and potentially GET /order in future)
type OrderDTO struct {
	ID       string         `json:"id"`
	Items    []OrderItemDTO `json:"items"`
	Products []ProductDTO   `json:"products"`
}

// ApiResponseDTO matches components.schemas.ApiResponse
type ApiResponseDTO struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}
