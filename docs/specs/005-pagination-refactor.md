# Spec 005: Shared Pagination Helper + Products Refactor

## Goal
Consolidate list-query pagination behavior into a shared reusable helper.

## Scope
- Introduce shared pagination parser/validator and response metadata helper.
- Refactor `GET /v1/products` to use shared pagination helper.
- Adopt stable default page/limit and max limit enforcement.

## Go/net/http implementation notes
- Put reusable pagination parsing/response helpers under `internal/platform/` rather than duplicating them per module.
- Keep the helper HTTP-agnostic where practical: parse from query values, then let handlers apply it.
- Refactor only products in this spec; do not preemptively change other list routes.

## Out of scope
- Rewriting unrelated list routes not covered by this spec.

## Acceptance criteria
- `GET /v1/products` supports `page` and `limit` query params.
- Response includes pagination metadata (`page`, `limit`, `total`, `totalPages`).
- Invalid pagination params return `400` with standard envelope.
- Existing list behavior remains backward compatible where possible.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
