package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
)

var ErrNotFound = errors.New("order not found")
var ErrIneligibleStatus = errors.New("order status does not allow cancellation")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Cancel(ctx context.Context, id int64) (orders.Order, error) {
	if r.db == nil {
		return orders.Order{}, fmt.Errorf("orders command repository is not configured")
	}

	var item orders.Order
	if err := pgxscan.Get(ctx, r.db, &item, `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1 AND status = $3
		RETURNING id, status, customer_id, created_at, updated_at, total_amount
	`, id, orders.StatusCancelled, orders.StatusPending); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orders.Order{}, r.cancelFailure(ctx, id)
		}

		return orders.Order{}, fmt.Errorf("cancel order: %w", err)
	}

	itemsByOrderID, err := r.listItemsByOrderIDs(ctx, []int64{id})
	if err != nil {
		return orders.Order{}, err
	}
	item.Items = itemsByOrderID[id]

	return item, nil
}

func (r *Repository) listItemsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64][]orders.Item, error) {
	var rows []orders.Item
	if err := pgxscan.Select(ctx, r.db, &rows, `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, total_amount
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY order_id ASC, id ASC
	`, orderIDs); err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}

	itemsByOrderID := make(map[int64][]orders.Item, len(orderIDs))
	for _, item := range rows {
		itemsByOrderID[item.OrderID] = append(itemsByOrderID[item.OrderID], item)
	}

	return itemsByOrderID, nil
}

func (r *Repository) cancelFailure(ctx context.Context, id int64) error {
	var status string
	if err := r.db.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, id).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}

		return fmt.Errorf("lookup order status for cancellation: %w", err)
	}

	return ErrIneligibleStatus
}
