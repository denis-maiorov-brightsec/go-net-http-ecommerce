package service

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
	productsrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products/repository"
)

type repositoryStub struct {
	getErr    error
	updateErr error
	searchErr error
}

func (r repositoryStub) List(context.Context, products.ListInput) (products.ListResult, error) {
	return products.ListResult{}, nil
}

func (r repositoryStub) Search(context.Context, products.SearchInput) ([]products.Product, error) {
	return nil, r.searchErr
}

func (r repositoryStub) GetByID(context.Context, int64) (products.Product, error) {
	return products.Product{}, r.getErr
}

func (r repositoryStub) Create(context.Context, products.CreateInput) (products.Product, error) {
	return products.Product{}, nil
}

func (r repositoryStub) Update(context.Context, int64, products.UpdateInput) (products.Product, error) {
	return products.Product{}, r.updateErr
}

func (r repositoryStub) Delete(context.Context, int64) error {
	return nil
}

func TestCreateValidatesRequiredFields(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Create(context.Background(), products.CreateInput{})
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

	if len(appErr.Details) != 4 {
		t.Fatalf("expected 4 validation details, got %d", len(appErr.Details))
	}
}

func TestCreateUsesCanonicalStockKeepingUnitValidationField(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Create(context.Background(), products.CreateInput{
		Name:   "Desk",
		Price:  129.99,
		Status: "draft",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) == 0 || appErr.Details[0].Field != "stockKeepingUnit" {
		t.Fatalf("expected stockKeepingUnit validation detail, got %#v", appErr.Details)
	}
}

func TestUpdateRejectsEmptyPatch(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Update(context.Background(), 1, products.UpdateInput{})
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

func TestUpdateUsesCanonicalStockKeepingUnitValidationField(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})
	value := ""

	_, err := svc.Update(context.Background(), 1, products.UpdateInput{StockKeepingUnit: &value})
	if err == nil {
		t.Fatal("expected validation error")
	}

	var appErr *apierror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected api error, got %T", err)
	}

	if len(appErr.Details) == 0 || appErr.Details[0].Field != "stockKeepingUnit" {
		t.Fatalf("expected stockKeepingUnit validation detail, got %#v", appErr.Details)
	}
}

func TestGetByIDMapsMissingProductToNotFound(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{getErr: productsrepository.ErrNotFound})

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

func TestUpdateMapsUnexpectedRepositoryErrorsToInternal(t *testing.T) {
	t.Parallel()

	name := "Desk"
	svc := New(repositoryStub{updateErr: errors.New("db offline")})

	_, err := svc.Update(context.Background(), 1, products.UpdateInput{Name: &name})
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

func TestSearchRejectsBlankQuery(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{})

	_, err := svc.Search(context.Background(), products.SearchInput{Query: "   "})
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

	if len(appErr.Details) != 1 || appErr.Details[0].Field != "q" {
		t.Fatalf("expected q validation detail, got %#v", appErr.Details)
	}
}

func TestSearchMapsUnexpectedRepositoryErrorsToInternal(t *testing.T) {
	t.Parallel()

	svc := New(repositoryStub{searchErr: errors.New("db offline")})

	_, err := svc.Search(context.Background(), products.SearchInput{Query: "desk"})
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
