package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := NewMiddleware(DefaultStubAuthenticator()).Require(ManagePromotionsPermission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("expected handler not to be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/promotions", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestMiddlewareRejectsAuthenticatedPrincipalWithoutPermission(t *testing.T) {
	t.Parallel()

	handler := NewMiddleware(DefaultStubAuthenticator()).Require(ManagePromotionsPermission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("expected handler not to be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/promotions", nil)
	req.Header.Set("Authorization", "Bearer catalog-readonly")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestMiddlewareAllowsAuthenticatedPrincipalWithPermission(t *testing.T) {
	t.Parallel()

	called := false
	handler := NewMiddleware(DefaultStubAuthenticator()).Require(ManagePromotionsPermission)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/promotions", nil)
	req.Header.Set("Authorization", "Bearer promotions-admin")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
