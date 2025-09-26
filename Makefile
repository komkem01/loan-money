# Makefile for Loan Money API

.PHONY: build run test clean install-deps dev setup-db

# Variables
APP_NAME=loan-money
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go

# Default target
all: clean install-deps build

# Install dependencies
install-deps:
	go mod download
	go mod tidy

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

# Run the application in development mode
dev:
	@echo "Starting development server..."
	go run $(MAIN_PATH)

# Run the built application
run: build
	@echo "Starting $(APP_NAME) server..."
	./$(BUILD_DIR)/$(APP_NAME)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Setup development environment
setup-dev:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo ".env file created"; fi
	go mod download

# Database setup (requires PostgreSQL running)
setup-db:
	@echo "Setting up database..."
	psql -U postgres -c "CREATE DATABASE loan_money;" || true

# Run API tests (requires server to be running)
test-api:
	@echo "Running API tests..."
	@if command -v bash >/dev/null 2>&1; then \
		chmod +x test_api.sh && ./test_api.sh; \
	else \
		powershell -ExecutionPolicy Bypass -File test_api.ps1; \
	fi

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, install deps, and build"
	@echo "  install-deps - Download and tidy Go modules"
	@echo "  build        - Build the application"
	@echo "  dev          - Run in development mode"
	@echo "  run          - Build and run the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Remove build artifacts"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Run linter (requires golangci-lint)"
	@echo "  setup-dev    - Setup development environment"
	@echo "  setup-db     - Create database"
	@echo "  test-api     - Run API integration tests"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  help         - Show this help message"