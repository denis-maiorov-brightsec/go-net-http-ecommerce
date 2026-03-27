package requestlog

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewarePropagatesIncomingRequestID(t *testing.T) {
	t.Parallel()

	const requestID = "client-request-id"

	var capturedRequestID string
	handler := Middleware(testLogger(io.Discard), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	req.Header.Set(HeaderName, requestID)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedRequestID != requestID {
		t.Fatalf("expected request id %q in context, got %q", requestID, capturedRequestID)
	}

	if got := rec.Header().Get(HeaderName); got != requestID {
		t.Fatalf("expected response request id %q, got %q", requestID, got)
	}
}

func TestMiddlewareGeneratesRequestIDWhenMissing(t *testing.T) {
	t.Parallel()

	var capturedRequestID string
	handler := Middleware(testLogger(io.Discard), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedRequestID == "" {
		t.Fatal("expected generated request id in context")
	}

	if got := rec.Header().Get(HeaderName); got == "" {
		t.Fatal("expected generated request id header")
	} else if got != capturedRequestID {
		t.Fatalf("expected response request id %q, got %q", capturedRequestID, got)
	}
}

func TestMiddlewareEmitsStructuredRequestLog(t *testing.T) {
	t.Parallel()

	var logOutput bytes.Buffer
	handler := Middleware(testLogger(&logOutput), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/products", nil)
	req.Header.Set(HeaderName, "request-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var payload map[string]any
	if err := json.Unmarshal(logOutput.Bytes(), &payload); err != nil {
		t.Fatalf("decode structured log: %v", err)
	}

	if payload["msg"] != "http request completed" {
		t.Fatalf("expected message %q, got %#v", "http request completed", payload["msg"])
	}

	if payload["request_id"] != "request-123" {
		t.Fatalf("expected request_id %q, got %#v", "request-123", payload["request_id"])
	}

	if payload["method"] != http.MethodPost {
		t.Fatalf("expected method %q, got %#v", http.MethodPost, payload["method"])
	}

	if payload["path"] != "/v1/products" {
		t.Fatalf("expected path %q, got %#v", "/v1/products", payload["path"])
	}

	if payload["status"] != float64(http.StatusCreated) {
		t.Fatalf("expected status %d, got %#v", http.StatusCreated, payload["status"])
	}

	if _, ok := payload["latency_ms"]; !ok {
		t.Fatal("expected latency_ms field to be present")
	}
}

func testLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil))
}
