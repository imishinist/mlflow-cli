.PHONY: build test clean install deps cross-compile dev docker-build docker-up docker-down docker-logs e2e-test e2e-test-debug e2e-test-all e2e-test-full e2e-test-all-debug e2e-test-full-debug help

# Build the binary
build:
	go build -o mlflow-cli main.go

# Run unit tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f mlflow-cli mlflow-cli-dev
	rm -rf dist/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install the binary locally
install: build
	cp mlflow-cli $(GOPATH)/bin/

# Cross-compile for multiple platforms
cross-compile:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/mlflow-cli-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o dist/mlflow-cli-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o dist/mlflow-cli-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o dist/mlflow-cli-windows-amd64.exe main.go

# Development build with race detection
dev:
	go build -race -o mlflow-cli-dev main.go

# Docker operations
docker-build:
	docker-compose build mlflow

docker-up: docker-build
	docker-compose up -d
	@echo "MLflow server starting at http://localhost:5001"
	@echo "Use 'make docker-logs' to view logs"

docker-down:
	docker-compose down -v
	rm -rf mlflow-data

docker-logs:
	docker-compose logs -f mlflow

# E2E Tests (default: fast mode using existing server)
e2e-test: build
	@echo "Running E2E tests (fast mode - using existing MLflow server)..."
	@echo "Note: Make sure MLflow server is running with 'make docker-up'"
	./scripts/e2e-test.sh --skip-docker

# E2E Tests with debug output (fast mode)
e2e-test-debug: build
	@echo "Running E2E tests with debug output (fast mode)..."
	@echo "Note: Make sure MLflow server is running with 'make docker-up'"
	./scripts/e2e-test.sh --debug --skip-docker

# E2E Tests (full mode with Docker setup)
e2e-test-all: build docker-build
	@echo "Running E2E tests (full mode - with Docker setup)..."
	./scripts/e2e-test.sh

e2e-test-full: e2e-test-all

# E2E Tests with debug output (full mode)
e2e-test-all-debug: build docker-build
	@echo "Running E2E tests with debug output (full mode)..."
	./scripts/e2e-test.sh --debug

e2e-test-full-debug: e2e-test-all-debug

# Help target
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  test               - Run unit tests"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install dependencies"
	@echo "  install            - Install binary locally"
	@echo "  cross-compile      - Cross-compile for multiple platforms"
	@echo "  dev                - Development build with race detection"
	@echo ""
	@echo "Docker operations:"
	@echo "  docker-build       - Build MLflow Docker image"
	@echo "  docker-up          - Start MLflow server"
	@echo "  docker-down        - Stop MLflow server"
	@echo "  docker-logs        - View MLflow server logs"
	@echo ""
	@echo "E2E Tests:"
	@echo "  e2e-test           - Run E2E tests (fast mode, default)"
	@echo "  e2e-test-debug     - Run E2E tests with debug (fast mode)"
	@echo "  e2e-test-all       - Run E2E tests (full mode with Docker)"
	@echo "  e2e-test-full      - Alias for e2e-test-all"
	@echo "  e2e-test-all-debug - Run E2E tests with debug (full mode)"
	@echo ""
	@echo "Development workflow:"
	@echo "  1. make docker-up     # Start MLflow server once"
	@echo "  2. make e2e-test      # Run tests multiple times (fast)"
	@echo "  3. make docker-down   # Stop server when done"
