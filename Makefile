# dm-nkp-gitops-custom-mcp-server Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=dm-nkp-gitops-mcp-server
A2A_BINARY_NAME=dm-nkp-gitops-a2a-server
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

# Build the A2A server binary
.PHONY: build-a2a
build-a2a:
	@echo "Building $(A2A_BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(A2A_BINARY_NAME) ./cmd/a2a-server

# Build both MCP and A2A servers
.PHONY: build-both
build-both: build build-a2a

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BINARY_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(A2A_BINARY_NAME)-darwin-amd64 ./cmd/a2a-server
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(A2A_BINARY_NAME)-darwin-arm64 ./cmd/a2a-server
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(A2A_BINARY_NAME)-linux-amd64 ./cmd/a2a-server
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(A2A_BINARY_NAME)-linux-arm64 ./cmd/a2a-server

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

# =============================================================================
# A2A Server Targets
# =============================================================================

# Run the A2A server
.PHONY: run-a2a
run-a2a: build-a2a
	@echo "Running $(A2A_BINARY_NAME)..."
	./$(BINARY_DIR)/$(A2A_BINARY_NAME) serve

# Run A2A server with debug logging
.PHONY: run-a2a-debug
run-a2a-debug: build-a2a
	@echo "Running $(A2A_BINARY_NAME) with debug logging..."
	./$(BINARY_DIR)/$(A2A_BINARY_NAME) serve --log-level=debug

# Run A2A server on custom port
.PHONY: run-a2a-8081
run-a2a-8081: build-a2a
	@echo "Running $(A2A_BINARY_NAME) on port 8081..."
	./$(BINARY_DIR)/$(A2A_BINARY_NAME) serve --port 8081

# Test A2A agent card
.PHONY: test-a2a-card
test-a2a-card:
	@echo "Testing A2A agent card..."
	@curl -s http://localhost:8080/.well-known/agent.json | jq '.name, .version, (.skills | length) as $$n | "\($$n) skills available"'

# Test A2A health
.PHONY: test-a2a-health
test-a2a-health:
	@echo "Testing A2A health endpoint..."
	@curl -s http://localhost:8080/health | jq

# Test A2A task creation
.PHONY: test-a2a-task
test-a2a-task:
	@echo "Creating A2A task..."
	@curl -s -X POST http://localhost:8080/ \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tasks/create","params":{"skill":"list-contexts","input":{}}}' | jq

# Run multi-agent demo
.PHONY: demo-multi-agent
demo-multi-agent: build-a2a
	@echo "Running multi-agent orchestrator demo..."
	@echo "(Make sure A2A server is running: make run-a2a)"
	go run examples/multi-agent/orchestrator/main.go

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

# Kind cluster testing
.PHONY: kind-setup
kind-setup: build
	@echo "Setting up kind cluster with Flux CD..."
	./scripts/test-with-kind.sh setup

.PHONY: kind-test
kind-test: build
	@echo "Running MCP tests against kind cluster..."
	./scripts/test-with-kind.sh test

.PHONY: kind-all
kind-all: build
	@echo "Full kind cluster test (setup + test)..."
	./scripts/test-with-kind.sh all

.PHONY: kind-cleanup
kind-cleanup:
	@echo "Cleaning up kind cluster..."
	./scripts/test-with-kind.sh cleanup

.PHONY: kind-interactive
kind-interactive: build
	@echo "Starting interactive MCP testing..."
	./scripts/test-with-kind.sh interactive

# Test all methods
.PHONY: test-all-methods
test-all-methods: build
	@echo "Testing all MCP methods..."
	@echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test"}}}' | ./$(BINARY_DIR)/$(BINARY_NAME) serve --read-only 2>/dev/null | jq -r '.result.serverInfo'
	@echo ""
	@echo "Tools available:"
	@echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./$(BINARY_DIR)/$(BINARY_NAME) serve --read-only 2>/dev/null | jq -r '.result.tools[].name'

# =============================================================================
# Docker Targets
# =============================================================================

DOCKER_REGISTRY?=ghcr.io/deepak-muley
DOCKER_IMAGE_NAME?=dm-nkp-gitops-a2a-server
DOCKER_TAG?=$(VERSION)

.PHONY: docker-build-a2a
docker-build-a2a:
	@echo "Building Docker image..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG) \
		-f Dockerfile .

.PHONY: docker-push-a2a
docker-push-a2a: docker-build-a2a
	@echo "Pushing Docker image..."
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# =============================================================================
# Helm Targets
# =============================================================================

HELM_CHART_PATH=chart/dm-nkp-gitops-a2a-server

.PHONY: helm-lint-a2a
helm-lint-a2a:
	@echo "Linting Helm chart..."
	helm lint $(HELM_CHART_PATH)

