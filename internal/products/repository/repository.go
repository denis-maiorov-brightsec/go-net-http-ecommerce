package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

var ErrNotFound = errors.New("product not found")

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, input products.ListInput) (products.ListResult, error) {
	if r.db == nil {
		return products.ListResult{}, fmt.Errorf("products repository is not configured")
	}

	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return products.ListResult{}, fmt.Errorf("count products: %w", err)
	}

	var items []products.Product
	if err := pgxscan.Select(ctx, r.db, &items, `
		SELECT id, name, sku, price, status, category_id, created_at, updated_at
		FROM products
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`, input.Limit, (input.Page-1)*input.Limit); err != nil {
		return products.ListResult{}, fmt.Errorf("list products: %w", err)
	}

	return products.ListResult{
		Items: items,
		Total: total,
	}, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (products.Product, error) {
	if r.db == nil {
		return products.Product{}, fmt.Errorf("products repository is not configured")
	}

	var item products.Product
	if err := pgxscan.Get(ctx, r.db, &item, `
		SELECT id, name, sku, price, status, category_id, created_at, updated_at
		FROM products
		WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return products.Product{}, ErrNotFound
		}

		return products.Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return item, nil
}

func (r *Repository) Create(ctx context.Context, input products.CreateInput) (products.Product, error) {
	if r.db == nil {
		return products.Product{}, fmt.Errorf("products repository is not configured")
	}

	var item products.Product
	if err := pgxscan.Get(ctx, r.db, &item, `
		INSERT INTO products (name, sku, price, status, category_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, sku, price, status, category_id, created_at, updated_at
	`, input.Name, input.SKU, input.Price, input.Status, input.CategoryID); err != nil {
		return products.Product{}, fmt.Errorf("create product: %w", err)
	}

	return item, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input products.UpdateInput) (products.Product, error) {
	if r.db == nil {
		return products.Product{}, fmt.Errorf("products repository is not configured")
	}

	assignments := make([]string, 0, 5)
	args := make([]any, 0, 6)

	if input.Name != nil {
		assignments = append(assignments, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, *input.Name)
	}
	if input.SKU != nil {
		assignments = append(assignments, fmt.Sprintf("sku = $%d", len(args)+1))
		args = append(args, *input.SKU)
	}
	if input.Price != nil {
		assignments = append(assignments, fmt.Sprintf("price = $%d", len(args)+1))
		args = append(args, *input.Price)
	}
	if input.Status != nil {
		assignments = append(assignments, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *input.Status)
	}
	if input.CategoryID.Set {
		assignments = append(assignments, fmt.Sprintf("category_id = $%d", len(args)+1))
		args = append(args, input.CategoryID.Value)
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, sku, price, status, category_id, created_at, updated_at
	`, strings.Join(assignments, ", "), len(args)+1)
	args = append(args, id)

	var item products.Product
	if err := pgxscan.Get(ctx, r.db, &item, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return products.Product{}, ErrNotFound
		}

		return products.Product{}, fmt.Errorf("update product: %w", err)
	}

	return item, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if r.db == nil {
		return fmt.Errorf("products repository is not configured")
	}

	result, err := r.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
