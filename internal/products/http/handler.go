package http

import (
	"context"
	"net/http"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/pagination"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/validation"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products"
)

type writer interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Service interface {
	List(context.Context, products.ListInput) (products.ListResult, error)
	Search(context.Context, products.SearchInput) ([]products.Product, error)
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
	h.register(mux, nil)
}

func (h *Handler) RegisterWithWriteMiddleware(mux *http.ServeMux, middleware func(http.Handler) http.Handler) {
	h.register(mux, middleware)
}

func (h *Handler) register(mux *http.ServeMux, writeMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /v1/products", apierror.Adapt(h.list))
	mux.Handle("GET /v1/search/products", apierror.Adapt(h.search))
	mux.Handle("POST /v1/products", wrap(apierror.Adapt(h.create), writeMiddleware))
	mux.Handle("GET /v1/products/{id}", apierror.Adapt(h.getByID))
	mux.Handle("PATCH /v1/products/{id}", wrap(apierror.Adapt(h.update), writeMiddleware))
	mux.Handle("DELETE /v1/products/{id}", wrap(apierror.Adapt(h.delete), writeMiddleware))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) error {
	params, err := pagination.Parse(r.URL.Query())
	if err != nil {
		return err
	}

	items, err := h.service.List(r.Context(), products.ListInput{
		Page:  params.Page,
		Limit: params.Limit,
	})
	if err != nil {
		return err
	}

	responseItems := make([]productResponse, 0, len(items.Items))
	for _, item := range items.Items {
		responseItems = append(responseItems, newProductResponse(item))
	}

	return writeJSON(w, http.StatusOK, listProductsResponse{
		Items:    responseItems,
		Metadata: pagination.NewMetadata(params, items.Total),
	})
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

func (h *Handler) search(w http.ResponseWriter, r *http.Request) error {
	items, err := h.service.Search(r.Context(), products.SearchInput{
		Query: r.URL.Query().Get("q"),
	})
	if err != nil {
		return err
	}

	responseItems := make([]productResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, newProductResponse(item))
	}

	return writeJSON(w, http.StatusOK, searchProductsResponse{
		Items: responseItems,
	})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) error {
	var request createProductRequest
	if err := validation.DecodeJSON(w, r, &request); err != nil {
		return err
	}

	input, err := request.toCreateInput()
	if err != nil {
		return err
	}

	item, err := h.service.Create(r.Context(), input)
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

	input, err := request.toUpdateInput()
	if err != nil {
		return err
	}

	item, err := h.service.Update(r.Context(), id, input)
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

func wrap(handler http.Handler, middleware func(http.Handler) http.Handler) http.Handler {
	if middleware == nil {
		return handler
	}

	return middleware(handler)
}
