# Stack Profile

Fill this file before starting implementation.
Use it for stack/tooling decisions only; API behavior/contracts are defined by specs.

## Product Context
- Project name: `<project-name>`
- Domain: `<ecommerce-backoffice | other>`
- API style: `<REST | GraphQL | RPC | mixed>`

## Core Tech Choices
- Language: `<language>`
- Runtime: `<runtime>`
- Framework: `<framework>`
- ORM / Data Mapper: `<orm-or-data-layer>`
- Database: `<database>`

## Repository Conventions
- Package/dependency manager: `<tool>`
- Migration strategy: `<tooling + process>`
- Configuration style: `<env/config conventions>`

## Repository Topology Contract
- Source root path: `<exact path, e.g. src/>`
- Module path pattern: `<exact pattern, e.g. src/<module>/...>`
- Shared/common code path: `<exact path>`
- DB/migrations path: `<exact path>`
- Test path strategy: `<co-located | separate>` + `<exact path/pattern>`
- API docs artifact path (if generated): `<exact path>`
- Prohibited top-level paths (to avoid drift): `<list>`

Use concrete paths, not abstract descriptions. Example:
- Source root path: `src/`
- Module path pattern: `src/<module>/{controller,service,repository,dto}`
- Shared/common code path: `src/common/`
- DB/migrations path: `src/db/migrations/`
- Test path strategy: `separate` + `test/**/*.test.ts`
- API docs artifact path: `docs/openapi.json`
- Prohibited top-level paths: `lib/`, `misc/`

## Quality Gates
- Lint command: `<command>`
- Unit test command: `<command>`
- Integration/e2e test command: `<command>`
- Type-check/static-analysis command: `<command>`

## Implementation Preferences (Optional)
- Validation library preference (for example `zod` or `ajv`): `<optional>`
- Logging library preference: `<optional>`
- API docs tool preference (OpenAPI/Swagger/etc): `<optional>`
- Auth library preference: `<optional>`

## Additional Constraints
- Performance/security/compliance requirements: `<constraints>`
- Deployment/runtime environment: `<environment>`
- Backward-compatibility rules: `<rules>`

## Precedence Rules
- Specs are the source of truth for API behavior and contracts.
- If this profile conflicts with a spec, follow the spec.
