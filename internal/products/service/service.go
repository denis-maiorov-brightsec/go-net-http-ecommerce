package service

import (
	"context"
	"errors"
	"strings"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
	productsrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products/repository"
)

type Repository interface {
	List(context.Context, products.ListInput) (products.ListResult, error)
	Search(context.Context, products.SearchInput) ([]products.Product, error)
	GetByID(context.Context, int64) (products.Product, error)
	Create(context.Context, products.CreateInput) (products.Product, error)
	Update(context.Context, int64, products.UpdateInput) (products.Product, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, input products.ListInput) (products.ListResult, error) {
	items, err := s.repo.List(ctx, input)
	if err != nil {
		return products.ListResult{}, apierror.Internal(err)
	}

	return items, nil
}

func (s *Service) Search(ctx context.Context, input products.SearchInput) ([]products.Product, error) {
	if err := validateSearchInput(input); err != nil {
		return nil, err
	}

	items, err := s.repo.Search(ctx, products.SearchInput{
		Query: strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, apierror.Internal(err)
	}

	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (products.Product, error) {
	if err := validateID(id); err != nil {
		return products.Product{}, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, productsrepository.ErrNotFound) {
			return products.Product{}, apierror.NotFound("Product not found")
		}

		return products.Product{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Create(ctx context.Context, input products.CreateInput) (products.Product, error) {
	if err := validateCreateInput(input); err != nil {
		return products.Product{}, err
	}

	item, err := s.repo.Create(ctx, input)
	if err != nil {
		return products.Product{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Update(ctx context.Context, id int64, input products.UpdateInput) (products.Product, error) {
	if err := validateID(id); err != nil {
		return products.Product{}, err
	}

	if err := validateUpdateInput(input); err != nil {
		return products.Product{}, err
	}

	item, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, productsrepository.ErrNotFound) {
			return products.Product{}, apierror.NotFound("Product not found")
		}

		return products.Product{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, productsrepository.ErrNotFound) {
			return apierror.NotFound("Product not found")
		}

		return apierror.Internal(err)
	}

	return nil
}

func validateID(id int64) error {
	if id <= 0 {
		return apierror.Validation([]apierror.Detail{{
			Field:       "id",
			Constraints: []string{"id must be a positive integer"},
		}})
	}

	return nil
}

func validateCreateInput(input products.CreateInput) error {
	details := make([]apierror.Detail, 0, 5)

	if strings.TrimSpace(input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if strings.TrimSpace(input.SKU) == "" {
		details = append(details, apierror.Detail{
			Field:       "sku",
			Constraints: []string{"sku must not be empty"},
		})
	}
	if input.Price <= 0 {
		details = append(details, apierror.Detail{
			Field:       "price",
			Constraints: []string{"price must be greater than 0"},
		})
	}
	if strings.TrimSpace(input.Status) == "" {
		details = append(details, apierror.Detail{
			Field:       "status",
			Constraints: []string{"status must not be empty"},
		})
	}
	if input.CategoryID != nil && *input.CategoryID <= 0 {
		details = append(details, apierror.Detail{
			Field:       "categoryId",
			Constraints: []string{"categoryId must be greater than 0"},
		})
	}

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}

func validateUpdateInput(input products.UpdateInput) error {
	if !input.HasUpdates() {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must include at least one updatable field"},
		}})
	}

	details := make([]apierror.Detail, 0, 5)

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if input.SKU != nil && strings.TrimSpace(*input.SKU) == "" {
		details = append(details, apierror.Detail{
			Field:       "sku",
			Constraints: []string{"sku must not be empty"},
		})
	}
	if input.Price != nil && *input.Price <= 0 {
		details = append(details, apierror.Detail{
			Field:       "price",
			Constraints: []string{"price must be greater than 0"},
		})
	}
	if input.Status != nil && strings.TrimSpace(*input.Status) == "" {
		details = append(details, apierror.Detail{
			Field:       "status",
			Constraints: []string{"status must not be empty"},
		})
	}
	if input.CategoryID.Set && input.CategoryID.Value != nil && *input.CategoryID.Value <= 0 {
		details = append(details, apierror.Detail{
			Field:       "categoryId",
			Constraints: []string{"categoryId must be greater than 0"},
		})
	}

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}

func validateSearchInput(input products.SearchInput) error {
	if strings.TrimSpace(input.Query) == "" {
		return apierror.Validation([]apierror.Detail{{
			Field:       "q",
			Constraints: []string{"q must not be empty"},
		}})
	}

	return nil
}
