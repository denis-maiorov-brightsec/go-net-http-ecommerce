package commands

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type repositoryStub struct {
	cancelErr  error
	cancelItem orders.Order
}

func (r repositoryStub) Cancel(context.Context, int64) (orders.Order, error) {
	return r.cancelItem, r.cancelErr
}

func TestCancelMapsMissingOrderToNotFound(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{cancelErr: ErrNotFound})

	_, err := svc.Cancel(context.Background(), 42)
	if err == nil {
		t.Fatal("expected not found error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if appErr.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, appErr.Status)
	}
}

func TestCancelMapsIneligibleStatusToConflict(t *testing.T) {
	t.Parallel()

	svc := NewService(repositoryStub{cancelErr: ErrIneligibleStatus})

	_, err := svc.Cancel(context.Background(), 42)
	if err == nil {
		t.Fatal("expected conflict error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if appErr.Status != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, appErr.Status)
	}

	if appErr.Code != "CONFLICT" {
		t.Fatalf("expected code %q, got %q", "CONFLICT", appErr.Code)
	}
}

func TestCancelReturnsUpdatedOrder(t *testing.T) {
	t.Parallel()

	expected := orders.Order{ID: 42, Status: orders.StatusCancelled}
	svc := NewService(repositoryStub{cancelItem: expected})

	item, err := svc.Cancel(context.Background(), 42)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	if item.ID != expected.ID || item.Status != expected.Status {
		t.Fatalf("expected %#v, got %#v", expected, item)
	}
}
