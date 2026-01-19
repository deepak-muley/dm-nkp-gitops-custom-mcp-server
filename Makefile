# dm-nkp-gitops-custom-mcp-server Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=dm-nkp-gitops-mcp-server
BINARY_DIR=bin

# Version info
VERSION?=0.1.0
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/server

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/server

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Verify dependencies
.PHONY: verify
verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Run the server locally
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME) serve

# Run in read-only mode
.PHONY: run-readonly
run-readonly: build
	@echo "Running $(BINARY_NAME) in read-only mode..."
	./$(BINARY_DIR)/$(BINARY_NAME) serve --read-only

# Run with debug logging
.PHONY: run-debug
run-debug: build
	@echo "Running $(BINARY_NAME) with debug logging..."
	./$(BINARY_DIR)/$(BINARY_NAME) serve --log-level=debug

# Install to GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Install to /usr/local/bin (requires sudo)
.PHONY: install-global
install-global: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo cp $(BINARY_DIR)/$(BINARY_NAME) /usr/local/bin/

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# Generate mocks (if needed)
.PHONY: generate
generate:
	@echo "Running go generate..."
	$(GOCMD) generate ./...

# Test MCP protocol (sends tools/list request)
.PHONY: test-mcp
test-mcp: build
	@echo "Testing MCP protocol..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./$(BINARY_DIR)/$(BINARY_NAME) serve

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy dependencies"
	@echo "  verify         - Verify dependencies"
	@echo "  run            - Build and run the server"
	@echo "  run-readonly   - Run in read-only mode"
	@echo "  run-debug      - Run with debug logging"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  install-global - Install to /usr/local/bin"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  generate       - Run go generate"
	@echo "  test-mcp       - Test MCP protocol"
	@echo "  help           - Show this help"
