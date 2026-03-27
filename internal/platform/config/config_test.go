package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaultsWhenEnvIsUnset(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("expected development app env, got %q", cfg.AppEnv)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("expected :8080 http addr, got %q", cfg.HTTPAddr)
	}

	if cfg.HTTPShutdownTimeout != 10*time.Second {
		t.Fatalf("expected 10s shutdown timeout, got %s", cfg.HTTPShutdownTimeout)
	}

	if cfg.WriteRateLimitRequests != 5 {
		t.Fatalf("expected default write rate limit requests 5, got %d", cfg.WriteRateLimitRequests)
	}

	if cfg.WriteRateLimitWindow != time.Minute {
		t.Fatalf("expected default write rate limit window 1m, got %s", cfg.WriteRateLimitWindow)
	}
}

func TestLoadReadsOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "3s")
	t.Setenv("WRITE_RATE_LIMIT_REQUESTS", "7")
	t.Setenv("WRITE_RATE_LIMIT_WINDOW", "15s")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("DATABASE_MAX_CONNS", "23")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("expected test app env, got %q", cfg.AppEnv)
	}

	if cfg.HTTPAddr != ":9000" {
		t.Fatalf("expected :9000 http addr, got %q", cfg.HTTPAddr)
	}

	if cfg.HTTPShutdownTimeout != 3*time.Second {
		t.Fatalf("expected 3s shutdown timeout, got %s", cfg.HTTPShutdownTimeout)
	}

	if cfg.WriteRateLimitRequests != 7 {
		t.Fatalf("expected write rate limit requests 7, got %d", cfg.WriteRateLimitRequests)
	}

	if cfg.WriteRateLimitWindow != 15*time.Second {
		t.Fatalf("expected write rate limit window 15s, got %s", cfg.WriteRateLimitWindow)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("expected database url to be set")
	}

	if cfg.DatabaseMaxConns != 23 {
		t.Fatalf("expected 23 database max conns, got %d", cfg.DatabaseMaxConns)
	}
}

func TestLoadReturnsErrorForInvalidDuration(t *testing.T) {
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestLoadReturnsErrorForNonPositiveWriteRateLimitRequests(t *testing.T) {
	t.Setenv("WRITE_RATE_LIMIT_REQUESTS", "0")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive write rate limit requests")
	}
}

func TestLoadReturnsErrorForNonPositiveWriteRateLimitWindow(t *testing.T) {
	t.Setenv("WRITE_RATE_LIMIT_WINDOW", "0s")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for non-positive write rate limit window")
	}
}
