# Spec 003: Add `/v1/products` CRUD

## Goal
Ship the first ecommerce resource route group with full CRUD and validation.

## Scope
- Create products module/router/handler/service/repository layers.
- Endpoints:
  - `GET /v1/products`
  - `GET /v1/products/:id`
  - `POST /v1/products`
  - `PATCH /v1/products/:id`
  - `DELETE /v1/products/:id`
- Product fields (initial contract):
  - `id`, `name`, `sku`, `price`, `status`, `categoryId?`, `createdAt`, `updatedAt`

## Out of scope
- Shared pagination helper abstraction (spec 005).
- Search endpoint (spec 012).
- `sku` rename (spec 013).

## Behavior rules
- `POST` validates required fields and positive price.
- `GET :id` returns `404` when missing.
- `PATCH` supports partial updates with validation.
- `DELETE` returns `204`; deleting missing id returns `404`.

## Acceptance criteria
- All five endpoints functional under `/v1`.
- Validation and error envelope from spec 002 applied.
- Tests cover happy paths + key error cases (`400`, `404`).

## Verification
- Run unit + integration/e2e tests for touched resource.
