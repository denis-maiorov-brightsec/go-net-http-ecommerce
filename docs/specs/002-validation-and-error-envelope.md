# Spec 002: Global Validation Middleware + Consistent Error Envelope

## Goal
Standardize request validation and error responses before ecommerce resource APIs expand.

## Scope
- Register app-level validation middleware/pipeline.
- Add centralized error handling that normalizes all errors to one response shape.
- Ensure request schema validation errors are mapped into structured `details`.

## Go/net/http implementation notes
- Implement validation and error writing in explicit helper/middleware packages under `internal/platform/`; do not rely on framework binding.
- Decode request bodies with `encoding/json`, validate DTOs explicitly, and surface envelope writes from one shared path.
- Keep the error contract independent from feature modules so products/categories/orders/promotions can reuse it without copy-paste.

## Out of scope
- Request ID injection (added later in spec 011).
- Request logging middleware.

## Response contract
Error response format:

```json
{
  "timestamp": "2026-01-01T00:00:00.000Z",
  "path": "/v1/products",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      { "field": "name", "constraints": ["name must not be empty"] }
    ]
  }
}
```

## Acceptance criteria
- Invalid payloads return `400` with the envelope above (same top-level shape for all errors).
- Missing routes return `404` with same envelope format.
- Unknown runtime errors return `500` with sanitized message.
- Existing happy-path routes still work unchanged.

## Verification
- Unit/integration/e2e tests for:
  - Validation failure path
  - Not found path
  - Generic internal error mapping
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
