package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
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
	_ = deps

	mux := http.NewServeMux()

	mux.Handle("GET /v1/health", apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
		return nil
	}))
	mux.Handle("GET /{$}", apierror.Adapt(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Deprecation", "true")
		writeJSON(w, http.StatusOK, deprecatedRootResponse{
			Message: "This unversioned root is deprecated. Migrate to /v1/health.",
		})
		return nil
	}))

	return apierror.Recover(deps.Logger, apierror.NormalizeServeMux(mux))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
