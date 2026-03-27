# go-net-http-ecommerce

This repository is a spec-driven ecommerce backoffice API scaffold built around bare-bones Go: `net/http`, `http.ServeMux`, PostgreSQL, `pgx`, plain SQL migrations, and explicit package boundaries.

The development flow stays spec-first. `docs/STACK_PROFILE.md` defines the stack and repo contract, `docs/SPECS_INDEX.md` defines the dependency order, and `docs/specs/*.md` define each increment of behavior.

## Repository layout

- `cmd/api/`: application entrypoint
- `internal/api/`: top-level HTTP route wiring
- `internal/platform/`: config, HTTP server, and PostgreSQL helpers
- `db/migrations/`: SQL migration files
- `test/integration/`: integration/e2e-oriented HTTP tests
- `docs/`: stack profile, backlog, and per-spec requirements
- `scripts/run-specs-harness.mjs`: spec automation harness

## Local development

1. Copy `.env.example` to `.env` if you want local defaults.
2. Start PostgreSQL:

```bash
docker compose up -d postgres
```

3. Run the API:

```bash
go run ./cmd/api
```

The baseline scaffold intentionally does not implement spec `001` yet. The server wiring, config, DB package, migration path, and tests are in place so the next spec pass can start from real code instead of placeholders.

## Quality gates

```bash
gofmt -w $(find cmd internal test -name '*.go')
go vet ./...
go test ./...
go test -tags=integration ./...
XDG_CACHE_HOME=.cache ./bin/staticcheck ./...
```

Convenience targets are also available through `make`.

## Local tools

Install repo-local tooling:

```bash
make install-tools
```

This installs:
- `./bin/staticcheck`
- `./bin/migrate`

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
