package apierror

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAdaptMapsUnknownErrorsToInternalEnvelope(t *testing.T) {
	t.Parallel()

	handler := Adapt(func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("database offline")
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/fail", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	assertErrorEnvelope(t, rec, "/v1/fail", "INTERNAL_SERVER_ERROR", "Internal server error", nil)
}

func TestRecoverMapsPanicsToInternalEnvelope(t *testing.T) {
	t.Parallel()

	handler := Recover(slog.Default(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	assertErrorEnvelope(t, rec, "/v1/panic", "INTERNAL_SERVER_ERROR", "Internal server error", nil)
}

func TestNormalizeServeMuxMapsMissingRoutesToNotFoundEnvelope(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.Handle("GET /v1/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/missing", nil)
	rec := httptest.NewRecorder()

	NormalizeServeMux(mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	assertErrorEnvelope(t, rec, "/v1/missing", "NOT_FOUND", "Route not found", nil)
}

func TestNormalizeServeMuxMapsMethodNotAllowedToEnvelope(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.Handle("GET /v1/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/health", nil)
	rec := httptest.NewRecorder()

	NormalizeServeMux(mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}

	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("expected Allow header %q, got %q", "GET, HEAD", got)
	}

	assertErrorEnvelope(t, rec, "/v1/health", "METHOD_NOT_ALLOWED", "Method not allowed", nil)
}

func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder, path, code, message string, details []Detail) {
	t.Helper()

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var payload Envelope
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

	if len(payload.Error.Details) != len(details) {
		t.Fatalf("expected %d details, got %d", len(details), len(payload.Error.Details))
	}

	for i, detail := range details {
		if !reflect.DeepEqual(payload.Error.Details[i], detail) {
			t.Fatalf("expected detail %+v, got %+v", detail, payload.Error.Details[i])
		}
	}
}
