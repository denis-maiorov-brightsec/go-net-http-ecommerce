package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/validation"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

type writer interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Service interface {
	List(context.Context) ([]products.Product, error)
	GetByID(context.Context, int64) (products.Product, error)
	Create(context.Context, products.CreateInput) (products.Product, error)
	Update(context.Context, int64, products.UpdateInput) (products.Product, error)
	Delete(context.Context, int64) error
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("GET /v1/products", apierror.Adapt(h.list))
	mux.Handle("POST /v1/products", apierror.Adapt(h.create))
	mux.Handle("GET /v1/products/{id}", apierror.Adapt(h.getByID))
	mux.Handle("PATCH /v1/products/{id}", apierror.Adapt(h.update))
	mux.Handle("DELETE /v1/products/{id}", apierror.Adapt(h.delete))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.List(r.Context())
	if err != nil {
		return err
	}

	responseItems := make([]productResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, newProductResponse(item))
	}

	return writeJSON(w, http.StatusOK, listProductsResponse{Items: responseItems})
}

func (h *Handler) getByID(w http.ResponseWriter, r *http.Request) error {
	id, err := productIDFromPath(r)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newProductResponse(item))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var request createProductRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Create(r.Context(), products.CreateInput{
		Name:       request.Name,
		SKU:        request.SKU,
		Price:      request.Price,
		Status:     request.Status,
		CategoryID: request.CategoryID,
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusCreated, newProductResponse(item))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) error {
	id, err := productIDFromPath(r)
	if err != nil {
		return err
	}

	var request updateProductRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	item, err := h.service.Update(r.Context(), id, products.UpdateInput{
		Name:   request.Name,
		SKU:    request.SKU,
		Price:  request.Price,
		Status: request.Status,
		CategoryID: products.OptionalInt64{
			Set:   request.CategoryID.Set,
			Value: request.CategoryID.Value,
		},
	})
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, newProductResponse(item))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := productIDFromPath(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func productIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, apierror.Validation([]apierror.Detail{{
			Field:       "id",
			Constraints: []string{"id must be a positive integer"},
		}})
	}

	return id, nil
}
