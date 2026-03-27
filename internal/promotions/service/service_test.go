package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions"
	promotionsrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions/repository"
)

type repositoryStub struct {
	getItem   promotions.Promotion
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (r repositoryStub) List(context.Context) ([]promotions.Promotion, error) {
	return nil, nil
}

func (r repositoryStub) GetByID(context.Context, int64) (promotions.Promotion, error) {
	if r.getErr != nil {
		return promotions.Promotion{}, r.getErr
	}

	return r.getItem, nil
}

func (r repositoryStub) Create(context.Context, promotions.CreateInput) (promotions.Promotion, error) {
	return promotions.Promotion{}, r.createErr
}

func (r repositoryStub) Update(context.Context, int64, promotions.UpdateInput) (promotions.Promotion, error) {
	return promotions.Promotion{}, r.updateErr
}

func (r repositoryStub) Delete(context.Context, int64) error {
	return r.deleteErr
}

func TestCreateValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Create(context.Background(), promotions.CreateInput{})
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

	if len(appErr.Details) != 5 {
		t.Fatalf("expected 5 validation details, got %d", len(appErr.Details))
	}
}

func TestCreateRejectsInvalidDateWindow(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})
	startsAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	_, err := svc.Create(context.Background(), promotions.CreateInput{
		Name:          "Spring Sale",
		Code:          "SPRING",
		DiscountType:  "percentage",
		DiscountValue: 15,
		StartsAt:      &startsAt,
		EndsAt:        &endsAt,
		Status:        "draft",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "startsAt" {
		t.Fatalf("expected startsAt validation detail, got %#v", appErr.Details)
	}
}

func TestCreateMapsDuplicateCodeToConflict(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{createErr: promotionsrepository.ErrDuplicateCode})

	_, err := svc.Create(context.Background(), promotions.CreateInput{
		Name:          "Spring Sale",
		Code:          "SPRING",
		DiscountType:  "percentage",
		DiscountValue: 15,
		Status:        "draft",
	})
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
}

func TestUpdateRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Update(context.Background(), 1, promotions.UpdateInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "body" {
		t.Fatalf("expected body validation detail, got %#v", appErr.Details)
	}
}

func TestUpdateRejectsInvalidDateWindowWhenBothFieldsProvided(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})
	startsAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	_, err := svc.Update(context.Background(), 1, promotions.UpdateInput{
		StartsAt: promotions.OptionalTime{Set: true, Value: &startsAt},
		EndsAt:   promotions.OptionalTime{Set: true, Value: &endsAt},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "startsAt" {
		t.Fatalf("expected startsAt validation detail, got %#v", appErr.Details)
	}
}

func TestUpdateRejectsInvalidDateWindowAgainstStoredPromotion(t *testing.T) {
	t.Parallel()

	endsAt := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	startsAt := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	svc := New(repositoryStub{
		getItem: promotions.Promotion{
			ID:     1,
			EndsAt: &endsAt,
			Status: "draft",
			Name:   "Spring Sale",
			Code:   "SPRING",
		},
	})

	_, err := svc.Update(context.Background(), 1, promotions.UpdateInput{
		StartsAt: promotions.OptionalTime{Set: true, Value: &startsAt},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "startsAt" {
		t.Fatalf("expected startsAt validation detail, got %#v", appErr.Details)
	}
}

func TestGetByIDMapsMissingPromotionToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{getErr: promotionsrepository.ErrNotFound})

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

func TestDeleteMapsMissingPromotionToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{deleteErr: promotionsrepository.ErrNotFound})

	err := svc.Delete(context.Background(), 42)
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
