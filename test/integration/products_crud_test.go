//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

func TestProductsCRUDFlow(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{
		"name":"Desk",
		"sku":"SKU-100",
		"price":129.99,
		"status":"draft"
	}`))
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created productResponse
	decodeResponse(t, createRec, &created)

	if created.ID == 0 {
		t.Fatal("expected created product id to be set")
	}
	if created.Name != "Desk" || created.SKU != "SKU-100" || created.Status != "draft" {
		t.Fatalf("unexpected created product: %#v", created)
	}
	if created.Price != 129.99 {
		t.Fatalf("expected price 129.99, got %v", created.Price)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("expected timestamps to be set, got %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/products", nil)
	listRec := httptest.NewRecorder()

	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var listBody listProductsResponse
	decodeResponse(t, listRec, &listBody)

	if len(listBody.Items) != 1 {
		t.Fatalf("expected 1 product, got %d", len(listBody.Items))
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/products/"+int64Path(created.ID), nil)
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var fetched productResponse
	decodeResponse(t, getRec, &fetched)

	if fetched.ID != created.ID {
		t.Fatalf("expected id %d, got %d", created.ID, fetched.ID)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/v1/products/"+int64Path(created.ID), strings.NewReader(`{
		"price":149.99,
		"status":"active",
		"categoryId":42
	}`))
	updateRec := httptest.NewRecorder()

	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRec.Code)
	}

	var updated productResponse
	decodeResponse(t, updateRec, &updated)

	if updated.Price != 149.99 || updated.Status != "active" {
		t.Fatalf("unexpected updated product: %#v", updated)
	}
	if updated.CategoryID == nil || *updated.CategoryID != 42 {
		t.Fatalf("expected categoryId 42, got %#v", updated.CategoryID)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/products/"+int64Path(created.ID), nil)
	deleteRec := httptest.NewRecorder()

	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/v1/products/"+int64Path(created.ID), nil)
	missingRec := httptest.NewRecorder()

	router.ServeHTTP(missingRec, missingReq)

	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, missingRec.Code)
	}

	assertIntegrationErrorEnvelope(t, missingRec, missingReq.URL.Path, "NOT_FOUND", "Product not found")
}

func TestCreateProductValidationErrors(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{
		"name":"",
		"sku":"",
		"price":0,
		"status":""
	}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload apierror.Envelope
	decodeResponse(t, rec, &payload)

	if len(payload.Error.Details) != 4 {
		t.Fatalf("expected 4 validation details, got %d", len(payload.Error.Details))
	}
}

func TestDeleteMissingProductReturnsNotFound(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodDelete, "/v1/products/999", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "NOT_FOUND", "Product not found")
}

func TestPatchProductValidationErrors(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodPatch, "/v1/products/1", strings.NewReader(`{"price":-10}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "VALIDATION_ERROR", "Request validation failed")
}

type listProductsResponse struct {
	Items []productResponse `json:"items"`
}

type productResponse struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
	CategoryID *int64  `json:"categoryId"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func int64Path(value int64) string {
	return strconv.FormatInt(value, 10)
}
