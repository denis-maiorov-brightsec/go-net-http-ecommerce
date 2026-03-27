//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/api"
)

const defaultIntegrationDatabaseURL = "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable"

var (
	integrationPoolOnce sync.Once
	integrationPool     *pgxpool.Pool
	integrationPoolErr  error
)

func newIntegrationRouter(t *testing.T) *api.Dependencies {
	t.Helper()

	pool := integrationDBPool(t)
	resetIntegrationDatabase(t, pool)

	return &api.Dependencies{DB: pool}
}

func integrationDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	integrationPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cfg, err := pgxpool.ParseConfig(integrationDatabaseURL())
		if err != nil {
			integrationPoolErr = fmt.Errorf("parse integration database url: %w", err)
			return
		}

		integrationPool, integrationPoolErr = pgxpool.NewWithConfig(ctx, cfg)
		if integrationPoolErr != nil {
			integrationPoolErr = fmt.Errorf("open integration database pool: %w", integrationPoolErr)
			return
		}

		if err := integrationPool.Ping(ctx); err != nil {
			integrationPool.Close()
			integrationPoolErr = fmt.Errorf("ping integration database: %w", err)
			return
		}
	})

	if integrationPoolErr != nil {
		t.Fatalf("integration database unavailable: %v", integrationPoolErr)
	}

	return integrationPool
}

func resetIntegrationDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statements := []string{
		`DROP SCHEMA IF EXISTS public CASCADE`,
		`CREATE SCHEMA public`,
		`GRANT ALL ON SCHEMA public TO postgres`,
		`GRANT ALL ON SCHEMA public TO public`,
	}
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatalf("reset integration database: %v", err)
		}
	}

	for _, path := range migrationFiles(t) {
		query, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}

		if _, err := pool.Exec(ctx, string(query)); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(path), err)
		}
	}
}

func migrationFiles(t *testing.T) []string {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(repoRoot(t), "db", "migrations", "*.up.sql"))
	if err != nil {
		t.Fatalf("glob migration files: %v", err)
	}

	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatal("expected at least one migration file")
	}

	return paths
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repo root: runtime caller unavailable")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func integrationDatabaseURL() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}

	return defaultIntegrationDatabaseURL
}
