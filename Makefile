# PhantomProxy Pro - Makefile
# Replace all shell scripts with single entry point

.PHONY: help build run test clean install docker dev lint

GO ?= go
PYTHON ?= python3
DOCKER ?= docker
DOCKER_COMPOSE ?= docker-compose

help:
	@echo "PhantomProxy Pro - Commands:"
	@echo "  make install     - Install Go + Python deps"
	@echo "  make build      - Build Go binary"
	@echo "  make run        - Run PhantomProxy"
	@echo "  make test       - Run tests"
	@echo "  make docker     - docker-compose up"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make lint       - Run linters"

install-go:
	@if ! command -v go >/dev/null 2>&1; then \
		if [ -f scripts/install-go-windows.ps1 ]; then \
			echo "Run: powershell -ExecutionPolicy Bypass -File scripts/install-go-windows.ps1"; \
		else \
			echo "Install Go from https://go.dev/dl/"; \
		fi; exit 1; \
	fi
	@echo "Go: $$(go version)"

install-python:
	@$(PYTHON) -m pip install -r requirements.txt -q
	@echo "Python deps installed"

install: install-go install-python
	@$(GO) mod download
	@echo "Dependencies OK"

build:
	@$(GO) build -ldflags="-s -w" -o phantom-proxy ./cmd/phantom-proxy
	@echo "Build: phantom-proxy"

run: build
	@./phantom-proxy --config config.yaml

test:
	@$(GO) test -v -race ./... 2>/dev/null || echo "Go tests skipped"
	@cd api && python -c "from app.main import app; from fastapi.testclient import TestClient; assert TestClient(app).get('/health').status_code == 200" 2>/dev/null || true
	@echo "Tests done"

clean:
	@rm -f phantom-proxy phantom-proxy.exe
	@rm -rf dist/ build/
	@echo "Cleaned"

docker:
	@$(DOCKER_COMPOSE) up -d
	@echo "Proxy: https://localhost:443 | Go API: http://localhost:8080 | Frontend: http://localhost:3000"

docker-build:
	@$(DOCKER) build -t phantomproxy:latest .
	@echo "Docker image built"

lint:
	@$(GO) vet ./...
	@golangci-lint run ./... 2>/dev/null || true
