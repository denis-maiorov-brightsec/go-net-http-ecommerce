package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions"
)

type promotionResponse struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discountType"`
	DiscountValue float64    `json:"discountValue"`
	StartsAt      *time.Time `json:"startsAt,omitempty"`
	EndsAt        *time.Time `json:"endsAt,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type listPromotionsResponse struct {
	Items []promotionResponse `json:"items"`
}

type createPromotionRequest struct {
	Name          string     `json:"name"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discountType"`
	DiscountValue float64    `json:"discountValue"`
	StartsAt      *time.Time `json:"startsAt"`
	EndsAt        *time.Time `json:"endsAt"`
	Status        string     `json:"status"`
}

type updatePromotionRequest struct {
	Name          *string      `json:"name"`
	Code          *string      `json:"code"`
	DiscountType  *string      `json:"discountType"`
	DiscountValue *float64     `json:"discountValue"`
	StartsAt      optionalTime `json:"startsAt"`
	EndsAt        optionalTime `json:"endsAt"`
	Status        *string      `json:"status"`
}

type optionalTime struct {
	Set   bool
	Value *time.Time
}

func (o *optionalTime) UnmarshalJSON(data []byte) error {
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

func newPromotionResponse(promotion promotions.Promotion) promotionResponse {
	return promotionResponse{
		ID:            promotion.ID,
		Name:          promotion.Name,
		Code:          promotion.Code,
		DiscountType:  promotion.DiscountType,
		DiscountValue: promotion.DiscountValue,
		StartsAt:      promotion.StartsAt,
		EndsAt:        promotion.EndsAt,
		Status:        promotion.Status,
		CreatedAt:     promotion.CreatedAt,
		UpdatedAt:     promotion.UpdatedAt,
	}
}

func writeJSON(w writer, status int, payload any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return fmt.Errorf("encode json response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(body.Bytes())
	return err
}
