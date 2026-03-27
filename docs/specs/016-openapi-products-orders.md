# Spec 016: OpenAPI Annotations for Products + Orders

## Goal
Document stabilized API surface for products and orders.

## Scope
- Add OpenAPI (or stack-equivalent API docs) setup in app startup.
- Document handlers and request/response schemas for:
  - Products endpoints
  - Orders list/detail/cancel endpoints
- Include examples for common request/response payloads.
- Group by tags and version context (`v1`).

## Out of scope
- Full documentation for every route group.
- SDK generation.

## Documentation requirements
- Include error response schema examples aligned with global envelope.
- Show auth requirement on promotions endpoints only if already implemented.
- Mark deprecated request field alias (`sku`) in docs notes.

## Acceptance criteria
- Docs endpoint (for example `/docs`) serves API docs UI/JSON.
- Products and orders routes appear with parameters and schemas.
- Cancel endpoint is documented as a state transition operation.

## Verification
- Manual check of generated docs UI/JSON.
- Optional snapshot test for API docs stability.
