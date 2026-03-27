# Spec 007: Add `/v1/orders/:id/cancel` State Transition

## Goal
Support explicit order cancellation as a controlled state transition.

## Scope
- Add `POST /v1/orders/:id/cancel` (or equivalent mutation endpoint in target API style).
- Enforce allowed transitions (for example: only `pending` -> `cancelled`).
- Persist transition timestamp/audit fields if project conventions require.

## Go/net/http implementation notes
- Keep the endpoint as a plain `net/http` mutation route under the orders module.
- Model transition checks in service/command code, not inline in the handler.
- Persist transition updates atomically in PostgreSQL; use a transaction when the repository needs to update multiple tables or audit fields.

## Out of scope
- Refund/payment side effects.
- Event bus integrations.

## Acceptance criteria
- Canceling eligible order returns success with updated order.
- Canceling non-existing order returns `404`.
- Canceling ineligible status returns conflict/business-rule error (`409` or equivalent).
- Transition rules are covered by deterministic tests.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
