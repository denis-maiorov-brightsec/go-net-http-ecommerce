package api

import (
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Logger *slog.Logger
	DB     *pgxpool.Pool
}

func NewRouter(deps Dependencies) http.Handler {
	_ = deps

	mux := http.NewServeMux()

	return mux
}
