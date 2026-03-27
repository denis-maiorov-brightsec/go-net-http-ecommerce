# Spec-Driven Implementation Agent Guide

## Objective
This repository is used to simulate iterative ecommerce backoffice API development through dependency-ordered specs.
Agents must implement work from `docs/specs/` one spec at a time, in order, with realistic churn (new endpoints, contract changes, refactors, deprecations).

Target stack is defined in:
- `docs/STACK_PROFILE.md`

## Source of truth
- Backlog order and dependencies: `docs/SPECS_INDEX.md`
- Single-task implementation details: `docs/specs/*.md`
- Stack conventions and commands: `docs/STACK_PROFILE.md` (must not override spec behavior/contracts)

## Execution protocol for every agent run
1. Pick exactly one spec that is `Ready` (or unblocked by dependencies if statuses lag behind).
2. Read the entire spec and its dependency list.
3. Implement only that spec's scope.
4. Follow stack constraints and repository topology contract from `docs/STACK_PROFILE.md`.
5. Run relevant tests and lint for touched areas.
6. Update docs/tests required by the spec.
7. Stop after completion criteria are met.

## Non-negotiable guardrails
- Do not pull in future-spec behavior unless explicitly required by backward compatibility in the current spec.
- Preserve existing API behavior unless current spec says to change it.
- Prefer small, reviewable changes with clear commit boundaries.
- If blocked by missing prerequisites, stop and mark the spec as blocked with a concrete reason.
- Implement persistence using the configured database + ORM/data layer from `docs/STACK_PROFILE.md` for all feature specs.
- Do not introduce or keep in-memory repositories for runtime feature behavior unless the spec explicitly allows it.
- Write tests against real persistence/integration boundaries when feasible; avoid mock-only feature coverage.

## Definition of done (per spec)
- All acceptance criteria from the target spec pass.
- Required tests were added/updated and pass locally.
- No unrelated refactors.
- Routes/DTO/auth/middleware/versioning behavior matches the spec exactly.
- Any deprecations are documented in code comments or docs when required.

## Branch and commit guidance
- Run specs sequentially on `main` unless project policy says otherwise.
- Do not create or switch branches during automated spec execution.
- Commit style:
  - `feat(spec-00x): ...` for behavior changes
  - `refactor(spec-00x): ...` for structural changes
  - `test(spec-00x): ...` for test-only work
  - `docs(spec-00x): ...` for docs-only follow-up

## Expected output from an implementation agent
- What changed (files + behavior)
- Acceptance criteria checklist
- Test command(s) executed and results
- Any follow-up risks or migration notes
