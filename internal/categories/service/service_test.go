package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories"
	categoriesrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories/repository"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type repositoryStub struct {
	getErr    error
	createErr error
	updateErr error
	deleteErr error
}

func (r repositoryStub) List(context.Context) ([]categories.Category, error) {
	return nil, nil
}

func (r repositoryStub) GetByID(context.Context, int64) (categories.Category, error) {
	return categories.Category{}, r.getErr
}

func (r repositoryStub) Create(context.Context, categories.CreateInput) (categories.Category, error) {
	return categories.Category{}, r.createErr
}

func (r repositoryStub) Update(context.Context, int64, categories.UpdateInput) (categories.Category, error) {
	return categories.Category{}, r.updateErr
}

func (r repositoryStub) Delete(context.Context, int64) error {
	return r.deleteErr
}

func TestCreateValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Create(context.Background(), categories.CreateInput{})
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

func TestCreateMapsDuplicateSlugToConflict(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{createErr: categoriesrepository.ErrDuplicateSlug})

	_, err := svc.Create(context.Background(), categories.CreateInput{Name: "Furniture", Slug: "furniture"})
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

func TestUpdateRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Update(context.Background(), 1, categories.UpdateInput{})
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

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "body" {
		t.Fatalf("expected body validation detail, got %#v", appErr.Details)
	}
}

func TestUpdateMapsDuplicateSlugToConflict(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{updateErr: categoriesrepository.ErrDuplicateSlug})

	name := "Furniture"
	slug := "furniture"
	_, err := svc.Update(context.Background(), 1, categories.UpdateInput{
		Name: &name,
		Slug: &slug,
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

	if appErr.Code != "CONFLICT" {
		t.Fatalf("expected code %q, got %q", "CONFLICT", appErr.Code)
	}
}

func TestGetByIDMapsMissingCategoryToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{getErr: categoriesrepository.ErrNotFound})

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

func TestDeleteMapsMissingCategoryToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{deleteErr: categoriesrepository.ErrNotFound})

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
