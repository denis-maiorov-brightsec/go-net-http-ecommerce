# Spec 005: Shared Pagination Helper + Products Refactor

## Goal
Consolidate list-query pagination behavior into a shared reusable helper.

## Scope
- Introduce shared pagination parser/validator and response metadata helper.
- Refactor `GET /v1/products` to use shared pagination helper.
- Adopt stable default page/limit and max limit enforcement.

## Out of scope
- Rewriting unrelated list routes not covered by this spec.

## Acceptance criteria
- `GET /v1/products` supports `page` and `limit` query params.
- Response includes pagination metadata (`page`, `limit`, `total`, `totalPages`).
- Invalid pagination params return `400` with standard envelope.
- Existing list behavior remains backward compatible where possible.

## Verification
- Integration/e2e tests for valid/invalid pagination combinations.
