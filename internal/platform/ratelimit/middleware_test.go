package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareBlocksRequestsAfterThresholdWithinWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 27, 12, 0, 0, 0, time.UTC)
	limiter := New(2, time.Minute, func() time.Time { return now })
	handler := limiter.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/products", nil)
		req.RemoteAddr = "192.0.2.10:1234"

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d: expected status %d, got %d", i+1, http.StatusNoContent, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/products", nil)
	req.RemoteAddr = "192.0.2.10:1234"

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestMiddlewareResetsCountersAfterWindowBoundary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 27, 12, 0, 0, 0, time.UTC)
	limiter := New(1, time.Minute, func() time.Time { return now })
	handler := limiter.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodDelete, "/v1/categories/1", nil)
	firstReq.RemoteAddr = "192.0.2.11:1234"
	handler.ServeHTTP(first, firstReq)

	if first.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, first.Code)
	}

	now = now.Add(time.Minute)

	second := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodDelete, "/v1/categories/1", nil)
	secondReq.RemoteAddr = "192.0.2.11:1234"
	handler.ServeHTTP(second, secondReq)

	if second.Code != http.StatusNoContent {
		t.Fatalf("expected status %d after window reset, got %d", http.StatusNoContent, second.Code)
	}
}

func TestMiddlewareBucketsAuthenticatedRequestsSeparately(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 27, 12, 0, 0, 0, time.UTC)
	limiter := New(1, time.Minute, func() time.Time { return now })
	handler := limiter.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	firstReq := httptest.NewRequest(http.MethodPatch, "/v1/promotions/1", nil)
	firstReq.RemoteAddr = "192.0.2.12:1234"
	firstReq.Header.Set("Authorization", "Bearer promotions-admin")
	handler.ServeHTTP(first, firstReq)

	if first.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, first.Code)
	}

	second := httptest.NewRecorder()
	secondReq := httptest.NewRequest(http.MethodPatch, "/v1/promotions/1", nil)
	secondReq.RemoteAddr = "192.0.2.12:1234"
	secondReq.Header.Set("Authorization", "Bearer catalog-readonly")
	handler.ServeHTTP(second, secondReq)

	if second.Code != http.StatusNoContent {
		t.Fatalf("expected independent bucket status %d, got %d", http.StatusNoContent, second.Code)
	}
}
