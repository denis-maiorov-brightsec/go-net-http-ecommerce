package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	categoryhttp "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories/http"
	categoriesrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories/repository"
	categoriesservice "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/categories/service"
	ordershttp "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/http"
	ordersrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/repository"
	ordersservice "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/orders/service"
	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
	producthttp "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products/http"
	productsrepository "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products/repository"
	productsservice "github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/products/service"
)

type Dependencies struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
}

type healthResponse struct {
	Status string `json:"status"`
}

type deprecatedRootResponse struct {
	Message string `json:"message"`
}

func NewRouter(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	categoryHandler := categoryhttp.New(categoriesservice.New(categoriesrepository.New(deps.DB)))
	orderHandler := ordershttp.New(ordersservice.New(ordersrepository.New(deps.DB)))
	productHandler := producthttp.New(productsservice.New(productsrepository.New(deps.DB)))

	mux.Handle("GET /v1/health", apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		return writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	}))
	mux.Handle("GET /{$}", apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Deprecation", "true")
		return writeJSON(w, http.StatusOK, deprecatedRootResponse{
			Message: "This unversioned root is deprecated. Migrate to /v1/health.",
		})
	}))
	categoryHandler.Register(mux)
	orderHandler.Register(mux)
	productHandler.Register(mux)

	return apierror.Recover(deps.Logger, apierror.NormalizeServeMux(mux))
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return fmt.Errorf("encode json response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err := w.Write(body.Bytes())
	return err
}
