//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
)

func TestProductsRoutesContract_ReadShapes(t *testing.T) {
	routerDeps := newIntegrationRouter(t)
	firstID := seedProductContract(t, routerDeps, seededProductContract{
		Name:             "Desk",
		StockKeepingUnit: "SKU-100",
		Price:            129.99,
		Status:           "draft",
		CreatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
	})
	secondID := seedProductContract(t, routerDeps, seededProductContract{
		Name:             "Lamp",
		StockKeepingUnit: "LMP-200",
		Price:            39.99,
		Status:           "active",
		CreatedAt:        time.Date(2026, time.January, 3, 11, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 4, 11, 0, 0, 0, time.UTC),
	})

	router := api.NewRouter(*routerDeps)

	listReq := httptest.NewRequest(http.MethodGet, "/v1/products", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var listBody map[string]any
	decodeResponse(t, listRec, &listBody)
	assertObjectKeys(t, listBody, "items", "limit", "page", "total", "totalPages")
	assertNumberValue(t, listBody["page"], 1, "page")
	assertNumberValue(t, listBody["limit"], 20, "limit")
	assertNumberValue(t, listBody["total"], 2, "total")
	assertNumberValue(t, listBody["totalPages"], 1, "totalPages")

	listItems := requireObjectSlice(t, listBody["items"], "items")
	if len(listItems) != 2 {
		t.Fatalf("expected 2 products, got %d", len(listItems))
	}

	assertProductContractShape(t, listItems[0], seededProductContract{
		ID:               firstID,
		Name:             "Desk",
		StockKeepingUnit: "SKU-100",
		Price:            129.99,
		Status:           "draft",
		CreatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
	})
	assertProductContractShape(t, listItems[1], seededProductContract{
		ID:               secondID,
		Name:             "Lamp",
		StockKeepingUnit: "LMP-200",
		Price:            39.99,
		Status:           "active",
		CreatedAt:        time.Date(2026, time.January, 3, 11, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 4, 11, 0, 0, 0, time.UTC),
	})

	getReq := httptest.NewRequest(http.MethodGet, "/v1/products/"+int64Path(secondID), nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var getBody map[string]any
	decodeResponse(t, getRec, &getBody)
	assertProductContractShape(t, getBody, seededProductContract{
		ID:               secondID,
		Name:             "Lamp",
		StockKeepingUnit: "LMP-200",
		Price:            39.99,
		Status:           "active",
		CreatedAt:        time.Date(2026, time.January, 3, 11, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 4, 11, 0, 0, 0, time.UTC),
	})

	searchReq := httptest.NewRequest(http.MethodGet, "/v1/search/products?q=sku", nil)
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)

	if searchRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, searchRec.Code)
	}

	var searchBody map[string]any
	decodeResponse(t, searchRec, &searchBody)
	assertObjectKeys(t, searchBody, "items")

	searchItems := requireObjectSlice(t, searchBody["items"], "items")
	if len(searchItems) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchItems))
	}

	assertProductContractShape(t, searchItems[0], seededProductContract{
		ID:               firstID,
		Name:             "Desk",
		StockKeepingUnit: "SKU-100",
		Price:            129.99,
		Status:           "draft",
		CreatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
		UpdatedAt:        time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC),
	})
}