.PHONY: helm-package-a2a
helm-package-a2a:
	@echo "Packaging Helm chart..."
	helm package $(HELM_CHART_PATH) --destination .helm-packages

.PHONY: helm-push-a2a
helm-push-a2a: helm-package-a2a
	@echo "Pushing Helm chart to OCI registry..."
	helm push .helm-packages/dm-nkp-gitops-a2a-server-$(VERSION).tgz oci://$(DOCKER_REGISTRY)/charts

# =============================================================================
# Enterprise Deployment Targets
# =============================================================================

# Install Kagent operator (required for MCPServer CRD mode)
.PHONY: install-kagent
install-kagent:
	@echo "Installing Kagent operator..."
	kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml
	@echo "Waiting for Kagent CRDs..."
	kubectl wait --for=condition=established crd/mcpservers.kagent.dev --timeout=60s
	@echo "Kagent installed successfully"

# Install with traditional Deployment (works everywhere)
.PHONY: helm-install-std
helm-install-std:
	@echo "Installing with traditional Deployment mode..."
	helm upgrade --install dm-nkp-gitops-a2a-server $(HELM_CHART_PATH) \
		--namespace gitops-agent \
		--create-namespace \
		--set deploymentMode=deployment \
		--set image.tag=$(VERSION)

# Install with Kagent MCPServer CRD (K8s-native, requires Kagent)
.PHONY: helm-install-mcpserver
helm-install-mcpserver:
	@echo "Installing with MCPServer CRD mode..."
	@echo "Note: Requires Kagent operator (run 'make install-kagent' first)"
	helm upgrade --install dm-nkp-gitops-a2a-server $(HELM_CHART_PATH) \
		--namespace gitops-agent \
		--create-namespace \
		--set deploymentMode=mcpserver \
		--set image.tag=$(VERSION)

# Install with enterprise production values
.PHONY: helm-install-enterprise
helm-install-enterprise:
	@echo "Installing with enterprise production values..."
	@echo "Note: Requires Kagent operator (run 'make install-kagent' first)"
	helm upgrade --install dm-nkp-gitops-a2a-server $(HELM_CHART_PATH) \
		--namespace gitops-agent \
		--create-namespace \
		-f $(HELM_CHART_PATH)/values-enterprise.yaml \
		--set image.tag=$(VERSION) \
		--set httpRoute.hostname=gitops-agent.$(shell kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}').nip.io

# Show what would be deployed (dry-run)
.PHONY: helm-template-mcpserver
helm-template-mcpserver:
	@echo "Rendering MCPServer mode templates..."
	helm template dm-nkp-gitops-a2a-server $(HELM_CHART_PATH) \
		--namespace gitops-agent \
		--set deploymentMode=mcpserver \
		--set image.tag=$(VERSION)

# Uninstall
.PHONY: helm-uninstall
helm-uninstall:
	@echo "Uninstalling dm-nkp-gitops-a2a-server..."
	helm uninstall dm-nkp-gitops-a2a-server --namespace gitops-agent || true

# =============================================================================
# E2E Testing
# =============================================================================

.PHONY: e2e-a2a-all
e2e-a2a-all:
	@echo "Running full A2A E2E test..."
	./scripts/e2e-a2a-test.sh all

.PHONY: e2e-a2a-setup
e2e-a2a-setup:
	@echo "Setting up A2A E2E environment..."
	./scripts/e2e-a2a-test.sh setup

.PHONY: e2e-a2a-test
e2e-a2a-test:
	@echo "Running A2A E2E tests..."
	./scripts/e2e-a2a-test.sh test

.PHONY: e2e-a2a-cleanup
e2e-a2a-cleanup:
	@echo "Cleaning up A2A E2E environment..."
	./scripts/e2e-a2a-test.sh cleanup

# =============================================================================
# Troubleshooting Examples
# =============================================================================

# Build troubleshooting example
.PHONY: build-troubleshooter
build-troubleshooter:
	@echo "Building troubleshooting example..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/troubleshooter ./examples/troubleshooting

# Run troubleshooting example
.PHONY: run-troubleshooter
run-troubleshooter: build-troubleshooter
	@echo "Running troubleshooting example..."
	@echo ""
	@echo "Available workflows: gitops-failure, cluster-node, app-deployment, all"
	@echo "Usage: ./$(BINARY_DIR)/troubleshooter <workflow_name>"
	@echo ""
	@./$(BINARY_DIR)/troubleshooter $(WORKFLOW) || echo "Usage: make run-troubleshooter WORKFLOW=gitops-failure"

