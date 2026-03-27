package categories

import "time"

type Category struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Description *string   `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type CreateInput struct {
	Name        string
	Slug        string
	Description *string
}

type OptionalString struct {
	Set   bool
	Value *string
}

type UpdateInput struct {
	Name        *string
	Slug        *string
	Description OptionalString
}

func (u UpdateInput) HasUpdates() bool {
	return u.Name != nil || u.Slug != nil || u.Description.Set
}
