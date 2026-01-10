# Vendatta - Development Environment Manager
# Makefile for development, testing, and CI workflows

.PHONY: help build install clean test test-unit test-integration test-e2e test-all lint fmt fmt-check docker-build docker-push release

# Default target
help: ## Show this help message
	@echo "Vendatta Development Makefile"
	@echo ""
	@echo "Development targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Testing targets:"
	@echo "  test-unit          Run unit tests with coverage"
	@echo "  test-integration   Run integration tests"
	@echo "  test-e2e           Run end-to-end tests"
	@echo "  test-all           Run all tests (unit + integration + e2e)"
	@echo ""
	@echo "CI targets:"
	@echo "  ci-check           Run all checks (lint, fmt, test)"
	@echo "  ci-build           Build for multiple platforms"
	@echo "  ci-docker          Build and push Docker image"

# Development
build: ## Build vendatta binary
	go build -o bin/vendatta cmd/oursky/main.go

install: build ## Install vendatta to ~/.local/bin
	cp bin/vendatta ~/.local/bin/vendatta
	chmod +x ~/.local/bin/vendatta

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format Go code
	go fmt ./...
	go mod tidy

fmt-check: ## Check if code is properly formatted
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		gofmt -l .; \
		exit 1; \
	fi

# Testing
test-unit: ## Run unit tests with coverage
	go test -v -race -coverprofile=coverage.out ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests
	go test -v -race -tags=integration ./pkg/...

test-e2e: ## Run end-to-end tests
	./scripts/run-e2e-tests.sh

test-all: test-unit test-integration test-e2e ## Run all tests

# Docker
docker-build: ## Build Docker image
	docker build -t vendatta:latest .

docker-push: ## Push Docker image
	docker tag vendatta:latest inizio/vendatta:latest
	docker push inizio/vendatta:latest

# CI Pipeline
ci-check: fmt-check lint test-all ## Run all CI checks

ci-build: ## Build for multiple platforms
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/vendatta-linux-amd64 cmd/oursky/main.go
	GOOS=linux GOARCH=arm64 go build -o dist/vendatta-linux-arm64 cmd/oursky/main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/vendatta-darwin-amd64 cmd/oursky/main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/vendatta-darwin-arm64 cmd/oursky/main.go
	GOOS=windows GOARCH=amd64 go build -o dist/vendatta-windows-amd64.exe cmd/oursky/main.go

ci-docker: docker-build docker-push ## Build and push Docker image

# Release
release: ci-check ci-build ## Create release artifacts
	@echo "Release artifacts created in dist/"

# Development helpers
dev-setup: ## Set up development environment
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev-test-watch: ## Run tests in watch mode (requires entr)
	find . -name "*.go" | entr -r make test-unit

# Performance testing
perf-test: ## Run performance tests
	go test -bench=. -benchmem ./pkg/...
	@echo "Performance test complete. Check memory usage and startup times."

# Security
security-scan: ## Run security vulnerability scan
	gosec ./...
	trivy filesystem --exit-code 1 --no-progress .

# Documentation
docs-build: ## Build documentation
	@echo "Building docs..."
	# Add documentation build commands here

docs-serve: ## Serve documentation locally
	@echo "Serving docs on http://localhost:8000"
	# Add docs serve commands here