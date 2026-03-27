package service

import (
	"context"
	"errors"
	"strings"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories"
	categoriesrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories/repository"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type Repository interface {
	List(context.Context) ([]categories.Category, error)
	GetByID(context.Context, int64) (categories.Category, error)
	Create(context.Context, categories.CreateInput) (categories.Category, error)
	Update(context.Context, int64, categories.UpdateInput) (categories.Category, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]categories.Category, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (categories.Category, error) {
	if err := validateID(id); err != nil {
		return categories.Category{}, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, categoriesrepository.ErrNotFound) {
			return categories.Category{}, apierror.NotFound("Category not found")
		}

		return categories.Category{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Create(ctx context.Context, input categories.CreateInput) (categories.Category, error) {
	if err := validateCreateInput(input); err != nil {
		return categories.Category{}, err
	}

	item, err := s.repo.Create(ctx, input)
	if err != nil {
		if errors.Is(err, categoriesrepository.ErrDuplicateSlug) {
			return categories.Category{}, apierror.New(409, "CONFLICT", "Category slug already exists", nil)
		}

		return categories.Category{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Update(ctx context.Context, id int64, input categories.UpdateInput) (categories.Category, error) {
	if err := validateID(id); err != nil {
		return categories.Category{}, err
	}

	if err := validateUpdateInput(input); err != nil {
		return categories.Category{}, err
	}

	item, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, categoriesrepository.ErrNotFound) {
			return categories.Category{}, apierror.NotFound("Category not found")
		}
		if errors.Is(err, categoriesrepository.ErrDuplicateSlug) {
			return categories.Category{}, apierror.New(409, "CONFLICT", "Category slug already exists", nil)
		}

		return categories.Category{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, categoriesrepository.ErrNotFound) {
			return apierror.NotFound("Category not found")
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

func validateCreateInput(input categories.CreateInput) error {
	details := make([]apierror.Detail, 0, 2)

	if strings.TrimSpace(input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if strings.TrimSpace(input.Slug) == "" {
		details = append(details, apierror.Detail{
			Field:       "slug",
			Constraints: []string{"slug must not be empty"},
		})
	}

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}

func validateUpdateInput(input categories.UpdateInput) error {
	if !input.HasUpdates() {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must include at least one updatable field"},
		}})
	}

	details := make([]apierror.Detail, 0, 2)

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if input.Slug != nil && strings.TrimSpace(*input.Slug) == "" {
		details = append(details, apierror.Detail{
			Field:       "slug",
			Constraints: []string{"slug must not be empty"},
		})
	}

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}
