package requestlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

const HeaderName = "X-Request-ID"

type contextKey string

const requestIDContextKey contextKey = "request-id"

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}

func Middleware(logger *slog.Logger, next http.Handler) http.Handler {
	if next == nil {
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}

	if logger == nil {
		logger = slog.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := requestIDFromHeader(r)
		if requestID == "" {
			requestID = newRequestID()
		}

		r = r.WithContext(context.WithValue(r.Context(), requestIDContextKey, requestID))

		startedAt := time.Now()
		recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		recorder.Header().Set(HeaderName, requestID)

		defer func() {
			logger.Info("http request completed",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"latency_ms", time.Since(startedAt).Milliseconds(),
			)
		}()

		next.ServeHTTP(recorder, r)
	})
}

func requestIDFromHeader(r *http.Request) string {
	if r == nil {
		return ""
	}

	return r.Header.Get(HeaderName)
}

func newRequestID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}

	return hex.EncodeToString(value[:])
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
