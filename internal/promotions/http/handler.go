package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/validation"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/promotions"
)

type writer interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Service interface {
	List(context.Context) ([]promotions.Promotion, error)
	GetByID(context.Context, int64) (promotions.Promotion, error)
	Create(context.Context, promotions.CreateInput) (promotions.Promotion, error)
	Update(context.Context, int64, promotions.UpdateInput) (promotions.Promotion, error)
	Delete(context.Context, int64) error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	h.register(mux, nil)
}

func (h *Handler) RegisterProtected(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	h.register(mux, middleware)
}

func (h *Handler) register(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	mux.Handle("GET /v1/promotions", wrap(apierror.Adapt(h.list), middleware))
	mux.Handle("POST /v1/promotions", wrap(apierror.Adapt(h.create), middleware))
	mux.Handle("GET /v1/promotions/{id}", wrap(apierror.Adapt(h.getByID), middleware))
	mux.Handle("PATCH /v1/promotions/{id}", wrap(apierror.Adapt(h.update), middleware))
	mux.Handle("DELETE /v1/promotions/{id}", wrap(apierror.Adapt(h.delete), middleware))
}

func wrap(handler http.Handler, middleware func(http.Handler) http.Handler) http.Handler {
	if middleware == nil {
		return handler
	}

	return middleware(handler)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.List(r.Context())
	if err != nil {
		return err
	}

	responseItems := make([]promotionResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, newPromotionResponse(item))
	}

	return writeJSON(w, http.StatusOK, listPromotionsResponse{Items: responseItems})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) error {
	id, err := promotionIDFromPath(r)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newPromotionResponse(item))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var request createPromotionRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Create(r.Context(), promotions.CreateInput{
		Name:          request.Name,
		Code:          request.Code,
		DiscountType:  request.DiscountType,
		DiscountValue: request.DiscountValue,
		StartsAt:      request.StartsAt,
		EndsAt:        request.EndsAt,
		Status:        request.Status,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, newPromotionResponse(item))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := promotionIDFromPath(r)
	if err != nil {
		return err
	}

	var request updatePromotionRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Update(r.Context(), id, promotions.UpdateInput{
		Name:          request.Name,
		Code:          request.Code,
		DiscountType:  request.DiscountType,
		DiscountValue: request.DiscountValue,
		StartsAt: promotions.OptionalTime{
			Set:   request.StartsAt.Set,
			Value: request.StartsAt.Value,
		},
		EndsAt: promotions.OptionalTime{
			Set:   request.EndsAt.Set,
			Value: request.EndsAt.Value,
		},
		Status: request.Status,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newPromotionResponse(item))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := promotionIDFromPath(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func promotionIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, apierror.Validation([]apierror.Detail{{
			Field:       "id",
			Constraints: []string{"id must be a positive integer"},
		}})
	}

	return id, nil
}
