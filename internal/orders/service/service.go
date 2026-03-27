package service

import (
	"context"
	"errors"
	"strings"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	ordersrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/repository"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type Repository interface {
	List(context.Context, orders.ListInput) (orders.ListResult, error)
	GetByID(context.Context, int64) (orders.Order, error)
	Cancel(context.Context, int64) (orders.Order, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, input orders.ListInput) (orders.ListResult, error) {
	if err := validateListInput(input); err != nil {
		return orders.ListResult{}, err
	}

	items, err := s.repo.List(ctx, input)
	if err != nil {
		return orders.ListResult{}, apierror.Internal(err)
	}

	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (orders.Order, error) {
	if err := validateID(id); err != nil {
		return orders.Order{}, err
	}

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ordersrepository.ErrNotFound) {
			return orders.Order{}, apierror.NotFound("Order not found")
		}

		return orders.Order{}, apierror.Internal(err)
	}

	return item, nil
}

func (s *Service) Cancel(ctx context.Context, id int64) (orders.Order, error) {
	if err := validateID(id); err != nil {
		return orders.Order{}, err
	}

	item, err := s.repo.Cancel(ctx, id)
	if err != nil {
		if errors.Is(err, ordersrepository.ErrNotFound) {
			return orders.Order{}, apierror.NotFound("Order not found")
		}
		if errors.Is(err, ordersrepository.ErrIneligibleStatus) {
			return orders.Order{}, apierror.New(409, "CONFLICT", "Order cannot be cancelled from current status", nil)
		}

		return orders.Order{}, apierror.Internal(err)
	}

	return item, nil
}

func validateListInput(input orders.ListInput) error {
	details := make([]apierror.Detail, 0, 2)

	if strings.TrimSpace(input.Status) != "" && strings.TrimSpace(input.Status) != input.Status {
		details = append(details, apierror.Detail{
			Field:       "status",
			Constraints: []string{"status must not include leading or trailing whitespace"},
		})
	}
	if input.From != nil && input.To != nil && input.From.After(*input.To) {
		details = append(details, apierror.Detail{
			Field:       "from",
			Constraints: []string{"from must be on or before to"},
		})
	}

	if len(details) > 0 {
		return apierror.Validation(details)
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
