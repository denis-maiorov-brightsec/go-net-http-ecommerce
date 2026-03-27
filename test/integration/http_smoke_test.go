//go:build integration

package integration_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/validation"
)

func TestVersionedHealthRoute(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(api.Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", body.Status)
	}
}

func TestDeprecatedRootRoute(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(api.Dependencies{}).ServeHTTP(rec, req)

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

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	api.NewRouter(api.Dependencies{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, "/missing", "NOT_FOUND", "Route not found")
}

func TestValidationFailureUsesSharedErrorEnvelope(t *testing.T) {
	t.Parallel()

	type requestBody struct {
		Name string `json:"name" validate:"required"`
	}

	handler := apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		var payload requestBody
		if err := validation.DecodeJSON(w, r, &payload); err != nil {
			return err
		}

		w.WriteHeader(http.StatusCreated)
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, "/v1/products", "VALIDATION_ERROR", "Request validation failed")
}

func TestUnknownRuntimeErrorUsesSanitizedEnvelope(t *testing.T) {
	t.Parallel()

	handler := apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("database offline")
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/products", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	assertIntegrationErrorEnvelope(t, rec, "/v1/products", "INTERNAL_SERVER_ERROR", "Internal server error")
}

func assertIntegrationErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, path, code, message string) {
	t.Helper()

	var payload apierror.Envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}

	if payload.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
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
