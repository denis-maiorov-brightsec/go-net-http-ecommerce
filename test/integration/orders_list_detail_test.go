//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOrdersListAndDetailFlow(t *testing.T) {
	routerDeps := newIntegrationRouter(t)
	firstOrderID := insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "pending",
		CustomerID: 101,
		CreatedAt:  time.Date(2026, time.January, 10, 9, 0, 0, 0, time.UTC),
		Items: []orderItemFixture{
			{ProductName: "Desk", Quantity: 1, UnitPrice: 129.99, TotalAmount: 129.99},
			{ProductName: "Lamp", Quantity: 2, UnitPrice: 25.00, TotalAmount: 50.00},
		},
	})
	insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "shipped",
		CustomerID: 202,
		CreatedAt:  time.Date(2026, time.January, 12, 13, 30, 0, 0, time.UTC),
		Items: []orderItemFixture{
			{ProductName: "Chair", Quantity: 4, UnitPrice: 49.50, TotalAmount: 198.00},
		},
	})

	router := api.NewRouter(*routerDeps)

	listReq := httptest.NewRequest(http.MethodGet, "/v1/orders", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var listBody listOrdersResponse
	decodeOrderResponse(t, listRec, &listBody)

	if len(listBody.Items) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(listBody.Items))
	}
	if listBody.Page != 1 || listBody.Limit != 20 || listBody.Total != 2 || listBody.TotalPages != 1 {
		t.Fatalf("unexpected pagination metadata: %#v", listBody)
	}
	if listBody.Items[0].Status != "shipped" {
		t.Fatalf("expected newest order first, got %#v", listBody.Items[0])
	}
	if len(listBody.Items[1].Items) != 2 {
		t.Fatalf("expected 2 line items, got %#v", listBody.Items[1].Items)
	}
	if listBody.Items[1].TotalAmount != 179.99 {
		t.Fatalf("expected totalAmount 179.99, got %v", listBody.Items[1].TotalAmount)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/orders/"+int64Path(firstOrderID), nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var fetched orderResponse
	decodeOrderResponse(t, getRec, &fetched)

	if fetched.ID != firstOrderID {
		t.Fatalf("expected id %d, got %d", firstOrderID, fetched.ID)
	}
	if fetched.CustomerID != 101 || len(fetched.Items) != 2 {
		t.Fatalf("unexpected order detail: %#v", fetched)
	}
}

func TestOrdersListSupportsFiltersAndPagination(t *testing.T) {
	routerDeps := newIntegrationRouter(t)
	insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "pending",
		CustomerID: 301,
		CreatedAt:  time.Date(2026, time.February, 1, 10, 0, 0, 0, time.UTC),
		Items:      []orderItemFixture{{ProductName: "Desk", Quantity: 1, UnitPrice: 100, TotalAmount: 100}},
	})
	insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "pending",
		CustomerID: 302,
		CreatedAt:  time.Date(2026, time.February, 2, 10, 0, 0, 0, time.UTC),
		Items:      []orderItemFixture{{ProductName: "Lamp", Quantity: 1, UnitPrice: 30, TotalAmount: 30}},
	})
	insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "shipped",
		CustomerID: 303,
		CreatedAt:  time.Date(2026, time.February, 3, 10, 0, 0, 0, time.UTC),
		Items:      []orderItemFixture{{ProductName: "Chair", Quantity: 1, UnitPrice: 60, TotalAmount: 60}},
	})

	router := api.NewRouter(*routerDeps)

	req := httptest.NewRequest(http.MethodGet, "/v1/orders?status=pending&from=2026-02-01&to=2026-02-02&page=2&limit=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body listOrdersResponse
	decodeOrderResponse(t, rec, &body)

	if len(body.Items) != 1 {
		t.Fatalf("expected 1 order on second page, got %d", len(body.Items))
	}
	if body.Page != 2 || body.Limit != 1 || body.Total != 2 || body.TotalPages != 2 {
		t.Fatalf("unexpected pagination metadata: %#v", body)
	}
	if body.Items[0].CustomerID != 301 {
		t.Fatalf("expected filtered second page order for customer 301, got %#v", body.Items[0])
	}
}

func TestOrderDetailReturnsNotFoundWhenMissing(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/orders/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "NOT_FOUND", "Order not found")
}

func TestOrdersListRejectsInvalidDateFilters(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/orders?from=2026-02-30&to=bad-date", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload apierror.Envelope
	decodeOrderResponse(t, rec, &payload)

	if len(payload.Error.Details) != 2 {
		t.Fatalf("expected 2 validation details, got %d", len(payload.Error.Details))
	}
}

