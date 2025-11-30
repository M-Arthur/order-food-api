package api

// ProductDTO matches components.schemas.Product exactly.
// @Description Product model
// @Name Product
// swagger:model ProductDTO
type ProductDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// OrderItemDTO is the inline object used in Order and OrderReq
// swagger:model OrderItemDTO
type OrderItemDTO struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// OrderReqDTO matches components.schema.OrderReq
// Used for POST /order request bodies
// swagger:model OrderReq
type OrderReqDTO struct {
	CouponCode *string        `json:"couponCode,omitempty"`
	Items      []OrderItemDTO `json:"items"`
}

// OrderDTO matches components.schemas.Order
// used for responses from POST /order (and potentially GET /order in future)
// swagger:model Order
type OrderDTO struct {
	ID         string         `json:"id"`
	Items      []OrderItemDTO `json:"items"`
	Products   []ProductDTO   `json:"products"`
	CouponCode string         `json:"couponCode"`
}

// ApiResponseDTO matches components.schemas.ApiResponse
// swagger:model ApiResponse
type ApiResponseDTO struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}
