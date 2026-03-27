package main

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/config"
)

func TestNewDependenciesPropagatesWriteRateLimitConfig(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Config{
		WriteRateLimitRequests: 7,
		WriteRateLimitWindow:   15 * time.Second,
	}

	deps := newDependencies(cfg, logger)

	if deps.Logger != logger {
		t.Fatal("expected logger to be propagated")
	}

	if deps.WriteRateLimitRequests != cfg.WriteRateLimitRequests {
		t.Fatalf("expected write rate limit requests %d, got %d", cfg.WriteRateLimitRequests, deps.WriteRateLimitRequests)
	}

	if deps.WriteRateLimitWindow != cfg.WriteRateLimitWindow {
		t.Fatalf("expected write rate limit window %s, got %s", cfg.WriteRateLimitWindow, deps.WriteRateLimitWindow)
	}
}
