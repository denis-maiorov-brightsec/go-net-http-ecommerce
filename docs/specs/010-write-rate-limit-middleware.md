# Spec 010: Write-Route Rate Limiting Middleware

## Goal
Protect write-heavy endpoints from abuse using basic rate limiting.

## Scope
- Apply rate limiting middleware to mutating endpoints (`POST`, `PATCH`, `DELETE`, and explicit state transitions).
- Keep read-only endpoints unaffected unless explicitly requested.
- Standardize limit-exceeded response via spec 002 envelope.

## Out of scope
- Distributed/global rate limiting for multi-node deployments unless already required.

## Acceptance criteria
- Exceeding rate threshold returns `429` with standard error envelope.
- Read-only routes remain unaffected.
- Configuration is environment-aware (defaults + overrides).
- Tests validate threshold behavior deterministically.

## Verification
- Integration tests for throttled and non-throttled paths.
