# Spec 004: Add `/v1/categories` CRUD

## Goal
Add categories resource to support product grouping.

## Scope
- Endpoints:
  - `GET /v1/categories`
  - `GET /v1/categories/:id`
  - `POST /v1/categories`
  - `PATCH /v1/categories/:id`
  - `DELETE /v1/categories/:id`
- Category fields:
  - `id`, `name`, `slug`, `description?`, `createdAt`, `updatedAt`

## Go/net/http implementation notes
- Create `internal/categories/` with explicit HTTP, service, and repository packages.
- Back the runtime behavior with PostgreSQL tables and SQL migrations; do not use in-memory storage.
- Keep duplicate-slug handling deterministic in the repository/service boundary so HTTP handlers stay thin.

## Out of scope
- Cascading delete strategy for related products unless explicitly required.

## Behavior rules
- `slug` is unique.
- Duplicate `slug` returns conflict (`409`) with standard envelope.
- `DELETE` semantics match products behavior from spec 003.

## Acceptance criteria
- Full CRUD behavior implemented and validated.
- Error envelope and validation behavior consistent with spec 002.
- Integration tests cover duplicate slug and not-found paths.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
