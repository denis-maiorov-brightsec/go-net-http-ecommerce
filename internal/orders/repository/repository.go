package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
)

var ErrNotFound = errors.New("order not found")

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, input orders.ListInput) (orders.ListResult, error) {
	if r.db == nil {
		return orders.ListResult{}, fmt.Errorf("orders repository is not configured")
	}

	whereClause, args := buildFilters(input)

	countQuery := `SELECT COUNT(*) FROM orders ` + whereClause
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return orders.ListResult{}, fmt.Errorf("count orders: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, input.Limit, (input.Page-1)*input.Limit)
	query := fmt.Sprintf(`
		SELECT id, status, customer_id, created_at, updated_at, total_amount
		FROM orders
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(args)+1, len(args)+2)

	var items []orders.Order
	if err := pgxscan.Select(ctx, r.db, &items, query, queryArgs...); err != nil {
		return orders.ListResult{}, fmt.Errorf("list orders: %w", err)
	}

	if err := r.loadItems(ctx, items); err != nil {
		return orders.ListResult{}, err
	}

	return orders.ListResult{
		Items: items,
		Total: total,
	}, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (orders.Order, error) {
	if r.db == nil {
		return orders.Order{}, fmt.Errorf("orders repository is not configured")
	}

	var item orders.Order
	if err := pgxscan.Get(ctx, r.db, &item, `
		SELECT id, status, customer_id, created_at, updated_at, total_amount
		FROM orders
		WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return orders.Order{}, ErrNotFound
		}

		return orders.Order{}, fmt.Errorf("get order by id: %w", err)
	}

	itemsByOrderID, err := r.listItemsByOrderIDs(ctx, []int64{id})
	if err != nil {
		return orders.Order{}, err
	}
	item.Items = itemsByOrderID[id]

	return item, nil
}

func (r *Repository) loadItems(ctx context.Context, list []orders.Order) error {
	if len(list) == 0 {
		return nil
	}

	orderIDs := make([]int64, 0, len(list))
	for _, item := range list {
		orderIDs = append(orderIDs, item.ID)
	}

	itemsByOrderID, err := r.listItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return err
	}

	for i := range list {
		list[i].Items = itemsByOrderID[list[i].ID]
	}

	return nil
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

func buildFilters(input orders.ListInput) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)

	if input.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)+1))
		args = append(args, input.Status)
	}
	if input.From != nil {
		clauses = append(clauses, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, *input.From)
	}
	if input.To != nil {
		clauses = append(clauses, fmt.Sprintf("created_at < $%d", len(args)+1))
		args = append(args, *input.To)
	}

	if len(clauses) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}
