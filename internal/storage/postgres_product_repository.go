package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/M-Arthur/order-food-api/internal/domain"
	"github.com/lib/pq"
)

type PgProductRepository struct {
	db *sql.DB
}

func NewPgProductRepository(db *sql.DB) domain.ProductRepository {
	return &PgProductRepository{
		db: db,
	}
}

func (r *PgProductRepository) ListProducts(ctx context.Context) ([]domain.Product, error) {
	const query = `
		SELECT id, name, price_cents, category
		FROM products
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var products []domain.Product
	for rows.Next() {
		var (
			id         string
			name       string
			priceCents int64
			category   string
		)

		if err := rows.Scan(&id, &name, &priceCents, &category); err != nil {
			return nil, fmt.Errorf("scan product row: %w", err)
		}

		products = append(products, domain.Product{
			ID:       domain.ProductID(id),
			Name:     name,
			Price:    domain.Money(priceCents),
			Category: category,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product rows: %w", err)
	}

	return products, nil
}

func (r *PgProductRepository) GetProductByID(ctx context.Context, id domain.ProductID) (*domain.Product, error) {
	const query = `
		SELECT id, name, price_cents, category
		FROM products
		WHERE id = $1
	`

	var (
		rawID      string
		name       string
		priceCents int64
		category   string
	)

	err := r.db.QueryRowContext(ctx, query, string(id)).
		Scan(&rawID, &name, &priceCents, &category)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get product by id %s: %w", id, err)
	}

	p := &domain.Product{
		ID:       domain.ProductID(rawID),
		Name:     name,
		Price:    domain.Money(priceCents),
		Category: category,
	}
	return p, nil
}

func (r *PgProductRepository) GetProductByIDs(ctx context.Context, ids []domain.ProductID) (map[domain.ProductID]domain.Product, error) {
	if len(ids) == 0 {
		return map[domain.ProductID]domain.Product{}, nil
	}

	idStrings := make([]string, 0, len(ids))
	for _, id := range ids {
		idStrings = append(idStrings, string(id))
	}

	const query = `
		SELECT id, name, price_cents, categroy
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(idStrings))
	if err != nil {
		return nil, fmt.Errorf("get products by ids: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	result := make(map[domain.ProductID]domain.Product, len(ids))

	for rows.Next() {
		var (
			rawID      string
			name       string
			priceCents int64
			category   string
		)

		if err := rows.Scan(&rawID, &name, &priceCents, &category); err != nil {
			return nil, fmt.Errorf("scan product row: %w", err)
		}

		p := domain.Product{
			ID:       domain.ProductID(rawID),
			Name:     name,
			Price:    domain.Money(priceCents),
			Category: category,
		}
		result[p.ID] = p
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product row %w", err)
	}

	return result, nil
}
