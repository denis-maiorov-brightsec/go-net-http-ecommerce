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
}

func TestLoadReadsOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "3s")
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
