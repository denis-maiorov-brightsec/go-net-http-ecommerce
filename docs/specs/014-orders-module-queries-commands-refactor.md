# Spec 014: Refactor Orders Module into `queries/commands`

## Goal
Improve internal orders module structure without changing external behavior.

## Scope
- Refactor orders module to separate read-path (`queries`) and write-path (`commands`) concerns.
- Update imports/wiring and keep route contracts unchanged.

## Out of scope
- Any behavioral change to orders endpoints.
- Cross-module architectural rewrite.

## Acceptance criteria
- Existing orders endpoints behave exactly as before.
- Tests continue to pass without endpoint contract changes.
- Refactor keeps clear boundaries between query and command handlers/services.

## Verification
- Run orders-related tests and regression e2e suite.
