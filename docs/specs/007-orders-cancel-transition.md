# Spec 007: Add `/v1/orders/:id/cancel` State Transition

## Goal
Support explicit order cancellation as a controlled state transition.

## Scope
- Add `POST /v1/orders/:id/cancel` (or equivalent mutation endpoint in target API style).
- Enforce allowed transitions (for example: only `pending` -> `cancelled`).
- Persist transition timestamp/audit fields if project conventions require.

## Out of scope
- Refund/payment side effects.
- Event bus integrations.

## Acceptance criteria
- Canceling eligible order returns success with updated order.
- Canceling non-existing order returns `404`.
- Canceling ineligible status returns conflict/business-rule error (`409` or equivalent).
- Transition rules are covered by deterministic tests.

## Verification
- Integration/e2e tests for valid and invalid transition scenarios.
