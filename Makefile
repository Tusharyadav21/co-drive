# Co-Drive Makefile

.PHONY: help build run test clean dev docker-up docker-down migrate generate css

# Default target
help:
	@echo "Co-Drive - Go + HTMX Vehicle Management"
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  css          - Build Tailwind CSS"
	@echo "  run          - Run the application locally"
	@echo "  dev          - Run with hot reload (requires air)"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-up    - Start docker-compose services"
	@echo "  docker-down  - Stop docker-compose services"
	@echo "  migrate      - Run database migrations"
	@echo "  generate     - Generate code (if needed)"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golangci-lint (if installed)"

# Build Tailwind CSS
css:
	@echo "Building CSS..."
	@./bin/tailwindcss -i static/css/input.css -o static/css/styles.css

# Build the application
build:
	@echo "Building Co-Drive..."
	@CGO_ENABLED=0 go build -o bin/codrive ./cmd/server

# Run the application
run: build
	@echo "Starting Co-Drive..."
	@./bin/codrive

# Run with hot reload (requires: go install github.com/air-verse/air@latest)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# Format Go code
fmt:
	@go fmt ./...

# Run go vet
vet:
	@go vet ./...

# Run golangci-lint (if installed)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Docker compose up
docker-up:
	@docker-compose up -d

# Docker compose down
docker-down:
	@docker-compose down

# Docker compose logs
docker-logs:
	@docker-compose logs -f

# Run database migrations
migrate:
	@echo "Running migrations..."
	@go run ./cmd/server -migrate

# Generate code (placeholder for future use)
generate:
	@echo "Code generation not needed (no sqlc, no ORM)"

# Download dependencies
deps:
	@go mod download
	@go mod tidy

# Verify dependencies
verify:
	@go mod verify

# Build Docker image
docker-build:
	@docker build -t codrive:latest .

# Run all checks
check: fmt vet test
	@echo "All checks passed!"

# Install development tools
install-tools:
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"