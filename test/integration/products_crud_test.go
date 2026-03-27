//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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
	if listBody.Page != 1 || listBody.Limit != 20 || listBody.Total != 1 || listBody.TotalPages != 1 {
		t.Fatalf("unexpected pagination metadata: %#v", listBody)
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

func TestListProductsSupportsPagination(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for i, sku := range []string{"SKU-100", "SKU-101", "SKU-102"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{
			"name":"Desk `+strconv.Itoa(i+1)+`",
			"sku":"`+sku+`",
			"price":129.99,
			"status":"draft"
		}`))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/products?page=2&limit=2", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body listProductsResponse
	decodeResponse(t, rec, &body)

	if len(body.Items) != 1 {
		t.Fatalf("expected 1 product on second page, got %d", len(body.Items))
	}
	if body.Page != 2 || body.Limit != 2 || body.Total != 3 || body.TotalPages != 2 {
		t.Fatalf("unexpected pagination metadata: %#v", body)
	}
	if body.Items[0].SKU != "SKU-102" {
		t.Fatalf("expected SKU-102, got %#v", body.Items[0])
	}
}

func TestListProductsRejectsInvalidPagination(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/products?page=0&limit=101", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload apierror.Envelope
	decodeResponse(t, rec, &payload)

	if len(payload.Error.Details) != 2 {
		t.Fatalf("expected 2 validation details, got %d", len(payload.Error.Details))
	}
}

func TestSearchProductsReturnsDeterministicMatches(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for _, body := range []string{
		`{"name":"beta chair","sku":"CHR-200","price":79.99,"status":"active"}`,
		`{"name":"Alpha Desk","sku":"SKU-300","price":129.99,"status":"active"}`,
		`{"name":"alpha lamp","sku":"LMP-100","price":39.99,"status":"draft"}`,
		`{"name":"Monitor","sku":"SKU-100","price":219.99,"status":"active"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/search/products?q=sku", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body searchProductsResponse
	decodeResponse(t, rec, &body)

	if len(body.Items) != 2 {
		t.Fatalf("expected 2 matching products, got %d", len(body.Items))
	}

	if body.Items[0].Name != "Alpha Desk" || body.Items[0].SKU != "SKU-300" {
		t.Fatalf("unexpected first search result: %#v", body.Items[0])
	}
	if body.Items[1].Name != "Monitor" || body.Items[1].SKU != "SKU-100" {
		t.Fatalf("unexpected second search result: %#v", body.Items[1])
	}
}

func TestSearchProductsReturnsEmptyItemsWhenNoMatches(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/search/products?q=missing", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body searchProductsResponse
	decodeResponse(t, rec, &body)

	if len(body.Items) != 0 {
		t.Fatalf("expected no search results, got %d", len(body.Items))
	}
}

func TestSearchProductsRequiresQuery(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for _, path := range []string{"/v1/search/products", "/v1/search/products?q=%20%20%20"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d for %s", http.StatusBadRequest, rec.Code, path)
		}

		assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "VALIDATION_ERROR", "Request validation failed")

		var payload apierror.Envelope
		decodeResponse(t, rec, &payload)

		if len(payload.Error.Details) != 1 || payload.Error.Details[0].Field != "q" {
			t.Fatalf("expected q validation detail for %s, got %#v", path, payload.Error.Details)
		}
	}
}

func TestProductWriteRoutesReturnRateLimitExceeded(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouterWithRateLimit(t, 1, time.Minute))

	firstReq := httptest.NewRequest(http.MethodDelete, "/v1/products/999", nil)
	firstRec := httptest.NewRecorder()
	router.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, firstRec.Code)
	}

	secondReq := httptest.NewRequest(http.MethodDelete, "/v1/products/999", nil)
	secondRec := httptest.NewRecorder()
	router.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, secondRec.Code)
	}

	assertIntegrationErrorEnvelope(t, secondRec, secondReq.URL.Path, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")

	readReq := httptest.NewRequest(http.MethodGet, "/v1/products/999", nil)
	readRec := httptest.NewRecorder()
	router.ServeHTTP(readRec, readReq)

	if readRec.Code != http.StatusNotFound {
		t.Fatalf("expected read status %d, got %d", http.StatusNotFound, readRec.Code)
	}
}

type listProductsResponse struct {
	Items      []productResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"totalPages"`
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

type searchProductsResponse struct {
	Items []productResponse `json:"items"`
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
