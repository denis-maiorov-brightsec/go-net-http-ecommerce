# E-commerce Spec-Driven Scaffold Template (Language/Framework Agnostic)

This repository is a reusable template for building ecommerce backoffice APIs through iterative specs with Codex agents.

It is language/framework/ORM agnostic, keeps an ecommerce-focused spec progression, and drives execution from `docs/STACK_PROFILE.md`.

## What this template includes

- `AGENTS.md`: execution protocol and guardrails for spec-by-spec implementation
- `docs/STACK_PROFILE.md`: single place for stack/tooling choices and quality-gate commands (not API contracts)
- `docs/SPECS_INDEX.md`: dependency-ordered backlog with statuses
- `docs/specs/*.md`: implementation-agnostic spec seed set
- `scripts/run-specs-harness.mjs`: implementer + reviewer two-pass automation harness
- `prompts/*.md`: copy/paste prompts to adapt specs and scaffold a fresh project

## Expected usage

1. Fill `docs/STACK_PROFILE.md` with your target stack, commands, and exact repository topology paths.
2. Use `prompts/01-adapt-specs-and-scaffold.md` as the first Codex prompt.
3. Run the harness per spec (or range).

## Harness quick start

```bash
node scripts/run-specs-harness.mjs --dry-run
```

Run up to 3 specs:

```bash
node scripts/run-specs-harness.mjs \
  --max-specs 3
```

## Notes

- Harness expects `docs/SPECS_INDEX.md` table format and `docs/specs/<id>-*.md` files.
- By default, harness enforces clean git state and branch `main`.
- Implementer must create a commit; reviewer commits only when fixes are required.
