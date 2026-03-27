package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
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
	v1Mux := http.NewServeMux()

	v1Mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	})

	mux.Handle("/v1/", http.StripPrefix("/v1", v1Mux))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Deprecation", "true")
		writeJSON(w, http.StatusOK, deprecatedRootResponse{
			Message: "This unversioned root is deprecated. Migrate to /v1/health.",
		})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
