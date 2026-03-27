package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/validation"
)

type writer interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Service interface {
	List(context.Context) ([]categories.Category, error)
	GetByID(context.Context, int64) (categories.Category, error)
	Create(context.Context, categories.CreateInput) (categories.Category, error)
	Update(context.Context, int64, categories.UpdateInput) (categories.Category, error)
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

func (h *Handler) RegisterWithWriteMiddleware(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	h.register(mux, middleware)
}

func (h *Handler) register(mux *http.ServeMux, writeMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /v1/categories", apierror.Adapt(h.list))
	mux.Handle("POST /v1/categories", wrap(apierror.Adapt(h.create), writeMiddleware))
	mux.Handle("GET /v1/categories/{id}", apierror.Adapt(h.getByID))
	mux.Handle("PATCH /v1/categories/{id}", wrap(apierror.Adapt(h.update), writeMiddleware))
	mux.Handle("DELETE /v1/categories/{id}", wrap(apierror.Adapt(h.delete), writeMiddleware))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.List(r.Context())
	if err != nil {
		return err
	}

	responseItems := make([]categoryResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, newCategoryResponse(item))
	}

	return writeJSON(w, http.StatusOK, listCategoriesResponse{Items: responseItems})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) error {
	id, err := categoryIDFromPath(r)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newCategoryResponse(item))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var request createCategoryRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Create(r.Context(), categories.CreateInput{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, newCategoryResponse(item))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := categoryIDFromPath(r)
	if err != nil {
		return err
	}

	var request updateCategoryRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Update(r.Context(), id, categories.UpdateInput{
		Name: request.Name,
		Slug: request.Slug,
		Description: categories.OptionalString{
			Set:   request.Description.Set,
			Value: request.Description.Value,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newCategoryResponse(item))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := categoryIDFromPath(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func categoryIDFromPath(r *http.Request) (int64, error) {
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
