# Spec 015: Add Contract Tests for Products Routes

## Goal
Harden products API with contract-style integration/e2e coverage after iterative changes.

## Scope
- Add/expand contract tests for:
  - `GET /v1/products`
  - `GET /v1/products/:id`
  - `POST /v1/products`
  - `PATCH /v1/products/:id`
  - `DELETE /v1/products/:id`
  - `GET /v1/search/products`
- Assert response shapes, status codes, and key validation behaviors.

## Out of scope
- Consumer-driven pact tooling.
- Contract tests for non-products route groups.

## Must-cover scenarios
- Valid create/update/delete flow.
- Validation error envelope compliance.
- Not-found behavior.
- `stockKeepingUnit` canonical response contract.
- Deprecated `sku` request alias acceptance.

## Acceptance criteria
- Contract tests fail on schema/status regressions.
- Tests are deterministic and isolated.
- Existing baseline integration/e2e tests continue to pass.

## Verification
- Run integration/e2e test command from stack profile.
