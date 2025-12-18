.PHONY: build clean test help

# Binary name
BINARY_NAME=go_notes

# Build directory
BUILD_DIR=./bin

# Get git commit hash
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

# Get build date in RFC3339 format
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS=-ldflags "-X main.GitCommit=$(GIT_COMMIT) -X main.BuildDate=$(BUILD_DATE)"

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build       - Build the application with build metadata"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  install     - Build and install to GOPATH/bin"
	@echo "  version     - Show version information"

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## install: Build and install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "Installation complete"

## version: Show version information
version:
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
