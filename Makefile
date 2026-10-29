# PhantomProxy: Go-first developer workflow.
# The application migration to a single Go runtime is tracked in docs/GO_ONLY_MIGRATION.md.

.PHONY: help install install-go build run test test-go clean docker docker-build docker-stop docker-legacy lint fmt check dev health status

GO ?= go
DOCKER ?= docker
DOCKER_COMPOSE ?= docker compose
BINARY ?= phantom-proxy
ENTRYPOINT ?= ./cmd/phantom-proxy-v14

help:
	@echo "PhantomProxy - Go developer commands"
	@echo ""
	@echo "  make install       - Download Go module dependencies"
	@echo "  make build         - Build the canonical Go entrypoint"
	@echo "  make run           - Build and run the Go entrypoint"
	@echo "  make test          - Run Go tests with race detection"
	@echo "  make test-go       - Alias for Go tests"
	@echo "  make lint          - Run go vet"
	@echo "  make fmt           - Format Go sources"
	@echo "  make check         - Run formatting check, vet, tests and build"
	@echo "  make docker-build  - Build the Go Docker image"
	@echo "  make docker-legacy - Start the legacy multi-runtime Compose stack"
	@echo "  make docker-stop   - Stop the legacy Compose stack"
	@echo "  make clean         - Remove local build artifacts"
	@echo "  make health        - Run the existing host health check"
	@echo "  make status        - Show Go and Docker versions"
	@echo ""
	@echo "Note: the legacy Compose stack is not Go-only yet; see docs/GO_ONLY_MIGRATION.md."

install-go:
	@command -v $(GO) >/dev/null 2>&1 || { echo "Install Go from https://go.dev/dl/"; exit 1; }
	@$(GO) version

install: install-go
	@$(GO) mod download

build:
	@$(GO) build -o $(BINARY) $(ENTRYPOINT)

run: build
	@./$(BINARY) --config config.yaml

test: test-go

test-go:
	@$(GO) test -race ./... -timeout 5m

clean:
	@rm -f $(BINARY) $(BINARY).exe
	@rm -rf dist/ build/

docker-build:
	@$(DOCKER) build -t phantomproxy:latest .

# Kept explicit during migration so the multi-runtime stack is not mistaken
# for the final Go-only deployment.
docker-legacy:
	@$(DOCKER_COMPOSE) -f docker-compose.yml up -d

docker-stop:
	@$(DOCKER_COMPOSE) -f docker-compose.yml down

lint:
	@$(GO) vet ./...

fmt:
	@$(GO) fmt ./...

check: fmt lint test build

dev:
	@$(GO) run $(ENTRYPOINT) --config config.yaml --debug

health:
	@bash healthcheck.sh

status:
	@echo "Go: $$($(GO) version 2>/dev/null || echo 'not installed')"
	@echo "Docker: $$($(DOCKER) --version 2>/dev/null || echo 'not installed')"
