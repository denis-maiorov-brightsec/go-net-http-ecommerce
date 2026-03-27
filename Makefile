GO ?= go
GOBIN_DIR ?= $(CURDIR)/bin

.PHONY: install-tools fmt lint test test-integration staticcheck run db-up db-down migrate-up migrate-down

install-tools:
	GOBIN=$(GOBIN_DIR) $(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	GOBIN=$(GOBIN_DIR) $(GO) install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

fmt:
	$(GO) fmt ./...

lint:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-integration:
	$(GO) test -tags=integration ./...

staticcheck:
	XDG_CACHE_HOME=$(CURDIR)/.cache $(GOBIN_DIR)/staticcheck ./...

run:
	$(GO) run ./cmd/api

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	$(GOBIN_DIR)/migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	$(GOBIN_DIR)/migrate -path db/migrations -database "$$DATABASE_URL" down 1
