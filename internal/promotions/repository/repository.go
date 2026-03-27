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

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions"
)

var (
	ErrNotFound      = errors.New("promotion not found")
	ErrDuplicateCode = errors.New("promotion code already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]promotions.Promotion, error) {
	if r.db == nil {
		return nil, fmt.Errorf("promotions repository is not configured")
	}

	var items []promotions.Promotion
	if err := pgxscan.Select(ctx, r.db, &items, `
		SELECT id, name, code, discount_type, discount_value, starts_at, ends_at, status, created_at, updated_at
		FROM promotions
		ORDER BY id ASC
	`); err != nil {
		return nil, fmt.Errorf("list promotions: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (promotions.Promotion, error) {
	if r.db == nil {
		return promotions.Promotion{}, fmt.Errorf("promotions repository is not configured")
	}

	var item promotions.Promotion
	if err := pgxscan.Get(ctx, r.db, &item, `
		SELECT id, name, code, discount_type, discount_value, starts_at, ends_at, status, created_at, updated_at
		FROM promotions
		WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return promotions.Promotion{}, ErrNotFound
		}

		return promotions.Promotion{}, fmt.Errorf("get promotion by id: %w", err)
	}

	return item, nil
}

func (r *Repository) Create(ctx context.Context, input promotions.CreateInput) (promotions.Promotion, error) {
	if r.db == nil {
		return promotions.Promotion{}, fmt.Errorf("promotions repository is not configured")
	}

	var item promotions.Promotion
	if err := pgxscan.Get(ctx, r.db, &item, `
		INSERT INTO promotions (name, code, discount_type, discount_value, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, code, discount_type, discount_value, starts_at, ends_at, status, created_at, updated_at
	`, input.Name, input.Code, input.DiscountType, input.DiscountValue, input.StartsAt, input.EndsAt, input.Status); err != nil {
		if isUniqueViolation(err) {
			return promotions.Promotion{}, ErrDuplicateCode
		}

		return promotions.Promotion{}, fmt.Errorf("create promotion: %w", err)
	}

	return item, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input promotions.UpdateInput) (promotions.Promotion, error) {
	if r.db == nil {
		return promotions.Promotion{}, fmt.Errorf("promotions repository is not configured")
	}

	assignments := make([]string, 0, 7)
	args := make([]any, 0, 8)

	if input.Name != nil {
		assignments = append(assignments, fmt.Sprintf("name = $%d", len(args)+1))
		args = append(args, *input.Name)
	}
	if input.Code != nil {
		assignments = append(assignments, fmt.Sprintf("code = $%d", len(args)+1))
		args = append(args, *input.Code)
	}
	if input.DiscountType != nil {
		assignments = append(assignments, fmt.Sprintf("discount_type = $%d", len(args)+1))
		args = append(args, *input.DiscountType)
	}
	if input.DiscountValue != nil {
		assignments = append(assignments, fmt.Sprintf("discount_value = $%d", len(args)+1))
		args = append(args, *input.DiscountValue)
	}
	if input.StartsAt.Set {
		assignments = append(assignments, fmt.Sprintf("starts_at = $%d", len(args)+1))
		args = append(args, input.StartsAt.Value)
	}
	if input.EndsAt.Set {
		assignments = append(assignments, fmt.Sprintf("ends_at = $%d", len(args)+1))
		args = append(args, input.EndsAt.Value)
	}
	if input.Status != nil {
		assignments = append(assignments, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, *input.Status)
	}

	query := fmt.Sprintf(`
		UPDATE promotions
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, code, discount_type, discount_value, starts_at, ends_at, status, created_at, updated_at
	`, strings.Join(assignments, ", "), len(args)+1)
	args = append(args, id)

	var item promotions.Promotion
	if err := pgxscan.Get(ctx, r.db, &item, query, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return promotions.Promotion{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return promotions.Promotion{}, ErrDuplicateCode
		}

		return promotions.Promotion{}, fmt.Errorf("update promotion: %w", err)
	}

	return item, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	if r.db == nil {
		return fmt.Errorf("promotions repository is not configured")
	}

	result, err := r.db.Exec(ctx, `DELETE FROM promotions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete promotion: %w", err)
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