func TestProductsRoutesContract_WriteFlowAndAliasCompatibility(t *testing.T) {
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

	var createBody map[string]any
	decodeResponse(t, createRec, &createBody)
	createdID := assertProductContractShapeWithoutTimestamps(t, createBody, "Desk", "SKU-100", 129.99, "draft")

	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/products/"+int64Path(createdID), strings.NewReader(`{
		"sku":"SKU-101",
		"status":"active"
	}`))
	patchRec := httptest.NewRecorder()
	router.ServeHTTP(patchRec, patchReq)

	if patchRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, patchRec.Code)
	}

	var patchBody map[string]any
	decodeResponse(t, patchRec, &patchBody)
	assertProductContractShapeWithoutTimestamps(t, patchBody, "Desk", "SKU-101", 129.99, "active")

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/products/"+int64Path(createdID), nil)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}
	if body := strings.TrimSpace(deleteRec.Body.String()); body != "" {
		t.Fatalf("expected empty delete body, got %q", body)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/v1/products/"+int64Path(createdID), nil)
	missingRec := httptest.NewRecorder()
	router.ServeHTTP(missingRec, missingReq)

	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, missingRec.Code)
	}

	assertIntegrationErrorEnvelope(t, missingRec, missingReq.URL.Path, "NOT_FOUND", "Product not found")
	assertErrorContractShape(t, missingRec, []string{})
}

func TestProductsRoutesContract_ValidationBehavior(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{
		"name":"",
		"stockKeepingUnit":"",
		"price":0,
		"status":""
	}`))
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, createRec.Code)
	}

	assertIntegrationErrorEnvelope(t, createRec, createReq.URL.Path, "VALIDATION_ERROR", "Request validation failed")
	assertErrorContractShape(t, createRec, []string{"name", "price", "status", "stockKeepingUnit"})

	conflictReq := httptest.NewRequest(http.MethodPatch, "/v1/products/1", strings.NewReader(`{
		"stockKeepingUnit":"SKU-100",
		"sku":"SKU-101"
	}`))
	conflictRec := httptest.NewRecorder()
	router.ServeHTTP(conflictRec, conflictReq)

	if conflictRec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, conflictRec.Code)
	}

	assertIntegrationErrorEnvelope(t, conflictRec, conflictReq.URL.Path, "VALIDATION_ERROR", "Request validation failed")
	assertErrorContractShape(t, conflictRec, []string{"stockKeepingUnit"})

	searchReq := httptest.NewRequest(http.MethodGet, "/v1/search/products?q=%20%20", nil)
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)

	if searchRec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, searchRec.Code)
	}

	assertIntegrationErrorEnvelope(t, searchRec, searchReq.URL.Path, "VALIDATION_ERROR", "Request validation failed")
	assertErrorContractShape(t, searchRec, []string{"q"})
}

