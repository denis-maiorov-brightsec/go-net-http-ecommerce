package dto

import (
	"bytes"
	"encoding/json"
	"time"
)

type CreatePromotionRequest struct {
	Name          string     `json:"name"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discountType"`
	DiscountValue float64    `json:"discountValue"`
	StartsAt      *time.Time `json:"startsAt"`
	EndsAt        *time.Time `json:"endsAt"`
	Status        string     `json:"status"`
}

type UpdatePromotionRequest struct {
	Name          *string      `json:"name"`
	Code          *string      `json:"code"`
	DiscountType  *string      `json:"discountType"`
	DiscountValue *float64     `json:"discountValue"`
	StartsAt      OptionalTime `json:"startsAt"`
	EndsAt        OptionalTime `json:"endsAt"`
	Status        *string      `json:"status"`
}

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

func (o *OptionalTime) UnmarshalJSON(data []byte) error {
	o.Set = true

	if bytes.Equal(data, []byte("null")) {
		o.Value = nil
		return nil
	}

	var value time.Time
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}
