# SSH Tunnel Manager Makefile

# Build variables
BINARY_NAME=ssh-tunnel
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"

# Go build flags
GO_BUILD_FLAGS=-trimpath
GO_TEST_FLAGS=-race -coverprofile=coverage.out

# Directories
BUILD_DIR=build
DIST_DIR=dist
CMD_DIR=cmd/cli

# Platforms for cross-compilation
PLATFORMS=linux/amd64 linux/arm64 linux/arm darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build clean test lint fmt vet tidy run install build-all release

# Default target
all: clean fmt vet test build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Build for all platforms
build-all: clean
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$$output_name ./$(CMD_DIR); \
		if [ $$? -ne 0 ]; then \
			echo "Failed to build for $$GOOS/$$GOARCH"; \
			exit 1; \
		fi; \
	done
	@echo "All builds completed successfully!"

# Create release packages
release: build-all
	@echo "Creating release packages..."
	@cd $(DIST_DIR) && for binary in *; do \
		if [ -f "$$binary" ]; then \
			echo "Packaging $$binary..."; \
			tar -czf "$$binary.tar.gz" "$$binary"; \
		fi; \
	done
	@echo "Release packages created in $(DIST_DIR)/"

# Run tests
test:
	@echo "Running tests..."
	go test $(GO_TEST_FLAGS) ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v $(GO_TEST_FLAGS) ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Run golangci-lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./$(CMD_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download

# Generate code (if needed)
generate:
	@echo "Generating code..."
	go generate ./...

# Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "Development environment setup complete!"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t ssh-tunnel-manager:$(VERSION) .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker run --rm -it ssh-tunnel-manager:$(VERSION)

# CI pipeline
ci: fmt vet lint test security build

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, format, vet, test, and build"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all supported platforms"
	@echo "  release      - Create release packages"
	@echo "  test         - Run tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  bench        - Run benchmarks"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run linter"
	@echo "  tidy         - Tidy dependencies"
	@echo "  run          - Build and run"
	@echo "  install      - Install to GOPATH/bin"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  generate     - Generate code"
	@echo "  security     - Run security scan"
	@echo "  dev-setup    - Setup development environment"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  ci           - Run CI pipeline"
	@echo "  help         - Show this help"
