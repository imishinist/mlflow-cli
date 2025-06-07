.PHONY: build test clean install deps cross-compile

# Build the binary
build:
	go build -o bin/mlflow-cli main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install the binary
install:
	go install

# Cross-compile for multiple platforms
cross-compile:
	GOOS=linux GOARCH=amd64 go build -o bin/mlflow-cli-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/mlflow-cli-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/mlflow-cli-darwin-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/mlflow-cli-windows-amd64.exe main.go

# Development build with race detection
dev:
	go build -race -o bin/mlflow-cli-dev main.go

# Run with example
example:
	./bin/mlflow-cli run start --experiment-name "test-experiment" --run-name "test-run"
