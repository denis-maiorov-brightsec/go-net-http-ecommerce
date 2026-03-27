# Stack Profile

Use this file for stack/tooling decisions only; API behavior/contracts are defined by specs.

## Product Context
- Project name: `go-net-http-ecommerce`
- Domain: `ecommerce-backoffice`
- API style: `REST`

## Core Tech Choices
- Language: `Go`
- Runtime: `Go 1.24`
- Framework: `none`; use the standard library `net/http` stack with `http.ServeMux`
- ORM / Data Mapper: `pgx/v5` for PostgreSQL access, with `scany/pgxscan` as a lightweight row mapper; no ORM
- Database: `PostgreSQL 16+`

## Repository Conventions
- Package/dependency manager: `go mod`
- Migration strategy: `golang-migrate` with plain SQL migration files in `db/migrations/`; migrations are applied explicitly, never implicitly on app startup
- Configuration style: environment variables only, loaded into a small typed config package; `.env` support is for local development only and must stay optional

## Repository Topology Contract
- Entry point path: `cmd/api/`
- Source root path: `internal/`
- Module path pattern: `internal/<module>/...` with module-local packages such as `dto/`, `http/`, `repository/`, and `service/`; add `queries/` and `commands/` only when a spec requires that split
- Shared/common code path: `internal/platform/`
- DB/migrations path: `db/migrations/`
- Test path strategy: `co-located` + `**/*_test.go`; integration/e2e tests live under `test/integration/`
- API docs artifact path (if generated): `docs/openapi/openapi.json`
- Prohibited top-level paths (to avoid drift): `src/`, `app/`, `pkg/`, `lib/`, `misc/`

## Quality Gates
- Lint command: `go vet ./...`
- Unit test command: `go test ./...`
- Integration/e2e test command: `go test -tags=integration ./...`
- Type-check/static-analysis command: `XDG_CACHE_HOME=.cache ./bin/staticcheck ./...`

## Implementation Preferences (Optional)
- Validation library preference: prefer explicit request DTO decoding plus `go-playground/validator/v10` for field rules; avoid framework-style binding layers
- Logging library preference: `log/slog`
- API docs tool preference: checked-in OpenAPI document served by the app at `/docs`; keep docs generation/serving lightweight and framework-free
- Auth library preference: no auth framework; use small stdlib middleware around Bearer-token style auth when a spec requires it

## Additional Constraints
- Performance/security/compliance requirements: use parameterized SQL only; set server, handler, and DB timeouts explicitly; propagate `context.Context` through handlers/services/repositories; keep request bodies bounded; avoid global mutable runtime state; prefer forward-only SQL migrations
- Deployment/runtime environment: single stateless HTTP service running in a Linux container or similar environment, with PostgreSQL reached through `DATABASE_URL`; local dev entrypoint is `go run ./cmd/api`
- Backward-compatibility rules: keep existing route behavior, status codes, and JSON field names stable unless the active spec explicitly changes them; use `/v1` for public API evolution; temporary deprecated aliases are allowed only when a spec requires them

## Precedence Rules
- Specs are the source of truth for API behavior and contracts.
- If this profile conflicts with a spec, follow the spec.
