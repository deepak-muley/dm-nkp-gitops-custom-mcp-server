#!/bin/bash
# test-with-kind.sh - Test dm-nkp-gitops-mcp-server with a local kind cluster
#
# This script sets up a kind cluster with Flux CD and sample resources
# to test the MCP server locally.

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="${CLUSTER_NAME:-mcp-test}"
# Leave empty to use the installed flux CLI version (recommended)
FLUX_VERSION="${FLUX_VERSION:-}"
KUBECONFIG_PATH="${KUBECONFIG_PATH:-$HOME/.kube/kind-${CLUSTER_NAME}.conf}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BINARY_PATH="${PROJECT_ROOT}/bin/dm-nkp-gitops-mcp-server"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    if ! command -v kind &> /dev/null; then
        missing+=("kind")
    fi
    
    if ! command -v kubectl &> /dev/null; then
        missing+=("kubectl")
    fi
    
    if ! command -v flux &> /dev/null; then
        log_warn "flux CLI not found - will skip Flux installation"
        log_warn "Install with: brew install fluxcd/tap/flux"
    fi
    
    if ! command -v jq &> /dev/null; then
        missing+=("jq")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing[*]}"
        log_info "Please install them and try again"
        exit 1
    fi
    
    log_success "All prerequisites met"
}

# Create kind cluster
create_cluster() {
    log_info "Checking for existing cluster '${CLUSTER_NAME}'..."
    
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_warn "Cluster '${CLUSTER_NAME}' already exists"
        read -p "Delete and recreate? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "Deleting existing cluster..."
            kind delete cluster --name "${CLUSTER_NAME}"
        else
            log_info "Using existing cluster"
            kind export kubeconfig --name "${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG_PATH}"
            export KUBECONFIG="${KUBECONFIG_PATH}"
            return
        fi
    fi
    
    log_info "Creating kind cluster '${CLUSTER_NAME}'..."
    
    cat <<EOF | kind create cluster --name "${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG_PATH}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
EOF

    export KUBECONFIG="${KUBECONFIG_PATH}"
    
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=120s
    
    log_success "Kind cluster '${CLUSTER_NAME}' created successfully"
}

