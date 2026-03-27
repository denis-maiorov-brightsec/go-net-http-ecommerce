package pagination

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

func TestParseUsesDefaults(t *testing.T) {
	t.Parallel()

	params, err := Parse(url.Values{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if params.Page != DefaultPage {
		t.Fatalf("expected default page %d, got %d", DefaultPage, params.Page)
	}
	if params.Limit != DefaultLimit {
		t.Fatalf("expected default limit %d, got %d", DefaultLimit, params.Limit)
	}
}

func TestParseRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	_, err := Parse(url.Values{
		"page":  []string{"0"},
		"limit": []string{"101"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if appErr.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, appErr.Status)
	}
	if len(appErr.Details) != 2 {
		t.Fatalf("expected 2 validation details, got %d", len(appErr.Details))
	}
}

func TestNewMetadataCalculatesTotalPages(t *testing.T) {
	t.Parallel()

	metadata := NewMetadata(Params{Page: 2, Limit: 2}, 5)

	if metadata.Page != 2 || metadata.Limit != 2 {
		t.Fatalf("unexpected params metadata: %#v", metadata)
	}
	if metadata.Total != 5 {
		t.Fatalf("expected total 5, got %d", metadata.Total)
	}
	if metadata.TotalPages != 3 {
		t.Fatalf("expected totalPages 3, got %d", metadata.TotalPages)
	}
}
