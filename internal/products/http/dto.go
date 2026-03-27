package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/pagination"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

type productResponse struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	StockKeepingUnit string    `json:"stockKeepingUnit"`
	Price            float64   `json:"price"`
	Status           string    `json:"status"`
	CategoryID       *int64    `json:"categoryId,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type listProductsResponse struct {
	Items []productResponse `json:"items"`
	pagination.Metadata
}

type searchProductsResponse struct {
	Items []productResponse `json:"items"`
}

func newProductResponse(product products.Product) productResponse {
	return productResponse{
		ID:               product.ID,
		Name:             product.Name,
		StockKeepingUnit: product.StockKeepingUnit,
		Price:            product.Price,
		Status:           product.Status,
		CategoryID:       product.CategoryID,
		CreatedAt:        product.CreatedAt,
		UpdatedAt:        product.UpdatedAt,
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
