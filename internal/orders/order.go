package orders

import "time"

type Order struct {
	ID          int64     `db:"id"`
	Status      string    `db:"status"`
	CustomerID  int64     `db:"customer_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Items       []Item
	TotalAmount float64 `db:"total_amount"`
}

type Item struct {
	ID          int64   `db:"id"`
	OrderID     int64   `db:"order_id"`
	ProductID   *int64  `db:"product_id"`
	ProductName string  `db:"product_name"`
	Quantity    int     `db:"quantity"`
	UnitPrice   float64 `db:"unit_price"`
	TotalAmount float64 `db:"total_amount"`
}

type ListInput struct {
	Page   int
	Limit  int
	Status string
	From   *time.Time
	To     *time.Time
}

type ListResult struct {
	Items []Order
	Total int64
}
