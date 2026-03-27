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

## Go/net/http implementation notes
- Create `internal/orders/` with explicit read-path handler/service/repository code. The later `queries/commands` split is deferred to spec `014`.
- Persist orders and line items in PostgreSQL and express list filtering through repository queries.
- Reuse pagination and envelope helpers already introduced by earlier specs instead of adding order-specific variants.

## Out of scope
- Order creation workflow.
- Order cancellation transition (spec 007).

## Acceptance criteria
- List endpoint supports filters and pagination conventions from spec 005.
- Detail endpoint returns `404` for missing order id.
- Date filter validation uses spec 002 error envelope.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
