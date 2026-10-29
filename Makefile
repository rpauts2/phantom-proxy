# PhantomProxy Go workflow.
# The supported launch path is the isolated Go simulation lab. Legacy commands
# are deliberately explicit while the rest of the repository is audited/migrated.

.PHONY: help install install-go build run test test-lab test-go clean docker-build docker-stop docker-legacy lint lint-lab lint-go fmt check dev lab legacy-build legacy-run legacy-health status

GO ?= go
DOCKER ?= docker
DOCKER_COMPOSE ?= docker compose
BINARY ?= simulation-lab
ENTRYPOINT ?= ./cmd/simulation-lab
LAB_CAMPAIGNS ?= demo
SCOPE_DURATION ?= 8h
LAB_DATA ?= .phantom-lab/events.json

help:
	@echo "PhantomProxy - Go-first commands"
	@echo ""
	@echo "  make install       - Verify Go and download modules"
	@echo "  make build         - Build the supported Go simulation lab"
	@echo "  make run           - Build and run the local Go simulation lab"
	@echo "  make lab           - Run the Go simulation lab without building a binary"
	@echo "  LAB_DATA           - Path for persistent synthetic-event snapshot"
	@echo "  make test          - Test the supported Go simulation lab with race detection"
	@echo "  make test-go       - Test all legacy Go packages (may expose existing issues)"
	@echo "  make lint          - Vet the supported Go simulation lab"
	@echo "  make lint-go       - Vet all legacy Go packages"
	@echo "  make fmt           - Format supported Go sources"
	@echo "  make check         - Format, vet, test and build the supported lab"
	@echo "  make docker-build  - Build the standalone Go lab image"
	@echo "  make docker-legacy - Start the legacy multi-runtime Compose stack"
	@echo "  make legacy-build  - Build the old PhantomProxy v14 command"
	@echo "  make legacy-run    - Run the old command with config.yaml"
	@echo "  make clean         - Remove local build artifacts"
	@echo "  make status        - Show Go and Docker versions"
	@echo ""
	@echo "See docs/GO_ONLY_MIGRATION.md for the full migration status."

install-go:
	@command -v $(GO) >/dev/null 2>&1 || { echo "Install Go from https://go.dev/dl/"; exit 1; }
	@$(GO) version

install: install-go
	@$(GO) mod download

build:
	@$(GO) build -trimpath -o $(BINARY) $(ENTRYPOINT)

run: build
	@./$(BINARY) -campaigns "$(LAB_CAMPAIGNS)" -scope-duration "$(SCOPE_DURATION)" -data "$(LAB_DATA)"

lab:
	@$(GO) run ./cmd/simulation-lab -campaigns "$(LAB_CAMPAIGNS)" -scope-duration "$(SCOPE_DURATION)" -data "$(LAB_DATA)"

test: test-lab

test-lab:
	@$(GO) test -race ./internal/simulation -timeout 2m

# Full-repository checks remain available for migration work, but are not the
# release gate for the supported simulation-only product yet.
test-go:
	@$(GO) test -race ./... -timeout 5m

clean:
	@rm -f $(BINARY) $(BINARY).exe phantom-proxy-legacy
	@rm -rf dist/ build/

docker-build:
	@$(DOCKER) build -f Dockerfile.simulation -t phantomproxy-lab:latest .

# Explicit legacy commands; do not mistake these for the Go-only product.
docker-legacy:
	@$(DOCKER_COMPOSE) -f docker-compose.yml up -d

docker-stop:
	@$(DOCKER_COMPOSE) -f docker-compose.yml down

lint: lint-lab

lint-lab:
	@$(GO) vet ./internal/simulation ./cmd/simulation-lab

lint-go:
	@$(GO) vet ./...

fmt:
	@$(GO) fmt ./internal/simulation ./cmd/simulation-lab

check: fmt lint test build

dev:
	@$(GO) run ./cmd/simulation-lab -campaigns "$(LAB_CAMPAIGNS)" -scope-duration "$(SCOPE_DURATION)"

legacy-build:
	@$(GO) build -trimpath -o phantom-proxy-legacy ./cmd/phantom-proxy-v14

legacy-run: legacy-build
	@./phantom-proxy-legacy --config config.yaml

legacy-health:
	@bash healthcheck.sh

status:
	@echo "Go: $$($(GO) version 2>/dev/null || echo 'not installed')"
	@echo "Docker: $$($(DOCKER) --version 2>/dev/null || echo 'not installed')"
