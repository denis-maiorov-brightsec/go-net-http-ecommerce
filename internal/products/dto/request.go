package dto

import (
	"bytes"
	"encoding/json"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

type CreateProductRequest struct {
	Name             string  `json:"name"`
	StockKeepingUnit *string `json:"stockKeepingUnit"`
	// Deprecated: sku remains accepted as an input alias during the spec-013 transition.
	SKU        *string `json:"sku"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CategoryID *int64  `json:"categoryId"`
}

type UpdateProductRequest struct {
	Name             *string `json:"name"`
	StockKeepingUnit *string `json:"stockKeepingUnit"`
	// Deprecated: sku remains accepted as an input alias during the spec-013 transition.
	SKU        *string       `json:"sku"`
	Price      *float64      `json:"price"`
	Status     *string       `json:"status"`
	CategoryID OptionalInt64 `json:"categoryId"`
}

type OptionalInt64 struct {
	Set   bool
	Value *int64
}

func (o *OptionalInt64) UnmarshalJSON(data []byte) error {
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

func (r CreateProductRequest) ToCreateInput() (products.CreateInput, error) {
	stockKeepingUnit, err := resolveStockKeepingUnit(r.StockKeepingUnit, r.SKU)
	if err != nil {
		return products.CreateInput{}, err
	}

	input := products.CreateInput{
		Name:       r.Name,
		Price:      r.Price,
		Status:     r.Status,
		CategoryID: r.CategoryID,
	}
	if stockKeepingUnit != nil {
		input.StockKeepingUnit = *stockKeepingUnit
	}

	return input, nil
}

func (r UpdateProductRequest) ToUpdateInput() (products.UpdateInput, error) {
	stockKeepingUnit, err := resolveStockKeepingUnit(r.StockKeepingUnit, r.SKU)
	if err != nil {
		return products.UpdateInput{}, err
	}

	return products.UpdateInput{
		Name:             r.Name,
		StockKeepingUnit: stockKeepingUnit,
		Price:            r.Price,
		Status:           r.Status,
		CategoryID: products.OptionalInt64{
			Set:   r.CategoryID.Set,
			Value: r.CategoryID.Value,
		},
	}, nil
}

func resolveStockKeepingUnit(stockKeepingUnit, sku *string) (*string, error) {
	if stockKeepingUnit != nil && sku != nil && *stockKeepingUnit != *sku {
		return nil, apierror.Validation([]apierror.Detail{{
			Field:       "stockKeepingUnit",
			Constraints: []string{"stockKeepingUnit and deprecated sku must match when both are provided"},
		}})
	}

	if stockKeepingUnit != nil {
		return stockKeepingUnit, nil
	}

	return sku, nil
}
