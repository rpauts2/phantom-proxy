# Makefile for PhantomProxy v14.0
# Main entry point for development tasks

.PHONY: help build run test clean install docker dev lint fmt check

GO ?= go
PYTHON ?= python3
DOCKER ?= docker
DOCKER_COMPOSE ?= docker-compose

help:
	@echo "PhantomProxy v14.0 - Commands:"
	@echo ""
	@echo "  make install     - Install Go + Python dependencies"
	@echo "  make build       - Build Go binary"
	@echo "  make run         - Run PhantomProxy"
	@echo "  make test        - Run all tests"
	@echo "  make test-go     - Run Go tests only"
	@echo "  make test-python - Run Python tests only"
	@echo "  make docker      - Start Docker Compose"
	@echo "  make docker-stop - Stop Docker Compose"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make lint        - Run linters"
	@echo "  make fmt         - Format code"
	@echo "  make check       - Full validation (lint + test + build)"
	@echo "  make dev         - Development mode"
	@echo "  make backup      - Create backup"
	@echo "  make health      - Health check"
	@echo ""

install-go:
	@if ! command -v go >/dev/null 2>&1; then \
		echo "Please install Go from https://go.dev/dl/"; exit 1; \
	fi
	@echo "Go: $$(go version)"

install-python:
	@$(PYTHON) -m pip install -r requirements.txt -q
	@echo "Python dependencies installed"

install: install-go install-python
	@$(GO) mod download
	@echo "All dependencies installed"

build:
	@echo "Building PhantomProxy..."
	@$(GO) build -ldflags="-s -w" -o phantom-proxy ./cmd/phantom-proxy-v14
	@echo "Build successful: phantom-proxy"

run: build
	@echo "Starting PhantomProxy..."
	@./phantom-proxy --config config.yaml

test: test-go test-python
	@echo ""
	@echo "======================================"
	@echo "All tests completed!"
	@echo "======================================"

test-go:
	@echo ""
	@echo "Running Go tests..."
	@$(GO) test -v -race ./internal/... ./cmd/... -timeout 5m

test-python:
	@echo ""
	@echo "Running Python tests..."
	@$(PYTHON) -m pytest tests/ -v

clean:
	@echo "Cleaning..."
	@rm -f phantom-proxy phantom-proxy.exe
	@rm -rf dist/ build/ __pycache__/ .pytest_cache/
	@rm -rf ai_service/__pycache__/ api/__pycache__/
	@find . -type f -name "*.pyc" -delete 2>/dev/null || true
	@echo "Cleaned"

docker:
	@$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "Services started:"
	@echo "  - Proxy: https://localhost:8443"
	@echo "  - Go API: http://localhost:8080"
	@echo "  - Python API: http://localhost:8000"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Grafana: http://localhost:3001 (admin/admin)"
	@echo "  - Prometheus: http://localhost:9090"

docker-stop:
	@$(DOCKER_COMPOSE) down
	@echo "Services stopped"

docker-build:
	@$(DOCKER) build -t phantomproxy:latest .
	@echo "Docker image built"

docker-logs:
	@$(DOCKER_COMPOSE) logs -f

lint:
	@echo "Running Go vet..."
	@$(GO) vet ./...
	@echo "Running golangci-lint..."
	@golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed"
	@echo "Lint completed"

fmt:
	@echo "Formatting Go code..."
	@$(GO) fmt ./...
	@echo "Code formatted"

check: lint test build
	@echo ""
	@echo "======================================"
	@echo "Full validation completed!"
	@echo "======================================"

dev:
	@echo "Starting development mode..."
	@$(GO) run ./cmd/phantom-proxy-v14/main.go --config config.yaml --debug

backup:
	@echo "Creating backup..."
	@$(PYTHON) backup.py backup

health:
	@echo "Running health check..."
	@bash healthcheck.sh

status:
	@echo "PhantomProxy v14.0 Status:"
	@echo "  Go version: $$(go version 2>/dev/null || echo 'not installed')"
	@echo "  Python version: $$(python3 --version 2>/dev/null || echo 'not installed')"
	@echo "  Docker version: $$(docker --version 2>/dev/null || echo 'not installed')"
