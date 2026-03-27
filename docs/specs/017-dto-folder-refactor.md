# Spec 017: Refactor Request Schemas into `/dto` Folders and Update Imports

## Goal
Apply a late-stage structural refactor to consolidate request DTO/schema definitions.

## Scope
- Move request DTO/schema definitions into `/dto` (or stack-equivalent) folders per module.
- Update imports and wiring across products/orders/categories/promotions modules.
- Keep runtime behavior unchanged.

## Out of scope
- Contract changes to endpoints.
- Business logic changes.

## Acceptance criteria
- Application behavior and API contracts remain unchanged.
- All tests pass after import/path updates.
- Folder structure is consistent and discoverable.

## Verification
- Run full lint/type-check/test suite.