func TestOrderCancelReturnsUpdatedOrderWhenPending(t *testing.T) {
	routerDeps := newIntegrationRouter(t)
	orderID := insertOrder(t, routerDeps.DB, orderFixture{
		Status:     orders.StatusPending,
		CustomerID: 501,
		CreatedAt:  time.Date(2026, time.March, 1, 11, 0, 0, 0, time.UTC),
		Items: []orderItemFixture{
			{ProductName: "Monitor", Quantity: 1, UnitPrice: 220, TotalAmount: 220},
		},
	})

	router := api.NewRouter(*routerDeps)

	req := httptest.NewRequest(http.MethodPost, "/v1/orders/"+int64Path(orderID)+"/cancel", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body orderResponse
	decodeOrderResponse(t, rec, &body)

	if body.ID != orderID {
		t.Fatalf("expected id %d, got %d", orderID, body.ID)
	}
	if body.Status != orders.StatusCancelled {
		t.Fatalf("expected status %q, got %q", orders.StatusCancelled, body.Status)
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 order item, got %#v", body.Items)
	}

	var persistedStatus string
	if err := routerDeps.DB.QueryRow(context.Background(), `SELECT status FROM orders WHERE id = $1`, orderID).Scan(&persistedStatus); err != nil {
		t.Fatalf("load cancelled order: %v", err)
	}
	if persistedStatus != orders.StatusCancelled {
		t.Fatalf("expected persisted status %q, got %q", orders.StatusCancelled, persistedStatus)
	}
}

func TestOrderCancelReturnsNotFoundWhenMissing(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodPost, "/v1/orders/999/cancel", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "NOT_FOUND", "Order not found")
}

func TestOrderCancelReturnsConflictForIneligibleStatus(t *testing.T) {
	routerDeps := newIntegrationRouter(t)
	orderID := insertOrder(t, routerDeps.DB, orderFixture{
		Status:     "shipped",
		CustomerID: 502,
		CreatedAt:  time.Date(2026, time.March, 2, 12, 0, 0, 0, time.UTC),
		Items: []orderItemFixture{
			{ProductName: "Keyboard", Quantity: 1, UnitPrice: 80, TotalAmount: 80},
		},
	})

	router := api.NewRouter(*routerDeps)

	req := httptest.NewRequest(http.MethodPost, "/v1/orders/"+int64Path(orderID)+"/cancel", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "CONFLICT", "Order cannot be cancelled from current status")

	var persistedStatus string
	if err := routerDeps.DB.QueryRow(context.Background(), `SELECT status FROM orders WHERE id = $1`, orderID).Scan(&persistedStatus); err != nil {
		t.Fatalf("load uncancelled order: %v", err)
	}
	if persistedStatus != "shipped" {
		t.Fatalf("expected persisted status %q, got %q", "shipped", persistedStatus)
	}
}

type listOrdersResponse struct {
	Items      []orderResponse `json:"items"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	Total      int64           `json:"total"`
	TotalPages int             `json:"totalPages"`
}

type orderResponse struct {
	ID          int64               `json:"id"`
	Status      string              `json:"status"`
	CustomerID  int64               `json:"customerId"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
	Items       []orderItemResponse `json:"items"`
	TotalAmount float64             `json:"totalAmount"`
}

type orderItemResponse struct {
	ID          int64   `json:"id"`
	ProductID   *int64  `json:"productId"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	TotalAmount float64 `json:"totalAmount"`
}

type orderFixture struct {
	Status     string
	CustomerID int64
	CreatedAt  time.Time
	Items      []orderItemFixture
}

type orderItemFixture struct {
	ProductID   *int64
	ProductName string
	Quantity    int
	UnitPrice   float64
	TotalAmount float64
}

func insertOrder(t *testing.T, db *pgxpool.Pool, fixture orderFixture) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	totalAmount := 0.0
	for _, item := range fixture.Items {
		totalAmount += item.TotalAmount
	}

	var orderID int64
	if err := db.QueryRow(ctx, `
		INSERT INTO orders (status, customer_id, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id
	`, fixture.Status, fixture.CustomerID, totalAmount, fixture.CreatedAt).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}

	for _, item := range fixture.Items {
		if _, err := db.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, total_amount)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, orderID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice, item.TotalAmount); err != nil {
			t.Fatalf("insert order item: %v", err)
		}
	}

	return orderID
}

func decodeOrderResponse(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
