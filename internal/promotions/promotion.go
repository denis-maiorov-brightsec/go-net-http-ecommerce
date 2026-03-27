package promotions

import "time"

type Promotion struct {
	ID            int64      `db:"id"`
	Name          string     `db:"name"`
	Code          string     `db:"code"`
	DiscountType  string     `db:"discount_type"`
	DiscountValue float64    `db:"discount_value"`
	StartsAt      *time.Time `db:"starts_at"`
	EndsAt        *time.Time `db:"ends_at"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

type CreateInput struct {
	Name          string
	Code          string
	DiscountType  string
	DiscountValue float64
	StartsAt      *time.Time
	EndsAt        *time.Time
	Status        string
}

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

type UpdateInput struct {
	Name          *string
	Code          *string
	DiscountType  *string
	DiscountValue *float64
	StartsAt      OptionalTime
	EndsAt        OptionalTime
	Status        *string
}

func (u UpdateInput) HasUpdates() bool {
	return u.Name != nil ||
		u.Code != nil ||
		u.DiscountType != nil ||
		u.DiscountValue != nil ||
		u.StartsAt.Set ||
		u.EndsAt.Set ||
		u.Status != nil
}
