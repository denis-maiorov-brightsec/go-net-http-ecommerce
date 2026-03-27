package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/pagination"
)

type orderResponse struct {
	ID          int64               `json:"id"`
	Status      string              `json:"status"`
	CustomerID  int64               `json:"customerId"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	Items       []orderItemResponse `json:"items"`
	TotalAmount float64             `json:"totalAmount"`
}

type orderItemResponse struct {
	ID          int64   `json:"id"`
	ProductID   *int64  `json:"productId,omitempty"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalAmount float64 `json:"totalAmount"`
}

type listOrdersResponse struct {
	Items []orderResponse `json:"items"`
	pagination.Metadata
}

func newOrderResponse(order orders.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, orderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalAmount: item.TotalAmount,
		})
	}

	return orderResponse{
		ID:          order.ID,
		Status:      order.Status,
		CustomerID:  order.CustomerID,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
		Items:       items,
		TotalAmount: order.TotalAmount,
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
