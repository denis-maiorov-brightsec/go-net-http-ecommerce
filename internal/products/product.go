package products

import "time"

type Product struct {
	ID         int64     `db:"id"`
	Name       string    `db:"name"`
	SKU        string    `db:"sku"`
	Price      float64   `db:"price"`
	Status     string    `db:"status"`
	CategoryID *int64    `db:"category_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type ListInput struct {
	Page  int
	Limit int
}

type ListResult struct {
	Items []Product
	Total int64
}

type SearchInput struct {
	Query string
}

type CreateInput struct {
	Name       string
	SKU        string
	Price      float64
	Status     string
	CategoryID *int64
}

type OptionalInt64 struct {
	Set   bool
	Value *int64
}

type UpdateInput struct {
	Name       *string
	SKU        *string
	Price      *float64
	Status     *string
	CategoryID OptionalInt64
}

func (u UpdateInput) HasUpdates() bool {
	return u.Name != nil || u.SKU != nil || u.Price != nil || u.Status != nil || u.CategoryID.Set
}
