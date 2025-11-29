package domain

import (
	"context"
	"errors"
	"fmt"
	"math"
)

var (
	ErrEmptyOrderItems   = errors.New("order must contain at least one item")
	ErrInvalidQuantity   = errors.New("item quantity must be >= 1")
	ErrInvalidProductID  = errors.New("product ID must be non-empty")
	ErrInvalidOrderID    = errors.New("order ID must be non-empty")
	ErrInvalidCouponCode = errors.New("coupon code cannot be empty string") // if present

	ErrProductNotFound = errors.New("product not found")
)

// Strongly typed IDs for clarity and type safety.
type (
	ProductID string
	OrderID   string
)

// Money represents currency in integer cents to avoid float issues.
type Money int64

// NewMoneyFromFloat creates Money from a float, rounding to the nearest cent.
func NewMoneyFromFloat(amount float64) Money {
	return Money(math.Round(amount * 100))
}

func (m Money) ToFloat() float64 {
	return float64(m) / 100
}

// Product is a domain representation of a purchasable item.
type Product struct {
	ID       ProductID
	Name     string
	Price    Money
	Category string
}

// OrderItem represents a product + quantity within an order.
type OrderItem struct {
	ProductID ProductID
	Quantity  int
}

// Order is a domain aggregate for a placed order.
type Order struct {
	ID         OrderID
	Items      []OrderItem
	CouponCode *string // optional
}

// NewOrder builds a valid Order and enforces basic invariants.
//
// It defensively copies the items slice so callers cannot mutate internal state.
func NewOrder(id OrderID, items []OrderItem, couponCode *string) (*Order, error) {
	if id == "" {
		return nil, ErrInvalidOrderID
	}
	if len(items) == 0 {
		return nil, ErrEmptyOrderItems
	}

	for i, item := range items {
		if item.ProductID == "" {
			return nil, fmt.Errorf("item[%d]: %w", i, ErrInvalidProductID)
		}
		if item.Quantity < 1 {
			return nil, fmt.Errorf("item[%d]: %w", i, ErrInvalidQuantity)
		}
	}

	if couponCode != nil && *couponCode == "" {
		return nil, ErrInvalidCouponCode
	}

	itemsCopy := make([]OrderItem, len(items))
	copy(itemsCopy, items)

	return &Order{
		ID:         id,
		Items:      itemsCopy,
		CouponCode: couponCode,
	}, nil
}

// ProductRepository is the hexagonal port for accessing products.
//
// Adapters (in-memory, DB, etc.) live outside domain and implement this.
type ProductRepository interface {
	ListProducts(ctx context.Context) ([]Product, error)

	// GetProductByID returns the product base on given ID
	//
	// domain.ErrProductNotFound should be returned when no product can be found
	// based on the given ID
	GetProductByID(ctx context.Context, id ProductID) (*Product, error)
	GetProductByIDs(ctx context.Context, ids []ProductID) (map[ProductID]Product, error)
}

type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
}
