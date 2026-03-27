package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	ordersrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/repository"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type repositoryStub struct {
	listErr    error
	getErr     error
	cancelErr  error
	cancelItem orders.Order
}

func (r repositoryStub) List(context.Context, orders.ListInput) (orders.ListResult, error) {
	return orders.ListResult{}, r.listErr
}

func (r repositoryStub) GetByID(context.Context, int64) (orders.Order, error) {
	return orders.Order{}, r.getErr
}

func (r repositoryStub) Cancel(context.Context, int64) (orders.Order, error) {
	return r.cancelItem, r.cancelErr
}

func TestListRejectsInvalidDateRange(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})
	from := time.Date(2026, time.January, 3, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)

	_, err := svc.List(context.Background(), orders.ListInput{
		Page:  1,
		Limit: 20,
		From:  &from,
		To:    &to,
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
}

func TestGetByIDMapsMissingOrderToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{getErr: ordersrepository.ErrNotFound})

	_, err := svc.GetByID(context.Background(), 42)
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

func TestListMapsUnexpectedRepositoryErrorsToInternal(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{listErr: errors.New("db offline")})

	_, err := svc.List(context.Background(), orders.ListInput{Page: 1, Limit: 20})
	if err == nil {
		t.Fatal("expected internal error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if appErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, appErr.Status)
	}
}

func TestCancelMapsMissingOrderToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{cancelErr: ordersrepository.ErrNotFound})

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

	svc := New(repositoryStub{cancelErr: ordersrepository.ErrIneligibleStatus})

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
	svc := New(repositoryStub{cancelItem: expected})

	item, err := svc.Cancel(context.Background(), 42)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}

	if item.ID != expected.ID || item.Status != expected.Status {
		t.Fatalf("expected %#v, got %#v", expected, item)
	}
}
