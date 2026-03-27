# Spec 012: Add `/v1/search/products?q=`

## Goal
Add a lightweight product search endpoint that reuses products service/repository logic.

## Scope
- Endpoint: `GET /v1/search/products?q=`.
- Search over product `name` and `sku` fields (case-insensitive contains by default).
- Reuse existing products data layer to avoid duplicate query logic.

## Out of scope
- Full-text search engine integration.
- Complex ranking and typo tolerance.

## Acceptance criteria
- Empty/missing `q` is validated per spec 002 conventions.
- Search returns deterministic ordering (document sorting rules).
- Response contract aligns with products list projections.
- Tests cover matches/no-matches/validation error cases.

## Verification
- Integration/e2e tests for search behavior and edge cases.
