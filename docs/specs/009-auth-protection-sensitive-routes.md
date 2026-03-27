# Spec 009: Auth Middleware Stub + Protect Promotions Endpoints

## Goal
Introduce baseline auth middleware and protect promotions routes.

## Scope
- Add authentication middleware/guard abstraction with a deterministic test stub strategy.
- Protect all promotions endpoints from unauthenticated access.
- Keep products/categories/orders behavior unchanged unless explicitly required.

## Out of scope
- Full identity provider integration.
- Fine-grained RBAC/ABAC policy matrix.

## Acceptance criteria
- Promotions endpoints reject unauthenticated requests with `401` (or stack-equivalent).
- Unauthorized role/permission path returns `403` when applicable.
- Non-protected endpoints remain accessible.
- Tests verify authenticated and unauthenticated paths.

## Verification
- Integration/e2e tests with and without auth context.