# Test all troubleshooting workflows
.PHONY: test-troubleshooter
test-troubleshooter: build-troubleshooter
	@echo "Testing all troubleshooting workflows..."
	@echo ""
	@echo "=== GitOps Failure Workflow ==="
	@./$(BINARY_DIR)/troubleshooter gitops-failure
	@echo ""
	@echo "=== Cluster Node Workflow ==="
	@./$(BINARY_DIR)/troubleshooter cluster-node
	@echo ""
	@echo "=== App Deployment Workflow ==="
	@./$(BINARY_DIR)/troubleshooter app-deployment
	@echo ""
	@echo "=== All Workflows as JSON ==="
	@./$(BINARY_DIR)/troubleshooter all | jq '.' || ./$(BINARY_DIR)/troubleshooter all

# Show troubleshooting workflows as JSON
.PHONY: troubleshoot-workflows-json
troubleshoot-workflows-json: build-troubleshooter
	@echo "Troubleshooting workflows (JSON format):"
	@./$(BINARY_DIR)/troubleshooter all | jq '.' || ./$(BINARY_DIR)/troubleshooter all

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build          - Build the MCP server binary"
	@echo "  build-a2a      - Build the A2A server binary"
	@echo "  build-both     - Build both MCP and A2A servers"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Run MCP Server:"
	@echo "  run            - Build and run the MCP server"
	@echo "  run-readonly   - Run in read-only mode"
	@echo "  run-debug      - Run with debug logging"
	@echo ""
	@echo "Run A2A Server:"
	@echo "  run-a2a        - Build and run the A2A server (port 8080)"
	@echo "  run-a2a-debug  - Run A2A server with debug logging"
	@echo "  run-a2a-8081   - Run A2A server on port 8081 (for multi-agent)"
	@echo ""
	@echo "Test MCP:"
	@echo "  test           - Run Go unit tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  test-mcp       - Test MCP protocol basics"
	@echo "  test-all-methods - Test all MCP methods"
	@echo ""
	@echo "Test A2A:"
	@echo "  test-a2a-card  - Test A2A agent card endpoint"
	@echo "  test-a2a-health - Test A2A health endpoint"
	@echo "  test-a2a-task  - Test A2A task creation"
	@echo "  demo-multi-agent - Run multi-agent orchestrator demo"
	@echo ""
	@echo "Kind Cluster Testing:"
	@echo "  kind-setup     - Create kind cluster with Flux CD"
	@echo "  kind-test      - Run MCP tests against kind cluster"
	@echo "  kind-all       - Full setup and test"
	@echo "  kind-cleanup   - Delete kind cluster"
	@echo "  kind-interactive - Interactive MCP testing"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps           - Download dependencies"
	@echo "  tidy           - Tidy dependencies"
	@echo "  verify         - Verify dependencies"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  generate       - Run go generate"
	@echo ""
	@echo "Install:"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  install-global - Install to /usr/local/bin"
	@echo ""
	@echo "Docker & Helm:"
	@echo "  docker-build-a2a - Build A2A Docker image"
	@echo "  docker-push-a2a  - Push A2A Docker image to registry"
	@echo "  helm-lint-a2a    - Lint A2A Helm chart"
	@echo "  helm-package-a2a - Package A2A Helm chart"
	@echo "  helm-push-a2a    - Push Helm chart to OCI registry"
	@echo ""
	@echo "Enterprise Deployment:"
	@echo "  helm-install-std       - Install with traditional Deployment"
	@echo "  helm-install-mcpserver - Install with Kagent MCPServer CRD"
	@echo "  helm-install-enterprise - Install with enterprise values"
	@echo "  install-kagent         - Install Kagent operator (prereq for MCPServer)"
	@echo ""
	@echo "A2A E2E Testing:"
	@echo "  e2e-a2a-all      - Full E2E test (setup + deploy + test)"
	@echo "  e2e-a2a-setup    - Setup kind cluster with dependencies"
	@echo "  e2e-a2a-test     - Run A2A endpoint tests"
	@echo "  e2e-a2a-cleanup  - Cleanup E2E environment"
	@echo ""
	@echo "Other:"
	@echo "  help           - Show this help"
	@echo ""
	@echo "A2A Learning Path:"
	@echo "  1. make build-a2a           # Build the A2A server"
	@echo "  2. make run-a2a             # Start the server"
	@echo "  3. make test-a2a-card       # Test discovery"
	@echo "  4. make demo-multi-agent    # Run the demo"
	@echo "  5. See docs/A2A_LEARNING_GUIDE.md for more"
	@echo ""
	@echo "Troubleshooting Examples:"
	@echo "  build-troubleshooter        - Build troubleshooting example"
	@echo "  run-troubleshooter          - Run specific workflow (WORKFLOW=gitops-failure)"
	@echo "  test-troubleshooter         - Test all workflows"
	@echo "  troubleshoot-workflows-json - Show all workflows as JSON"
	@echo "  See examples/troubleshooting/README.md for learning guide"
