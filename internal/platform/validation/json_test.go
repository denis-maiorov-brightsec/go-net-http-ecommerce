package validation

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type createProductRequest struct {
	Name string `json:"name" validate:"required"`
}

func TestDecodeJSONReturnsValidationEnvelopeForMissingRequiredFields(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	var payload createProductRequest
	err := DecodeJSON(rec, req, &payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	appErr := apierror.Normalize(err)
	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, appErr.Status)
	}

	if appErr.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected code %q, got %q", "VALIDATION_ERROR", appErr.Code)
	}

	wantDetails := []apierror.Detail{{
		Field:       "name",
		Constraints: []string{"name must not be empty"},
	}}

	if len(appErr.Details) != len(wantDetails) {
		t.Fatalf("expected %d details, got %d", len(wantDetails), len(appErr.Details))
	}

	if !reflect.DeepEqual(appErr.Details[0], wantDetails[0]) {
		t.Fatalf("expected detail %+v, got %+v", wantDetails[0], appErr.Details[0])
	}
}

func TestDecodeJSONRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{"name":"Desk","sku":"SKU-1"}`))
	rec := httptest.NewRecorder()

	var payload struct {
		Name string `json:"name" validate:"required"`
	}

	err := DecodeJSON(rec, req, &payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	appErr := apierror.Normalize(err)
	if appErr.Details[0].Field != "sku" {
		t.Fatalf("expected field %q, got %q", "sku", appErr.Details[0].Field)
	}
}

func TestDecodeJSONReadsSingleJSONObject(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{"name":"Desk"}{"name":"Lamp"}`))
	rec := httptest.NewRecorder()

	var payload createProductRequest
	err := DecodeJSON(rec, req, &payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	appErr := apierror.Normalize(err)
	if appErr.Details[0].Field != "body" {
		t.Fatalf("expected field %q, got %q", "body", appErr.Details[0].Field)
	}
}

func TestValidationErrorCanBeWrittenAsSharedEnvelope(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/products", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	var payload createProductRequest
	err := DecodeJSON(rec, req, &payload)
	if err == nil {
		t.Fatal("expected validation error")
	}

	rec = httptest.NewRecorder()
	apierror.Write(rec, req, err)

	var body apierror.Envelope
	if decodeErr := json.Unmarshal(rec.Body.Bytes(), &body); decodeErr != nil {
		t.Fatalf("decode error envelope: %v", decodeErr)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	if body.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected code %q, got %q", "VALIDATION_ERROR", body.Error.Code)
	}

	if body.Error.Message != "Request validation failed" {
		t.Fatalf("expected message %q, got %q", "Request validation failed", body.Error.Message)
	}
}
