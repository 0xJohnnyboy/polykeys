.PHONY: all build build-all test clean install help

# Variables
BINARY_CLI=polykeys
BINARY_DAEMON=polykeysd
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform..."
	go build $(LDFLAGS) -o $(BINARY_CLI) ./cmd/polykeys
	go build $(LDFLAGS) -o $(BINARY_DAEMON) ./cmd/polykeysd
	@echo "Build complete: $(BINARY_CLI), $(BINARY_DAEMON)"

# Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "All builds complete"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p dist/linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/linux/$(BINARY_CLI) ./cmd/polykeys
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/linux/$(BINARY_DAEMON) ./cmd/polykeysd
	@echo "Linux build complete: dist/linux/"

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p dist/windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/windows/$(BINARY_CLI).exe ./cmd/polykeys
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/windows/$(BINARY_DAEMON).exe ./cmd/polykeysd
	@echo "Windows build complete: dist/windows/"

# Build for macOS
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p dist/darwin
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/darwin/$(BINARY_CLI) ./cmd/polykeys
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/darwin/$(BINARY_DAEMON) ./cmd/polykeysd
	@echo "macOS build complete: dist/darwin/"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install to system
install:
	@echo "Installing..."
	go install $(LDFLAGS) ./cmd/polykeys
	go install $(LDFLAGS) ./cmd/polykeysd
	@echo "Install complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_CLI) $(BINARY_DAEMON)
	rm -rf dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Update dependencies
deps:
	@echo "Updating dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies updated"

# Show help
help:
	@echo "Polykeys Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build         - Build for current platform (default)"
	@echo "  build-all     - Build for all platforms (Linux, Windows, macOS)"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-darwin  - Build for macOS"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  install       - Install binaries to system"
	@echo "  clean         - Remove build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  deps          - Update dependencies"
	@echo "  help          - Show this help"
