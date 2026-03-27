package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories"
)

var (
	ErrNotFound      = errors.New("category not found")
	ErrDuplicateSlug = errors.New("category slug already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]categories.Category, error) {
	if r.db == nil {
		return nil, fmt.Errorf("categories repository is not configured")
	}

	var items []categories.Category
	if err := pgxscan.Select(ctx, r.db, &items, `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		ORDER BY id ASC
	`); err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (categories.Category, error) {
	if r.db == nil {
		return categories.Category{}, fmt.Errorf("categories repository is not configured")
	}

	var item categories.Category
	if err := pgxscan.Get(ctx, r.db, &item, `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return categories.Category{}, ErrNotFound
		}

		return categories.Category{}, fmt.Errorf("get category by id: %w", err)
	}

	return item, nil
}

func (r *Repository) Create(ctx context.Context, input categories.CreateInput) (categories.Category, error) {
	if r.db == nil {
		return categories.Category{}, fmt.Errorf("categories repository is not configured")
	}

	var item categories.Category
	if err := pgxscan.Get(ctx, r.db, &item, `
		INSERT INTO categories (name, slug, description)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, description, created_at, updated_at
	`, input.Name, input.Slug, input.Description); err != nil {
		if isUniqueViolation(err) {
			return categories.Category{}, ErrDuplicateSlug
		}

		return categories.Category{}, fmt.Errorf("create category: %w", err)
	}

	return item, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input categories.UpdateInput) (categories.Category, error) {
	if r.db == nil {
		return categories.Category{}, fmt.Errorf("categories repository is not configured")
	}

	assignments := make([]string, 0, 3)
	args := make([]any, 0, 4)

	if input.Name != nil {
		assignments = append(assignments, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, *input.Name)
	}
	if input.Slug != nil {
		assignments = append(assignments, fmt.Sprintf("slug = $%d", len(args)+1))
		args = append(args, *input.Slug)
	}
	if input.Description.Set {
		assignments = append(assignments, fmt.Sprintf("description = $%d", len(args)+1))
		args = append(args, input.Description.Value)
	}

	query := fmt.Sprintf(`
		UPDATE categories
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, slug, description, created_at, updated_at
	`, strings.Join(assignments, ", "), len(args)+1)
	args = append(args, id)

	var item categories.Category
	if err := pgxscan.Get(ctx, r.db, &item, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return categories.Category{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return categories.Category{}, ErrDuplicateSlug
		}

		return categories.Category{}, fmt.Errorf("update category: %w", err)
	}

	return item, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if r.db == nil {
		return fmt.Errorf("categories repository is not configured")
	}

	result, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
