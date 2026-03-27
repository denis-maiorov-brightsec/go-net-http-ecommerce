package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders"
	orderdto "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/dto"
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

type CommandService interface {
	Cancel(context.Context, int64) (orders.Order, error)
}

type Handler struct {
	queryService   Service
	commandService CommandService
}

func New(queryService Service, commandService CommandService) *Handler {
	return &Handler{
		queryService:   queryService,
		commandService: commandService,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	h.register(mux, nil)
}

func (h *Handler) RegisterWithWriteMiddleware(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	h.register(mux, middleware)
}

func (h *Handler) register(mux *http.ServeMux, writeMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /v1/orders", apierror.Adapt(h.list))
	mux.Handle("GET /v1/orders/{id}", apierror.Adapt(h.getByID))
	mux.Handle("POST /v1/orders/{id}/cancel", wrap(apierror.Adapt(h.cancel), writeMiddleware))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	params, err := pagination.Parse(r.URL.Query())
	if err != nil {
		return err
	}

	filters, err := orderdto.ParseListFilters(r.URL.Query())
	if err != nil {
		return err
	}

	items, err := h.queryService.List(r.Context(), orders.ListInput{
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

	item, err := h.queryService.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newOrderResponse(item))
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) error {
	id, err := orderIDFromPath(r)
	if err != nil {
		return err
	}

	item, err := h.commandService.Cancel(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newOrderResponse(item))
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

func wrap(handler http.Handler, middleware func(http.Handler) http.Handler) http.Handler {
	if middleware == nil {
		return handler
	}

	return middleware(handler)
}
