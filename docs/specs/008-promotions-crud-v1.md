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
- Run resource-specific and e2e tests.
