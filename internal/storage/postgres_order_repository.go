package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
)

type PgOrderRepository struct {
	db *sql.DB
}

func NewPgOrderRepository(db *sql.DB) domain.OrderRepository {
	return &PgOrderRepository{
		db: db,
	}
}

func (r *PgOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx for save order: %w", err)
	}
	defer func() {
		// If commit wasn't called, rollback will be a no-op after Commit.
		_ = tx.Rollback()
	}()

	const insertOrder = `
		INSERT INTO orders (id, coupon_code)
		VALUES ($1, $2)
	`

	var couponCode *string
	if order.CouponCode != nil && *order.CouponCode != "" {
		couponCode = order.CouponCode
	}

	if _, err := tx.ExecContext(ctx, insertOrder, string(order.ID), couponCode); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	const insertItem = `
		INSERT INTO order_items(order_id, product_id, quantity)
		VALUES ($1, $2, $3)
	`

	for _, item := range order.Items {
		if _, err := tx.ExecContext(ctx, insertItem, string(order.ID), string(item.ProductID), item.Quantity); err != nil {
			return fmt.Errorf("insert order item (order_id=%s, product_id=%s): %w", order.ID, item.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx for save order: %w", err)
	}

	return nil
}
