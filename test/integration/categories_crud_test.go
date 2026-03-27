//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

func TestCategoriesCRUDFlow(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	createReq := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(`{
		"name":"Furniture",
		"slug":"furniture",
		"description":"Indoor furniture"
	}`))
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created categoryResponse
	decodeResponse(t, createRec, &created)

	if created.ID == 0 {
		t.Fatal("expected created category id to be set")
	}
	if created.Name != "Furniture" || created.Slug != "furniture" {
		t.Fatalf("unexpected created category: %#v", created)
	}
	if created.Description == nil || *created.Description != "Indoor furniture" {
		t.Fatalf("expected description to be set, got %#v", created.Description)
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("expected timestamps to be set, got %#v", created)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	listRec := httptest.NewRecorder()

	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var listBody listCategoriesResponse
	decodeResponse(t, listRec, &listBody)

	if len(listBody.Items) != 1 {
		t.Fatalf("expected 1 category, got %d", len(listBody.Items))
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/categories/"+int64Path(created.ID), nil)
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var fetched categoryResponse
	decodeResponse(t, getRec, &fetched)

	if fetched.ID != created.ID {
		t.Fatalf("expected id %d, got %d", created.ID, fetched.ID)
	}

	updateReq := httptest.NewRequest(http.MethodPatch, "/v1/categories/"+int64Path(created.ID), strings.NewReader(`{
		"name":"Home Furniture",
		"slug":"home-furniture",
		"description":null
	}`))
	updateRec := httptest.NewRecorder()

	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRec.Code)
	}

	var updated categoryResponse
	decodeResponse(t, updateRec, &updated)

	if updated.Name != "Home Furniture" || updated.Slug != "home-furniture" {
		t.Fatalf("unexpected updated category: %#v", updated)
	}
	if updated.Description != nil {
		t.Fatalf("expected description to be cleared, got %#v", updated.Description)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+int64Path(created.ID), nil)
	deleteRec := httptest.NewRecorder()

	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/v1/categories/"+int64Path(created.ID), nil)
	missingRec := httptest.NewRecorder()

	router.ServeHTTP(missingRec, missingReq)

	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, missingRec.Code)
	}

	assertIntegrationErrorEnvelope(t, missingRec, missingReq.URL.Path, "NOT_FOUND", "Category not found")
}

func TestCreateCategoryRejectsDuplicateSlug(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for _, body := range []string{
		`{"name":"Furniture","slug":"furniture"}`,
		`{"name":"Decor","slug":"furniture"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if body == `{"name":"Furniture","slug":"furniture"}` {
			if rec.Code != http.StatusCreated {
				t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
			}
			continue
		}

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}

		assertIntegrationErrorEnvelope(t, rec, "/v1/categories", "CONFLICT", "Category slug already exists")
	}
}

func TestUpdateCategoryRejectsDuplicateSlug(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for _, body := range []string{
		`{"name":"Furniture","slug":"furniture"}`,
		`{"name":"Decor","slug":"decor"}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/categories/2", strings.NewReader(`{"slug":"furniture"}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "CONFLICT", "Category slug already exists")
}

func TestCategoryNotFoundPaths(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	testCases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "get", method: http.MethodGet, path: "/v1/categories/999"},
		{name: "patch", method: http.MethodPatch, path: "/v1/categories/999", body: `{"name":"Updated"}`},
		{name: "delete", method: http.MethodDelete, path: "/v1/categories/999"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
			}

			assertIntegrationErrorEnvelope(t, rec, tc.path, "NOT_FOUND", "Category not found")
		})
	}
}

func TestCategoryValidationErrors(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(`{
		"name":"",
		"slug":""
	}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload apierror.Envelope
	decodeCategoryResponse(t, rec, &payload)

	if len(payload.Error.Details) != 2 {
		t.Fatalf("expected 2 validation details, got %d", len(payload.Error.Details))
	}
}

type listCategoriesResponse struct {
	Items []categoryResponse `json:"items"`
}

type categoryResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

func decodeCategoryResponse(t *testing.T, rec *httptest.ResponseRecorder, dst any) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	if err := json.Unmarshal(rec.Body.Bytes(), dst); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}
