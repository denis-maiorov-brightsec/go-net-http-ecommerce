//go:build integration

package integration_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

func TestPromotionsCRUDFlow(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	createReq := promotionsRequest(http.MethodPost, "/v1/promotions", strings.NewReader(`{
		"name":"Spring Sale",
		"code":"SPRING-2026",
		"discountType":"percentage",
		"discountValue":15,
		"startsAt":"2026-04-01T00:00:00Z",
		"endsAt":"2026-04-30T23:59:59Z",
		"status":"draft"
	}`), "promotions-admin")
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created promotionResponse
	decodeResponse(t, createRec, &created)

	if created.ID == 0 {
		t.Fatal("expected created promotion id to be set")
	}
	if created.Name != "Spring Sale" || created.Code != "SPRING-2026" {
		t.Fatalf("unexpected created promotion: %#v", created)
	}
	if created.StartsAt == nil || created.EndsAt == nil {
		t.Fatalf("expected date window to be set, got %#v", created)
	}

	listReq := promotionsRequest(http.MethodGet, "/v1/promotions", nil, "promotions-admin")
	listRec := httptest.NewRecorder()

	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}

	var listBody listPromotionsResponse
	decodeResponse(t, listRec, &listBody)

	if len(listBody.Items) != 1 {
		t.Fatalf("expected 1 promotion, got %d", len(listBody.Items))
	}

	getReq := promotionsRequest(http.MethodGet, "/v1/promotions/"+int64Path(created.ID), nil, "promotions-admin")
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	var fetched promotionResponse
	decodeResponse(t, getRec, &fetched)

	if fetched.ID != created.ID {
		t.Fatalf("expected id %d, got %d", created.ID, fetched.ID)
	}

	updateReq := promotionsRequest(http.MethodPatch, "/v1/promotions/"+int64Path(created.ID), strings.NewReader(`{
		"discountValue":20,
		"status":"active",
		"endsAt":null
	}`), "promotions-admin")
	updateRec := httptest.NewRecorder()

	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRec.Code)
	}

	var updated promotionResponse
	decodeResponse(t, updateRec, &updated)

	if updated.DiscountValue != 20 || updated.Status != "active" {
		t.Fatalf("unexpected updated promotion: %#v", updated)
	}
	if updated.EndsAt != nil {
		t.Fatalf("expected endsAt to be cleared, got %#v", updated.EndsAt)
	}

	deleteReq := promotionsRequest(http.MethodDelete, "/v1/promotions/"+int64Path(created.ID), nil, "promotions-admin")
	deleteRec := httptest.NewRecorder()

	router.ServeHTTP(deleteRec, deleteReq)

	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}

	missingReq := promotionsRequest(http.MethodGet, "/v1/promotions/"+int64Path(created.ID), nil, "promotions-admin")
	missingRec := httptest.NewRecorder()

	router.ServeHTTP(missingRec, missingReq)

	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, missingRec.Code)
	}

	assertIntegrationErrorEnvelope(t, missingRec, missingReq.URL.Path, "NOT_FOUND", "Promotion not found")
}

func TestCreatePromotionRejectsDuplicateCode(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	for _, body := range []string{
		`{"name":"Spring Sale","code":"SPRING-2026","discountType":"percentage","discountValue":15,"status":"draft"}`,
		`{"name":"Summer Sale","code":"SPRING-2026","discountType":"percentage","discountValue":10,"status":"draft"}`,
	} {
		req := promotionsRequest(http.MethodPost, "/v1/promotions", strings.NewReader(body), "promotions-admin")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if strings.Contains(body, `"name":"Spring Sale"`) {
			if rec.Code != http.StatusCreated {
				t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
			}
			continue
		}

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
		}

		assertIntegrationErrorEnvelope(t, rec, "/v1/promotions", "CONFLICT", "Promotion code already exists")
	}
}

