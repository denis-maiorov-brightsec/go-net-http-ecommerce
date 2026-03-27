You are the reviewer/fixer agent for the most recently completed spec.

Tasks:
1. Identify the most recently completed spec from git history and `docs/SPECS_INDEX.md`.
2. Review implementation strictly against the spec acceptance criteria.
3. Focus on bugs, regressions, missing tests, and contract mismatches.
4. Apply focused fixes only; avoid broad refactors.
5. Run relevant tests/lint for touched areas.
6. Commit fixes if needed using `fix(spec-00x): ...` or appropriate style from `AGENTS.md`.
7. If no changes are needed, report: `No fixes required.`

Return:
- Findings by severity
- Files changed (if any)
- Commands run and outcomes
