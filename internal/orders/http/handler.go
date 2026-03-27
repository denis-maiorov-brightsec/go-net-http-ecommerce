package http

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/pagination"
)

type writer interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Service interface {
	List(context.Context, orders.ListInput) (orders.ListResult, error)
	GetByID(context.Context, int64) (orders.Order, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /v1/orders", apierror.Adapt(h.list))
	mux.Handle("GET /v1/orders/{id}", apierror.Adapt(h.getByID))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	params, err := pagination.Parse(r.URL.Query())
	if err != nil {
		return err
	}

	filters, err := parseListFilters(r.URL.Query())
	if err != nil {
		return err
	}

	items, err := h.service.List(r.Context(), orders.ListInput{
		Page:   params.Page,
		Limit:  params.Limit,
		Status: filters.Status,
		From:   filters.From,
		To:     filters.To,
	})
	if err != nil {
		return err
	}

	responseItems := make([]orderResponse, 0, len(items.Items))
	for _, item := range items.Items {
		responseItems = append(responseItems, newOrderResponse(item))
	}

	return writeJSON(w, http.StatusOK, listOrdersResponse{
		Items:    responseItems,
		Metadata: pagination.NewMetadata(params, items.Total),
	})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) error {
	id, err := orderIDFromPath(r)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newOrderResponse(item))
}

type listFilters struct {
	Status string
	From   *time.Time
	To     *time.Time
}

func parseListFilters(values url.Values) (listFilters, error) {
	filters := listFilters{
		Status: strings.TrimSpace(values.Get("status")),
	}

	var details []apierror.Detail

	if rawFrom := values.Get("from"); rawFrom != "" {
		from, err := time.Parse(time.DateOnly, rawFrom)
		if err != nil {
			details = append(details, apierror.Detail{
				Field:       "from",
				Constraints: []string{"from must be a valid date in YYYY-MM-DD format"},
			})
		} else {
			from = from.UTC()
			filters.From = &from
		}
	}

	if rawTo := values.Get("to"); rawTo != "" {
		to, err := time.Parse(time.DateOnly, rawTo)
		if err != nil {
			details = append(details, apierror.Detail{
				Field:       "to",
				Constraints: []string{"to must be a valid date in YYYY-MM-DD format"},
			})
		} else {
			to = to.UTC().Add(24 * time.Hour)
			filters.To = &to
		}
	}

	if filters.From != nil && filters.To != nil && !filters.From.Before(*filters.To) {
		details = append(details, apierror.Detail{
			Field:       "from",
			Constraints: []string{"from must be on or before to"},
		})
	}

	if len(details) > 0 {
		return listFilters{}, apierror.Validation(details)
	}

	return filters, nil
}

func orderIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, apierror.Validation([]apierror.Detail{{
			Field:       "id",
			Constraints: []string{"id must be a positive integer"},
		}})
	}

	return id, nil
}
