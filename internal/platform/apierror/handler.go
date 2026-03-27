package apierror

import (
	"fmt"
	"log/slog"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func Adapt(next HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next(w, r); err != nil {
			Write(w, r, err)
		}
	})
}

func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if logger != nil {
					logger.Error("recovered panic from request", "path", r.URL.Path, "panic", recovered)
				}

				Write(w, r, Internal(fmt.Errorf("panic: %v", recovered)))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
