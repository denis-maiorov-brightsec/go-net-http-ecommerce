# Spec 011: Request ID + Structured Request Logging Middleware

## Goal
Improve observability with per-request correlation IDs and structured logs.

## Scope
- Inject request ID (accept incoming header or generate one).
- Include request ID in response headers.
- Emit structured request logs with status, latency, method/path, and request ID.

## Go/net/http implementation notes
- Use stdlib middleware and `context.Context` propagation for the request ID.
- Emit structured logs with `log/slog`.
- Keep middleware reusable across all routes by placing it in `internal/platform/`.

## Out of scope
- Full distributed tracing integration.
- Log shipping pipeline setup.

## Acceptance criteria
- Every request has a request ID available in handlers and logs.
- Response includes request ID header.
- Logs use structured format aligned with stack conventions.
- Tests verify request ID propagation behavior.

## Verification
- Run `go test ./...`.
- Run `go test -tags=integration ./...`.
- Manual request/response header checks.
