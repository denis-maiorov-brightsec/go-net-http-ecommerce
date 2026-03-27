package commands

import (
	"context"
	"errors"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type Writer interface {
	Cancel(context.Context, int64) (orders.Order, error)
}

type Service struct {
	repo Writer
}

func NewService(repo Writer) *Service {
	return &Service{repo: repo}
}

func (s *Service) Cancel(ctx context.Context, id int64) (orders.Order, error) {
	if err := validateID(id); err != nil {
		return orders.Order{}, err
	}

	item, err := s.repo.Cancel(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return orders.Order{}, apierror.NotFound("Order not found")
		}
		if errors.Is(err, ErrIneligibleStatus) {
			return orders.Order{}, apierror.New(409, "CONFLICT", "Order cannot be cancelled from current status", nil)
		}

		return orders.Order{}, apierror.Internal(err)
	}

	return item, nil
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
