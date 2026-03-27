# Spec 006: Add `/v1/orders` List + Detail

## Goal
Introduce order read endpoints with filtering support.

## Scope
- Endpoints:
  - `GET /v1/orders`
  - `GET /v1/orders/:id`
- Order fields (minimum):
  - `id`, `status`, `customerId`, `createdAt`, `updatedAt`, `items[]`, `totalAmount`
- List filters:
  - `status`
  - date range (`from`, `to`) or stack-equivalent query naming

## Out of scope
- Order creation workflow.
- Order cancellation transition (spec 007).

## Acceptance criteria
- List endpoint supports filters and pagination conventions from spec 005.
- Detail endpoint returns `404` for missing order id.
- Date filter validation uses spec 002 error envelope.

## Verification
- Integration/e2e coverage for filtered list and missing detail paths.