type seededProductContract struct {
	ID               int64
	Name             string
	StockKeepingUnit string
	Price            float64
	Status           string
	CategoryID       *int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func seedProductContract(t *testing.T, deps *api.Dependencies, product seededProductContract) int64 {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int64
	err := deps.DB.QueryRow(ctx, `
		INSERT INTO products (name, stock_keeping_unit, price, status, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, product.Name, product.StockKeepingUnit, product.Price, product.Status, product.CategoryID, product.CreatedAt, product.UpdatedAt).Scan(&id)
	if err != nil {
		t.Fatalf("seed product: %v", err)
	}

	return id
}

func assertProductContractShapeWithoutTimestamps(t *testing.T, body map[string]any, name, sku string, price float64, status string) int64 {
	t.Helper()

	assertObjectKeys(t, body, "createdAt", "id", "name", "price", "status", "stockKeepingUnit", "updatedAt")
	if _, exists := body["categoryId"]; exists {
		t.Fatalf("expected categoryId to be omitted when empty, got %#v", body["categoryId"])
	}

	id := assertNumberValue(t, body["id"], 0, "id")
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
	assertStringValue(t, body["name"], name, "name")
	assertStringValue(t, body["stockKeepingUnit"], sku, "stockKeepingUnit")
	assertFloatValue(t, body["price"], price, "price")
	assertStringValue(t, body["status"], status, "status")
	assertRFC3339Value(t, body["createdAt"], "createdAt")
	assertRFC3339Value(t, body["updatedAt"], "updatedAt")

	return int64(id)
}

func assertProductContractShape(t *testing.T, body map[string]any, expected seededProductContract) {
	t.Helper()

	assertObjectKeys(t, body, "createdAt", "id", "name", "price", "status", "stockKeepingUnit", "updatedAt")
	if _, exists := body["categoryId"]; exists {
		t.Fatalf("expected categoryId to be omitted when empty, got %#v", body["categoryId"])
	}

	assertNumberValue(t, body["id"], int(expected.ID), "id")
	assertStringValue(t, body["name"], expected.Name, "name")
	assertStringValue(t, body["stockKeepingUnit"], expected.StockKeepingUnit, "stockKeepingUnit")
	assertFloatValue(t, body["price"], expected.Price, "price")
	assertStringValue(t, body["status"], expected.Status, "status")
	assertTimeValue(t, body["createdAt"], expected.CreatedAt, "createdAt")
	assertTimeValue(t, body["updatedAt"], expected.UpdatedAt, "updatedAt")
}

func assertErrorContractShape(t *testing.T, rec *httptest.ResponseRecorder, expectedFields []string) {
	t.Helper()

	var body map[string]any
	decodeResponse(t, rec, &body)
	assertObjectKeys(t, body, "error", "path", "timestamp")

	errorBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %#v", body["error"])
	}
	if len(expectedFields) == 0 {
		assertObjectKeys(t, errorBody, "code", "message")
		if _, exists := errorBody["details"]; exists {
			t.Fatalf("expected error.details to be omitted, got %#v", errorBody["details"])
		}
		return
	}

	assertObjectKeys(t, errorBody, "code", "details", "message")

	details := requireObjectSlice(t, errorBody["details"], "error.details")
	gotFields := make([]string, 0, len(details))
	for _, detail := range details {
		assertObjectKeys(t, detail, "constraints", "field")
		field, ok := detail["field"].(string)
		if !ok {
			t.Fatalf("expected error detail field to be a string, got %#v", detail["field"])
		}
		if _, ok := detail["constraints"].([]any); !ok {
			t.Fatalf("expected constraints array for field %q, got %#v", field, detail["constraints"])
		}
		gotFields = append(gotFields, field)
	}

	sort.Strings(gotFields)
	sort.Strings(expectedFields)
	if strings.Join(gotFields, ",") != strings.Join(expectedFields, ",") {
		t.Fatalf("expected error detail fields %v, got %v", expectedFields, gotFields)
	}
}

func requireObjectSlice(t *testing.T, value any, field string) []map[string]any {
	t.Helper()

	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected %s to be an array, got %#v", field, value)
	}

	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("expected %s items to be objects, got %#v", field, item)
		}
		result = append(result, object)
	}

	return result
}

func assertObjectKeys(t *testing.T, object map[string]any, expectedKeys ...string) {
	t.Helper()

	gotKeys := make([]string, 0, len(object))
	for key := range object {
		gotKeys = append(gotKeys, key)
	}

	sort.Strings(gotKeys)
	sort.Strings(expectedKeys)
	if strings.Join(gotKeys, ",") != strings.Join(expectedKeys, ",") {
		t.Fatalf("expected keys %v, got %v", expectedKeys, gotKeys)
	}
}

func assertStringValue(t *testing.T, value any, expected, field string) {
	t.Helper()

	got, ok := value.(string)
	if !ok {
		t.Fatalf("expected %s to be a string, got %#v", field, value)
	}
	if got != expected {
		t.Fatalf("expected %s %q, got %q", field, expected, got)
	}
}

func assertFloatValue(t *testing.T, value any, expected float64, field string) {
	t.Helper()

	got, ok := value.(float64)
	if !ok {
		t.Fatalf("expected %s to be a number, got %#v", field, value)
	}
	if got != expected {
		t.Fatalf("expected %s %v, got %v", field, expected, got)
	}
}

func assertNumberValue(t *testing.T, value any, expected int, field string) int {
	t.Helper()

	got, ok := value.(float64)
	if !ok {
		t.Fatalf("expected %s to be a number, got %#v", field, value)
	}
	if expected > 0 && int(got) != expected {
		t.Fatalf("expected %s %d, got %v", field, expected, got)
	}

	return int(got)
}

func assertTimeValue(t *testing.T, value any, expected time.Time, field string) {
	t.Helper()

	got := assertRFC3339Value(t, value, field)
	if !got.Equal(expected) {
		t.Fatalf("expected %s %s, got %s", field, expected.Format(time.RFC3339Nano), got.Format(time.RFC3339Nano))
	}
}

func assertRFC3339Value(t *testing.T, value any, field string) time.Time {
	t.Helper()

	raw, ok := value.(string)
	if !ok {
		t.Fatalf("expected %s to be a string, got %#v", field, value)
	}

	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		t.Fatalf("expected %s to be RFC3339, got %q: %v", field, raw, err)
	}

	return parsed
}
