package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions"
	promotionsrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions/repository"
)

type Repository interface {
	List(context.Context) ([]promotions.Promotion, error)
	GetByID(context.Context, int64) (promotions.Promotion, error)
	Create(context.Context, promotions.CreateInput) (promotions.Promotion, error)
	Update(context.Context, int64, promotions.UpdateInput) (promotions.Promotion, error)
	Delete(context.Context, int64) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]promotions.Promotion, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apierror.Internal(err)
	}

	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (promotions.Promotion, error) {
	if err := validateID(id); err != nil {
		return promotions.Promotion{}, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, promotionsrepository.ErrNotFound) {
			return promotions.Promotion{}, apierror.NotFound("Promotion not found")
		}

		return promotions.Promotion{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Create(ctx context.Context, input promotions.CreateInput) (promotions.Promotion, error) {
	if err := validateCreateInput(input); err != nil {
		return promotions.Promotion{}, err
	}

	item, err := s.repo.Create(ctx, input)
	if err != nil {
		if errors.Is(err, promotionsrepository.ErrDuplicateCode) {
			return promotions.Promotion{}, apierror.New(409, "CONFLICT", "Promotion code already exists", nil)
		}

		return promotions.Promotion{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Update(ctx context.Context, id int64, input promotions.UpdateInput) (promotions.Promotion, error) {
	if err := validateID(id); err != nil {
		return promotions.Promotion{}, err
	}

	if err := validateUpdateInput(input); err != nil {
		return promotions.Promotion{}, err
	}

	item, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, promotionsrepository.ErrNotFound) {
			return promotions.Promotion{}, apierror.NotFound("Promotion not found")
		}
		if errors.Is(err, promotionsrepository.ErrDuplicateCode) {
			return promotions.Promotion{}, apierror.New(409, "CONFLICT", "Promotion code already exists", nil)
		}

		return promotions.Promotion{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := validateID(id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, promotionsrepository.ErrNotFound) {
			return apierror.NotFound("Promotion not found")
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

func validateCreateInput(input promotions.CreateInput) error {
	details := make([]apierror.Detail, 0, 6)

	if strings.TrimSpace(input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if strings.TrimSpace(input.Code) == "" {
		details = append(details, apierror.Detail{
			Field:       "code",
			Constraints: []string{"code must not be empty"},
		})
	}
	if strings.TrimSpace(input.DiscountType) == "" {
		details = append(details, apierror.Detail{
			Field:       "discountType",
			Constraints: []string{"discountType must not be empty"},
		})
	}
	if input.DiscountValue <= 0 {
		details = append(details, apierror.Detail{
			Field:       "discountValue",
			Constraints: []string{"discountValue must be greater than 0"},
		})
	}
	if strings.TrimSpace(input.Status) == "" {
		details = append(details, apierror.Detail{
			Field:       "status",
			Constraints: []string{"status must not be empty"},
		})
	}

	details = append(details, validateDateWindow(input.StartsAt, input.EndsAt)...)

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}

func validateUpdateInput(input promotions.UpdateInput) error {
	if !input.HasUpdates() {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must include at least one updatable field"},
		}})
	}

	details := make([]apierror.Detail, 0, 6)

	if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
		details = append(details, apierror.Detail{
			Field:       "name",
			Constraints: []string{"name must not be empty"},
		})
	}
	if input.Code != nil && strings.TrimSpace(*input.Code) == "" {
		details = append(details, apierror.Detail{
			Field:       "code",
			Constraints: []string{"code must not be empty"},
		})
	}
	if input.DiscountType != nil && strings.TrimSpace(*input.DiscountType) == "" {
		details = append(details, apierror.Detail{
			Field:       "discountType",
			Constraints: []string{"discountType must not be empty"},
		})
	}
	if input.DiscountValue != nil && *input.DiscountValue <= 0 {
		details = append(details, apierror.Detail{
			Field:       "discountValue",
			Constraints: []string{"discountValue must be greater than 0"},
		})
	}
	if input.Status != nil && strings.TrimSpace(*input.Status) == "" {
		details = append(details, apierror.Detail{
			Field:       "status",
			Constraints: []string{"status must not be empty"},
		})
	}

	startsAt, endsAt := currentDateWindow(input)
	details = append(details, validateDateWindow(startsAt, endsAt)...)

	if len(details) > 0 {
		return apierror.Validation(details)
	}

	return nil
}

func currentDateWindow(input promotions.UpdateInput) (*time.Time, *time.Time) {
	var startsAt *time.Time
	var endsAt *time.Time

	if input.StartsAt.Set {
		startsAt = input.StartsAt.Value
	}
	if input.EndsAt.Set {
		endsAt = input.EndsAt.Value
	}

	if !input.StartsAt.Set || !input.EndsAt.Set {
		return startsAt, endsAt
	}

	return startsAt, endsAt
}

func validateDateWindow(startsAt, endsAt *time.Time) []apierror.Detail {
	if startsAt == nil || endsAt == nil {
		return nil
	}

	if startsAt.After(*endsAt) {
		return []apierror.Detail{{
			Field:       "startsAt",
			Constraints: []string{"startsAt must be on or before endsAt"},
		}}
	}

	return nil
}
