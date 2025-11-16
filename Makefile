.PHONY: build test test-coverage run run-custom clean deps fmt lint help

# Build the application
build:
	@echo "Building OTLP Log Parser Assignment..."
	go build -o otlp-log-parser-assignment ./cmd

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the application with default settings
run: build
	@echo "Running OTLP Log Parser Assignment..."
	./otlp-log-parser-assignment

# Run with custom settings
run-custom: build
	@echo "Running OTLP Log Parser Assignment with custom settings..."
	./otlp-log-parser-assignment -attribute-key=foo -window-duration=5s -debug=true

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f otlp-log-parser-assignment coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Build and run with default settings"
	@echo "  run-custom    - Build and run with custom settings"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  help          - Show this help message"
