# Makefile for Desktop Surveillance Camera

# Variables
BINARY_NAME=surveillance-camera
WINDOWS_BINARY=$(BINARY_NAME).exe
LINUX_BINARY=$(BINARY_NAME)
SOURCE_DIR=.
BUILD_DIR=build
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: windows

# Windows build (native or cross-compile)
.PHONY: windows
windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(WINDOWS_BINARY) $(SOURCE_DIR)
	@echo "Windows binary created: $(BUILD_DIR)/$(WINDOWS_BINARY)"

# Windows build with 32-bit support
.PHONY: windows-386
windows-386:
	@echo "Building for Windows 32-bit..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=386 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-386.exe $(SOURCE_DIR)
	@echo "Windows 32-bit binary created: $(BUILD_DIR)/$(BINARY_NAME)-386.exe"

# Linux build (for development/testing)
.PHONY: linux
linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(LINUX_BINARY) $(SOURCE_DIR)
	@echo "Linux binary created: $(BUILD_DIR)/$(LINUX_BINARY)"

# Native build (for current platform)
.PHONY: native
native:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_DIR)
	@echo "Native binary created: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(WINDOWS_BINARY) $(LINUX_BINARY) $(BINARY_NAME)
	@echo "Clean completed"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run
	@echo "Linting completed"

# Create release package
.PHONY: release
release: clean windows
	@echo "Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(WINDOWS_BINARY) $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@echo '{"server":{"host":"0.0.0.0","port":9981},"capture":{"mode":"ondemand","interval":"5s"}}' > $(BUILD_DIR)/release/config.json.example
	@cd $(BUILD_DIR)/release && zip -r ../surveillance-camera-v$(VERSION)-windows.zip .
	@echo "Release package created: $(BUILD_DIR)/surveillance-camera-v$(VERSION)-windows.zip"

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Development environment ready"

# Quick test build and run (Windows)
.PHONY: run-windows
run-windows: windows
	@echo "Starting Windows binary..."
	@echo "Note: Run the following command on Windows:"
	@echo "$(BUILD_DIR)/$(WINDOWS_BINARY) -help"

# Quick test screenshot function
.PHONY: test-screenshot
test-screenshot: windows
	@echo "Testing screenshot function..."
	@echo "Note: Run the following command on Windows:"
	@echo "$(BUILD_DIR)/$(WINDOWS_BINARY) -test"

# Show build info
.PHONY: info
info:
	@echo "Build Information:"
	@echo "  Binary Name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Build Dir: $(BUILD_DIR)"
	@echo "  Go Version: $(shell go version)"
	@echo "  CGO Enabled: $(CGO_ENABLED)"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Build Windows binary (default)"
	@echo "  windows      - Build Windows 64-bit binary"
	@echo "  windows-386  - Build Windows 32-bit binary"
	@echo "  linux        - Build Linux binary"
	@echo "  native       - Build for current platform"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  release      - Create release package"
	@echo "  dev-setup    - Setup development environment"
	@echo "  run-windows  - Build and show run command for Windows"
	@echo "  test-screenshot - Build and show test command"
	@echo "  info         - Show build information"
	@echo "  help         - Show this help message"