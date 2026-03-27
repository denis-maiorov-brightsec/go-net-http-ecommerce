# Spec 008: Add `/v1/promotions` CRUD

## Goal
Introduce promotions management for ecommerce campaigns.

## Scope
- Endpoints:
  - `GET /v1/promotions`
  - `GET /v1/promotions/:id`
  - `POST /v1/promotions`
  - `PATCH /v1/promotions/:id`
  - `DELETE /v1/promotions/:id`
- Promotion fields (minimum):
  - `id`, `name`, `code`, `discountType`, `discountValue`, `startsAt?`, `endsAt?`, `status`, `createdAt`, `updatedAt`

## Go/net/http implementation notes
- Create `internal/promotions/` with explicit HTTP, service, and repository packages.
- Persist runtime behavior in PostgreSQL with migration-backed schema changes.
- Keep date-window validation in DTO/service code and uniqueness enforcement deterministic across repository + HTTP error mapping.

## Out of scope
- Promotion engine integration with checkout/order totals.

## Behavior rules
- Promotion `code` is unique.
- Date windows validate `startsAt <= endsAt` when both are provided.
- Deleting missing promotion returns `404`.

## Acceptance criteria
- Full CRUD behavior implemented under `/v1/promotions`.
- Validation and standard error envelope applied.
- Integration tests cover duplicate code and invalid date-window scenarios.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
