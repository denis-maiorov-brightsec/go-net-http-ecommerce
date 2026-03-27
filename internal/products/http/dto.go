package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

type productResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	Price      float64   `json:"price"`
	Status     string    `json:"status"`
	CategoryID *int64    `json:"categoryId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type listProductsResponse struct {
	Items []productResponse `json:"items"`
}

type createProductRequest struct {
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CategoryID *int64  `json:"categoryId"`
}

type updateProductRequest struct {
	Name       *string       `json:"name"`
	SKU        *string       `json:"sku"`
	Price      *float64      `json:"price"`
	Status     *string       `json:"status"`
	CategoryID optionalInt64 `json:"categoryId"`
}

type optionalInt64 struct {
	Set   bool
	Value *int64
}

func (o *optionalInt64) UnmarshalJSON(data []byte) error {
	o.Set = true

	if bytes.Equal(data, []byte("null")) {
		o.Value = nil
		return nil
	}

	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.Value = &value
	return nil
}

func newProductResponse(product products.Product) productResponse {
	return productResponse{
		ID:         product.ID,
		Name:       product.Name,
		SKU:        product.SKU,
		Price:      product.Price,
		Status:     product.Status,
		CategoryID: product.CategoryID,
		CreatedAt:  product.CreatedAt,
		UpdatedAt:  product.UpdatedAt,
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
