# Spec 013: Rename `product.sku` -> `product.stockKeepingUnit`

## Goal
Introduce a realistic contract rename while preserving compatibility during transition.

## Scope
- Update product domain model and API responses to use `stockKeepingUnit` as canonical field.
- Accept deprecated request alias `sku` for backward compatibility.
- Persist with canonical storage name according to stack/data-layer conventions.

## Go/net/http implementation notes
- Apply the rename across product DTOs, service logic, repository SQL, and docs/tests.
- Keep the PostgreSQL schema and SQL aliases deterministic during the transition; use migrations where storage changes are required.
- Do not leave both fields as independent canonical values in the codebase.

## Out of scope
- Removing backward compatibility alias in this spec.
- Renaming unrelated fields.

## Compatibility rules
- Responses expose `stockKeepingUnit`.
- Requests may provide `stockKeepingUnit` or deprecated `sku`; conflict rules must be deterministic and documented.
- Deprecation should be visible in docs/comments/tests.

## Acceptance criteria
- Products CRUD and search continue to work with canonical field.
- Deprecated `sku` input path remains accepted.
- Tests cover canonical and alias behavior, including conflict/validation cases.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