# Install Flux CD
install_flux() {
    if ! command -v flux &> /dev/null; then
        log_warn "Skipping Flux installation (flux CLI not found)"
        return
    fi
    
    # Get installed flux CLI version
    local flux_cli_version
    flux_cli_version=$(flux version --client -o json 2>/dev/null | jq -r '.flux' || flux version --client 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    log_info "Detected Flux CLI version: ${flux_cli_version}"
    
    # Check if Flux is already installed
    if kubectl get namespace flux-system &> /dev/null; then
        log_warn "Flux already installed, skipping..."
        return
    fi
    
    # Install Flux (use CLI version by default, or specific version if set)
    if [ -n "${FLUX_VERSION}" ]; then
        log_info "Installing Flux CD v${FLUX_VERSION}..."
        flux install --version "v${FLUX_VERSION}" --export | kubectl apply -f -
    else
        log_info "Installing Flux CD (using CLI version ${flux_cli_version})..."
        flux install --export | kubectl apply -f -
    fi
    
    log_info "Waiting for Flux controllers to be ready..."
    kubectl -n flux-system wait --for=condition=available --timeout=120s \
        deployment/source-controller \
        deployment/kustomize-controller \
        deployment/helm-controller \
        deployment/notification-controller
    
    log_success "Flux CD installed successfully"
}

# Create sample GitOps resources
create_sample_resources() {
    log_info "Creating sample GitOps resources..."
    
    # Create sample namespace
    kubectl create namespace demo-gitops --dry-run=client -o yaml | kubectl apply -f -
    
    # Create sample GitRepository
    cat <<EOF | kubectl apply -f -
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: podinfo
  namespace: flux-system
spec:
  interval: 5m
  url: https://github.com/stefanprodan/podinfo
  ref:
    branch: master
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: flux-monitoring
  namespace: flux-system
spec:
  interval: 10m
  url: https://github.com/fluxcd/flux2-monitoring-example
  ref:
    branch: main
EOF

    # Create sample Kustomization
    cat <<EOF | kubectl apply -f -
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: podinfo
  namespace: flux-system
spec:
  interval: 10m
  targetNamespace: demo-gitops
  sourceRef:
    kind: GitRepository
    name: podinfo
  path: ./kustomize
  prune: true
  timeout: 5m
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: infrastructure
  namespace: flux-system
spec:
  interval: 15m
  sourceRef:
    kind: GitRepository
    name: flux-monitoring
  path: ./manifests/monitoring/controllers
  prune: true
  wait: true
  timeout: 10m
  suspend: true
EOF

    # Create a sample HelmRepository and HelmRelease
    cat <<EOF | kubectl apply -f -
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: podinfo
  namespace: flux-system
spec:
  interval: 5m
  url: https://stefanprodan.github.io/podinfo
---
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: podinfo-helm
  namespace: flux-system
spec:
  interval: 5m
  chart:
    spec:
      chart: podinfo
      version: '>=6.0.0'
      sourceRef:
        kind: HelmRepository
        name: podinfo
      interval: 1m
  targetNamespace: demo-gitops
  values:
    replicaCount: 2
    resources:
      limits:
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 64Mi
EOF

    log_success "Sample GitOps resources created"
}

# Create sample events for debugging
create_sample_events() {
    log_info "Creating sample deployment and events..."
    
    # Create a deployment that will generate events
    cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  namespace: demo-gitops
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          limits:
            memory: "128Mi"
            cpu: "250m"
EOF

    log_success "Sample deployment created"
}

# Build the MCP server
build_server() {
    log_info "Building MCP server..."
    cd "${PROJECT_ROOT}"
    make build
    log_success "MCP server built at ${BINARY_PATH}"
}

# Test MCP server with JSON-RPC requests
test_mcp_server() {
    log_info "Testing MCP server..."
    
    if [ ! -f "${BINARY_PATH}" ]; then
        log_error "Binary not found at ${BINARY_PATH}"
        log_info "Run 'make build' first"
        exit 1
    fi
    
    export KUBECONFIG="${KUBECONFIG_PATH}"
    
    echo ""
    log_info "=== Test 1: Initialize ==="
    echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 2: List Tools ==="
    echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq '.result.tools[] | {name, description}'
    
    echo ""
    log_info "=== Test 3: Get Current Context ==="
    echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_current_context","arguments":{}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 4: List Kustomizations ==="
    echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_kustomizations","arguments":{"namespace":"flux-system"}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 5: Get GitOps Status ==="
    echo '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_gitops_status","arguments":{}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 6: List GitRepositories ==="
    echo '{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"list_gitrepositories","arguments":{}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 7: Get Events ==="
    echo '{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"get_events","arguments":{"namespace":"flux-system","limit":"5"}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    echo ""
    log_info "=== Test 8: Get HelmReleases ==="
    echo '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"get_helmreleases","arguments":{}}}' | \
        "${BINARY_PATH}" serve --read-only 2>/dev/null | jq .
    
    log_success "All MCP tests completed!"
}

# Interactive testing mode
interactive_test() {
    log_info "Starting interactive MCP test mode..."
    log_info "Enter JSON-RPC requests (one per line). Press Ctrl+D to exit."
    log_info "Example: {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}"
    echo ""
    
    export KUBECONFIG="${KUBECONFIG_PATH}"
    "${BINARY_PATH}" serve --read-only --log-level=debug
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        read -p "Delete cluster '${CLUSTER_NAME}'? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            kind delete cluster --name "${CLUSTER_NAME}"
            rm -f "${KUBECONFIG_PATH}"
            log_success "Cluster deleted"
        fi
    fi
}

# Print usage
print_usage() {
    cat <<EOF
Usage: $0 <command>

Commands:
    setup       Create kind cluster and install Flux CD with sample resources
    build       Build the MCP server binary
    test        Run MCP server tests
    interactive Start interactive MCP testing mode
    cleanup     Delete the kind cluster
    all         Run setup, build, and test
    help        Show this help message

Environment Variables:
    CLUSTER_NAME     Kind cluster name (default: mcp-test)
    FLUX_VERSION     Flux CD version (default: 2.2.3)
    KUBECONFIG_PATH  Path for kubeconfig (default: ~/.kube/kind-\$CLUSTER_NAME.conf)

Examples:
    # Full setup and test
    $0 all

    # Just create the cluster
    $0 setup

    # Run tests against existing cluster
    KUBECONFIG=~/.kube/kind-mcp-test.conf $0 test

    # Interactive testing
    $0 interactive
EOF
}

# Print test summary
print_summary() {
    echo ""
    log_success "======================================"
    log_success "Test environment ready!"
    log_success "======================================"
    echo ""
    echo "Cluster: ${CLUSTER_NAME}"
    echo "Kubeconfig: ${KUBECONFIG_PATH}"
    echo ""
    echo "To use this cluster:"
    echo "  export KUBECONFIG=${KUBECONFIG_PATH}"
    echo ""
    echo "To test the MCP server manually:"
    echo "  echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/list\"}' | ${BINARY_PATH} serve --read-only"
    echo ""
    echo "To configure Cursor, add to ~/.cursor/mcp.json:"
    cat <<EOF
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "${BINARY_PATH}",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "${KUBECONFIG_PATH}"
      }
    }
  }
}
EOF
    echo ""
}

# Main
main() {
    case "${1:-help}" in
        setup)
            check_prerequisites
            create_cluster
            install_flux
            create_sample_resources
            create_sample_events
            print_summary
            ;;
        build)
            build_server
            ;;
        test)
            test_mcp_server
            ;;
        interactive)
            interactive_test
            ;;
        cleanup)
            cleanup
            ;;
        all)
            check_prerequisites
            create_cluster
            install_flux
            create_sample_resources
            create_sample_events
            build_server
            
            # Wait for resources to reconcile
            log_info "Waiting 30s for resources to reconcile..."
            sleep 30
            
            test_mcp_server
            print_summary
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            log_error "Unknown command: $1"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"
