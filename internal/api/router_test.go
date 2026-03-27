package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	platformauth "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/auth"
)

func TestNewRouterServesVersionedHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status payload %q, got %q", "ok", body.Status)
	}
}

func TestNewRouterServesDeprecatedRoot(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Deprecation"); got != "true" {
		t.Fatalf("expected Deprecation header %q, got %q", "true", got)
	}

	var body struct {
		Message string `json:"message"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	wantMessage := "This unversioned root is deprecated. Migrate to /v1/health."
	if body.Message != wantMessage {
		t.Fatalf("expected message %q, got %q", wantMessage, body.Message)
	}
}

func TestNewRouterDoesNotExposeUnversionedHealth(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertRouterErrorEnvelope(t, rec, "/health", "NOT_FOUND", "Route not found")
}

func TestNewRouterReturnsNotFoundForUnknownRoute(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()

	NewRouter(Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertRouterErrorEnvelope(t, rec, "/unknown", "NOT_FOUND", "Route not found")
}

func TestNewRouterProtectsPromotionsEndpoints(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		PromotionAuthenticator: platformauth.DefaultStubAuthenticator(),
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/promotions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	assertRouterErrorEnvelope(t, rec, "/v1/promotions", "UNAUTHORIZED", "Authentication required")
}

func TestNewRouterLeavesNonProtectedEndpointsAccessible(t *testing.T) {
	t.Parallel()

	router := NewRouter(Dependencies{
		PromotionAuthenticator: platformauth.DefaultStubAuthenticator(),
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestWriteJSONReturnsEncodingErrorsThroughSharedEnvelope(t *testing.T) {
	t.Parallel()

	handler := apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		return writeJSON(w, http.StatusOK, struct {
			Stream chan int `json:"stream"`
		}{
			Stream: make(chan int),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/broken", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	assertRouterErrorEnvelope(t, rec, "/v1/broken", "INTERNAL_SERVER_ERROR", "Internal server error")
}

func assertRouterErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, path, code, message string) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var payload apierror.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}

	if payload.Path != path {
		t.Fatalf("expected path %q, got %q", path, payload.Path)
	}

	if payload.Error.Code != code {
		t.Fatalf("expected code %q, got %q", code, payload.Error.Code)
	}

	if payload.Error.Message != message {
		t.Fatalf("expected message %q, got %q", message, payload.Error.Message)
	}
}
