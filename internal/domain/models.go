package domain

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrEmptyOrderItems   = errors.New("order must contain at least one item")
	ErrInvalidQuantity   = errors.New("item quantity must be >=1 ")
	ErrInvalidProductID  = errors.New("product ID must be non-empty")
	ErrInvalidOrderID    = errors.New("order ID must be non-empty")
	ErrInvalidCouponCode = errors.New("coupon code cannot be empty string") // if present
)

// Strongly typed IDs for clarity
type (
	ProductID string
	OrderID   string
)

// Money - stored as integer cents to avoid float issues
type Money int64

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

// OrderItem represents a product + quantity within an order
type OrderItem struct {
	ProductID ProductID
	Quantity  int
}

type Order struct {
	ID         OrderID
	Items      []OrderItem
	CouponCode *string // optional
}

// NewOrder builds a valid Order an enforces basic invariants
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

	return &Order{
		ID:         id,
		Items:      items,
		CouponCode: couponCode,
	}, nil
}
