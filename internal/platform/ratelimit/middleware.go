package ratelimit

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type Middleware struct {
	now      func() time.Time
	window   time.Duration
	limit    int
	mu       sync.Mutex
	counters map[string]counter
}

type counter struct {
	windowStart time.Time
	requests    int
}

func New(limit int, window time.Duration, now func() time.Time) *Middleware {
	if now == nil {
		now = time.Now
	}

	return &Middleware{
		now:      now,
		window:   window,
		limit:    limit,
		counters: make(map[string]counter),
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if m == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.allow(clientKey(r)) {
			next.ServeHTTP(w, r)
			return
		}

		apierror.Write(w, r, apierror.New(http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Rate limit exceeded", nil))
	})
}

func (m *Middleware) allow(key string) bool {
	if m == nil || m.limit <= 0 || m.window <= 0 {
		return true
	}

	now := m.now().UTC()
	windowStart := truncateWindow(now, m.window)

	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.counters[key]
	if current.windowStart.IsZero() || !current.windowStart.Equal(windowStart) {
		current = counter{windowStart: windowStart}
	}

	if current.requests >= m.limit {
		m.counters[key] = current
		return false
	}

	current.requests++
	m.counters[key] = current
	return true
}

func clientKey(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("Authorization")); token != "" {
		return "auth:" + token
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return "ip:" + host
	}

	if addr := strings.TrimSpace(r.RemoteAddr); addr != "" {
		return "ip:" + addr
	}

	return "ip:unknown"
}

func truncateWindow(now time.Time, window time.Duration) time.Time {
	if window <= 0 {
		return now
	}

	span := int64(window)
	if span <= 0 {
		return now
	}

	return time.Unix(0, now.UnixNano()-(now.UnixNano()%span)).UTC()
}