func TestPromotionInvalidDateWindowReturnsValidationError(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := promotionsRequest(http.MethodPost, "/v1/promotions", strings.NewReader(`{
		"name":"Spring Sale",
		"code":"SPRING-2026",
		"discountType":"percentage",
		"discountValue":15,
		"startsAt":"2026-05-01T00:00:00Z",
		"endsAt":"2026-04-01T00:00:00Z",
		"status":"draft"
	}`), "promotions-admin")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var payload apierror.Envelope
	decodeResponse(t, rec, &payload)

	if len(payload.Error.Details) != 1 || payload.Error.Details[0].Field != "startsAt" {
		t.Fatalf("expected startsAt validation detail, got %#v", payload.Error.Details)
	}
}

func TestPromotionPartialUpdateRejectsInvalidDateWindow(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	createReq := promotionsRequest(http.MethodPost, "/v1/promotions", strings.NewReader(`{
		"name":"Spring Sale",
		"code":"SPRING-2026",
		"discountType":"percentage",
		"discountValue":15,
		"endsAt":"2026-04-01T00:00:00Z",
		"status":"draft"
	}`), "promotions-admin")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created promotionResponse
	decodeResponse(t, createRec, &created)

	updateReq := promotionsRequest(http.MethodPatch, "/v1/promotions/"+int64Path(created.ID), strings.NewReader(`{
		"startsAt":"2026-05-01T00:00:00Z"
	}`), "promotions-admin")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	if updateRec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, updateRec.Code)
	}

	var payload apierror.Envelope
	decodeResponse(t, updateRec, &payload)

	if len(payload.Error.Details) != 1 || payload.Error.Details[0].Field != "startsAt" {
		t.Fatalf("expected startsAt validation detail, got %#v", payload.Error.Details)
	}
}

func TestDeleteMissingPromotionReturnsNotFound(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := promotionsRequest(http.MethodDelete, "/v1/promotions/999", nil, "promotions-admin")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "NOT_FOUND", "Promotion not found")
}

func TestPromotionsRejectUnauthenticatedRequests(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/promotions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "UNAUTHORIZED", "Authentication required")
}

func TestPromotionsRejectAuthenticatedRequestsWithoutPermission(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouter(t))

	req := promotionsRequest(http.MethodGet, "/v1/promotions", nil, "catalog-readonly")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, req.URL.Path, "FORBIDDEN", "Forbidden")
}

func TestPromotionWriteRoutesReturnRateLimitExceededWhileReadsRemainAvailable(t *testing.T) {
	router := api.NewRouter(*newIntegrationRouterWithRateLimit(t, 1, time.Minute))

	firstReq := promotionsRequest(http.MethodDelete, "/v1/promotions/999", nil, "promotions-admin")
	firstRec := httptest.NewRecorder()
	router.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, firstRec.Code)
	}

	secondReq := promotionsRequest(http.MethodDelete, "/v1/promotions/999", nil, "promotions-admin")
	secondRec := httptest.NewRecorder()
	router.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, secondRec.Code)
	}

	assertIntegrationErrorEnvelope(t, secondRec, secondReq.URL.Path, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded")

	readReq := promotionsRequest(http.MethodGet, "/v1/promotions", nil, "promotions-admin")
	readRec := httptest.NewRecorder()
	router.ServeHTTP(readRec, readReq)

	if readRec.Code != http.StatusOK {
		t.Fatalf("expected read status %d, got %d", http.StatusOK, readRec.Code)
	}
}

func promotionsRequest(method, path string, body *strings.Reader, token string) *http.Request {
	var reader *strings.Reader
	if body != nil {
		reader = body
	} else {
		reader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, reader)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

type listPromotionsResponse struct {
	Items []promotionResponse `json:"items"`
}

type promotionResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Code          string  `json:"code"`
	DiscountType  string  `json:"discountType"`
	DiscountValue float64 `json:"discountValue"`
	StartsAt      *string `json:"startsAt"`
	EndsAt        *string `json:"endsAt"`
	Status        string  `json:"status"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}
