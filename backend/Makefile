# Makefile for Omni Go Backend
# Usage: make <target>

.PHONY: help build test test-unit test-integration test-security test-e2e test-all clean run docker-build docker-test mocks

# Default target
help:
	@echo "Omni Go Backend Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build             - Build the backend binary"
	@echo "  run               - Run the backend locally"
	@echo "  test              - Run unit tests only"
	@echo "  test-unit         - Run unit tests"
	@echo "  test-integration  - Run integration tests (requires running backend)"
	@echo "  test-security     - Run security tests (requires running backend)"
	@echo "  test-e2e          - Run end-to-end/lighthouse tests (requires running backend)"
	@echo "  test-all          - Run all tests"
	@echo "  mocks             - Generate mocks for all interfaces"
	@echo "  docker-build      - Build Docker image"
	@echo "  docker-test       - Run integration tests inside Docker"
	@echo "  clean             - Clean build artifacts"
	@echo ""

# Build
build:
	@echo "Building backend..."
	go build -o bin/omni-backend ./cmd/server

# Generate mocks for all interfaces
mocks:
	@echo "Generating mocks..."
	@mkdir -p internal/mocks
	mockery --all --dir ./internal --output internal/mocks --outpkg mocks

# Run locally
run:
	@echo "Running backend..."
	go run ./cmd/server

# Unit tests only (no external dependencies)
test: test-unit

test-unit:
	@echo "Running unit tests..."
	go test -v ./tests/unit/...

# Integration tests (requires running backend at localhost:3000)
test-integration:
	@echo "Running integration tests..."
	@echo "Note: Backend must be running at localhost:3000"
	go test -v ./tests/integration/...

# Security tests (requires running backend at localhost:3000)
test-security:
	@echo "Running security tests..."
	@echo "Note: Backend must be running at localhost:3000"
	go test -v ./tests/security/...

# E2E/Lighthouse tests (requires running backend at localhost:3000)
test-e2e:
	@echo "Running e2e/lighthouse tests..."
	@echo "Note: Backend must be running at localhost:3000"
	go test -v ./tests/e2e/...

# All tests
test-all:
	@echo "Running all tests..."
	go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t omni-backend:latest .

# Run integration tests inside Docker (backend must be running)
docker-test:
	@echo "Running integration tests inside Docker..."
	docker exec omni-backend sh -c '\
		BASE_URL="http://localhost:3000" \
		USERNAME="tester" \
		PASSWORD="tester@123" \
		&& echo "=== Quick Integration Test ===" \
		&& echo "1. Health check..." \
		&& wget -qO- $$BASE_URL/api/health | head -c 100 \
		&& echo "" \
		&& echo "2. Login test..." \
		&& LOGIN=$$(wget -qO- --header="Content-Type: application/json" --post-data="{\"username\":\"$$USERNAME\",\"password\":\"$$PASSWORD\"}" $$BASE_URL/api/auth/login) \
		&& echo "   Success: $$(echo $$LOGIN | grep -o "\"success\":true" | head -1)" \
		&& TOKEN=$$(echo "$$LOGIN" | sed -n "s/.*\"token\":\"\([^\"]*\)\".*/\1/p") \
		&& echo "3. Get tenants..." \
		&& wget -qO- --header="Authorization: Bearer $$TOKEN" $$BASE_URL/api/auth/tenants | head -c 150 \
		&& echo "" \
		&& echo "4. Switch tenant..." \
		&& SWITCH=$$(wget -qO- --header="Authorization: Bearer $$TOKEN" --header="Content-Type: application/json" --post-data="{\"tenant_id\":\"yumna_bertigamart\"}" $$BASE_URL/api/auth/switch-tenant) \
		&& echo "   Success: $$(echo $$SWITCH | grep -o "\"success\":true" | head -1)" \
		&& echo "=== Test Complete ===" \
	'

# Clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

# Lint (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Generate swagger docs (if using swag)
swagger:
	@echo "Generating swagger docs..."
	swag init -g cmd/server/main.go -o docs/swagger
