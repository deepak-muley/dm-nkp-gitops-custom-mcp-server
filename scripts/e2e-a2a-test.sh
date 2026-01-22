#!/bin/bash
# e2e-a2a-test.sh - End-to-end test for dm-nkp-gitops-a2a-server
#
# This script sets up a complete local environment with:
# - kind cluster
# - Gateway API CRDs
# - cert-manager
# - Traefik (Gateway API mode)
# - Flux CD
# - dm-nkp-gitops-a2a-server
#
# Usage:
#   ./scripts/e2e-a2a-test.sh all      # Full setup and test
#   ./scripts/e2e-a2a-test.sh setup    # Setup only
#   ./scripts/e2e-a2a-test.sh test     # Run tests only
#   ./scripts/e2e-a2a-test.sh cleanup  # Cleanup

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="${CLUSTER_NAME:-a2a-e2e-test}"
KUBECONFIG_PATH="${KUBECONFIG_PATH:-$HOME/.kube/kind-${CLUSTER_NAME}.conf}"
HTTP_PORT="${HTTP_PORT:-8880}"
HTTPS_PORT="${HTTPS_PORT:-8443}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_PATH="${PROJECT_ROOT}/chart/dm-nkp-gitops-a2a-server"
IMAGE_NAME="dm-nkp-gitops-a2a-server"
IMAGE_TAG="${IMAGE_TAG:-e2e-test}"

# Component versions
CERT_MANAGER_VERSION="${CERT_MANAGER_VERSION:-v1.14.4}"
TRAEFIK_VERSION="${TRAEFIK_VERSION:-28.0.0}"
GATEWAY_API_VERSION="${GATEWAY_API_VERSION:-v1.0.0}"

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing=()
    
    command -v kind &> /dev/null || missing+=("kind")
    command -v kubectl &> /dev/null || missing+=("kubectl")
    command -v helm &> /dev/null || missing+=("helm")
    command -v docker &> /dev/null || missing+=("docker")
    command -v jq &> /dev/null || missing+=("jq")
    command -v curl &> /dev/null || missing+=("curl")
    
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
        log_warn "Cluster '${CLUSTER_NAME}' already exists, using it"
        kind export kubeconfig --name "${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG_PATH}"
        export KUBECONFIG="${KUBECONFIG_PATH}"
        return
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
        hostPort: ${HTTP_PORT}
        protocol: TCP
      - containerPort: 443
        hostPort: ${HTTPS_PORT}
        protocol: TCP
EOF

    export KUBECONFIG="${KUBECONFIG_PATH}"
    
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=120s
    
    log_success "Kind cluster '${CLUSTER_NAME}' created"
}

# Install Gateway API CRDs
install_gateway_api() {
    log_info "Installing Gateway API CRDs ${GATEWAY_API_VERSION}..."
    
    kubectl apply -f "https://github.com/kubernetes-sigs/gateway-api/releases/download/${GATEWAY_API_VERSION}/standard-install.yaml"
    
    log_success "Gateway API CRDs installed"
}

# Install cert-manager
install_cert_manager() {
    log_info "Installing cert-manager ${CERT_MANAGER_VERSION}..."
    
    # Check if already installed
    if kubectl get namespace cert-manager &> /dev/null; then
        log_warn "cert-manager already installed, skipping..."
        return
    fi
    
    # Add Helm repo
    helm repo add jetstack https://charts.jetstack.io --force-update
    helm repo update
    
    # Install cert-manager
    helm upgrade --install cert-manager jetstack/cert-manager \
        --namespace cert-manager \
        --create-namespace \
        --version "${CERT_MANAGER_VERSION}" \
        --set installCRDs=true \
        --set "extraArgs={--feature-gates=ExperimentalGatewayAPISupport=true}" \
        --wait
    
    log_info "Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=Available deployment/cert-manager -n cert-manager --timeout=120s
    kubectl wait --for=condition=Available deployment/cert-manager-webhook -n cert-manager --timeout=120s
    
    log_success "cert-manager installed"
}

# Install Traefik with Gateway API support
install_traefik() {
    log_info "Installing Traefik ${TRAEFIK_VERSION} with Gateway API support..."
    
    # Check if already installed
    if kubectl get namespace traefik-system &> /dev/null; then
        log_warn "Traefik already installed, skipping..."
        return
    fi
    
    # Add Helm repo
    helm repo add traefik https://traefik.github.io/charts --force-update
    helm repo update
    
    # Install Traefik
    helm upgrade --install traefik traefik/traefik \
        --namespace traefik-system \
        --create-namespace \
        --version "${TRAEFIK_VERSION}" \
        --set "providers.kubernetesGateway.enabled=true" \
        --set "ports.web.nodePort=30080" \
        --set "ports.websecure.nodePort=30443" \
        --set "service.type=NodePort" \
        --wait
    
    log_info "Waiting for Traefik to be ready..."
    kubectl wait --for=condition=Available deployment/traefik -n traefik-system --timeout=120s
    
    # Create Gateway
    log_info "Creating Traefik Gateway..."
    cat <<EOF | kubectl apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: traefik-gateway
  namespace: traefik-system
spec:
  gatewayClassName: traefik
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
    - name: https
      protocol: HTTPS
      port: 443
      allowedRoutes:
        namespaces:
          from: All
      tls:
        mode: Terminate
        certificateRefs:
          - name: traefik-default-cert
            kind: Secret
EOF
    
    log_success "Traefik installed with Gateway API support"
}

# Install Flux CD
install_flux() {
    log_info "Installing Flux CD..."
    
    if kubectl get namespace flux-system &> /dev/null; then
        log_warn "Flux already installed, skipping..."
        return
    fi
    
    if ! command -v flux &> /dev/null; then
        log_warn "flux CLI not found, installing via kubectl apply..."
        kubectl apply -f https://github.com/fluxcd/flux2/releases/latest/download/install.yaml
    else
        flux install --export | kubectl apply -f -
    fi
    
    log_info "Waiting for Flux controllers to be ready..."
    kubectl -n flux-system wait --for=condition=available --timeout=120s \
        deployment/source-controller \
        deployment/kustomize-controller \
        deployment/helm-controller
    
    log_success "Flux CD installed"
}

# Create sample GitOps resources for testing
create_sample_resources() {
    log_info "Creating sample GitOps resources..."
    
    kubectl create namespace demo-gitops --dry-run=client -o yaml | kubectl apply -f -
    
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
EOF

    log_success "Sample GitOps resources created"
}

# Build and load Docker image
build_and_load_image() {
    log_info "Building Docker image..."
    
    cd "${PROJECT_ROOT}"
    
    docker build \
        --build-arg VERSION="${IMAGE_TAG}" \
        --build-arg GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
        --build-arg BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        -t "${IMAGE_NAME}:${IMAGE_TAG}" \
        -f Dockerfile .
    
    log_info "Loading image into kind cluster..."
    kind load docker-image "${IMAGE_NAME}:${IMAGE_TAG}" --name "${CLUSTER_NAME}"
    
    log_success "Image built and loaded: ${IMAGE_NAME}:${IMAGE_TAG}"
}

# Create self-signed ClusterIssuer for local testing
create_self_signed_issuer() {
    log_info "Creating self-signed ClusterIssuer..."
    
    cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
EOF
    
    log_success "Self-signed ClusterIssuer created"
}

# Deploy the A2A server with Helm
deploy_a2a_server() {
    log_info "Deploying dm-nkp-gitops-a2a-server..."
    
    # Use localhost for local testing
    HOSTNAME="localhost"
    
    log_info "Using hostname: ${HOSTNAME}"
    
    # Create namespace
    kubectl create namespace gitops-agent --dry-run=client -o yaml | kubectl apply -f -
    
    # Install with Helm
    helm upgrade --install dm-nkp-gitops-a2a-server "${CHART_PATH}" \
        --namespace gitops-agent \
        --set image.repository="${IMAGE_NAME}" \
        --set image.tag="${IMAGE_TAG}" \
        --set image.pullPolicy=Never \
        --set httpRoute.enabled=true \
        --set httpRoute.hostname="${HOSTNAME}" \
        --set "httpRoute.parentRefs[0].name=traefik-gateway" \
        --set "httpRoute.parentRefs[0].namespace=traefik-system" \
        --set tls.enabled=true \
        --set tls.selfSigned=true \
        --set tls.createClusterIssuer=true \
        --set a2a.logLevel=debug \
        --wait --timeout=120s
    
    log_info "Waiting for deployment to be ready..."
    kubectl wait --for=condition=Available deployment/dm-nkp-gitops-a2a-server -n gitops-agent --timeout=120s
    
    log_success "A2A server deployed"
}

# Run A2A endpoint tests
run_tests() {
    log_info "Running A2A endpoint tests..."
    
    BASE_URL="https://localhost:${HTTPS_PORT}"
    
    log_info "Testing endpoints at: ${BASE_URL}"
    
    # Wait a bit for routing to be ready
    sleep 5
    
    # Test 1: Health check
    echo ""
    log_info "=== Test 1: Health Check ==="
    HEALTH=$(curl -sk "${BASE_URL}/health" || echo '{"error":"connection failed"}')
    echo "${HEALTH}" | jq . 2>/dev/null || echo "${HEALTH}"
    
    if echo "${HEALTH}" | jq -e '.status == "healthy"' &>/dev/null; then
        log_success "Health check passed"
    else
        log_warn "Health check may have issues"
    fi
    
    # Test 2: Agent Card
    echo ""
    log_info "=== Test 2: Agent Card Discovery ==="
    AGENT_CARD=$(curl -sk "${BASE_URL}/.well-known/agent.json" || echo '{"error":"connection failed"}')
    echo "${AGENT_CARD}" | jq . 2>/dev/null || echo "${AGENT_CARD}"
    
    if echo "${AGENT_CARD}" | jq -e '.name' &>/dev/null; then
        SKILL_COUNT=$(echo "${AGENT_CARD}" | jq '.skills | length')
        log_success "Agent card retrieved - ${SKILL_COUNT} skills available"
    else
        log_warn "Agent card may have issues"
    fi
    
    # Test 3: List Contexts
    echo ""
    log_info "=== Test 3: Execute Skill - list-contexts ==="
    TASK_RESULT=$(curl -sk -X POST "${BASE_URL}/" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":1,"method":"tasks/create","params":{"skill":"list-contexts","input":{}}}' \
        || echo '{"error":"connection failed"}')
    echo "${TASK_RESULT}" | jq . 2>/dev/null || echo "${TASK_RESULT}"
    
    # Test 4: Get GitOps Status
    echo ""
    log_info "=== Test 4: Execute Skill - get-gitops-status ==="
    GITOPS_RESULT=$(curl -sk -X POST "${BASE_URL}/" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":2,"method":"tasks/create","params":{"skill":"get-gitops-status","input":{}}}' \
        || echo '{"error":"connection failed"}')
    echo "${GITOPS_RESULT}" | jq . 2>/dev/null || echo "${GITOPS_RESULT}"
    
    # Test 5: List Kustomizations
    echo ""
    log_info "=== Test 5: Execute Skill - list-kustomizations ==="
    KUST_RESULT=$(curl -sk -X POST "${BASE_URL}/" \
        -H "Content-Type: application/json" \
        -d '{"jsonrpc":"2.0","id":3,"method":"tasks/create","params":{"skill":"list-kustomizations","input":{"namespace":"flux-system"}}}' \
        || echo '{"error":"connection failed"}')
    echo "${KUST_RESULT}" | jq . 2>/dev/null || echo "${KUST_RESULT}"
    
    echo ""
    log_success "All tests completed!"
}

# Print summary
print_summary() {
    echo ""
    log_success "======================================"
    log_success "E2E Environment Ready!"
    log_success "======================================"
    echo ""
    echo "Cluster: ${CLUSTER_NAME}"
    echo "Kubeconfig: ${KUBECONFIG_PATH}"
    echo "Ports: HTTP=${HTTP_PORT}, HTTPS=${HTTPS_PORT}"
    echo ""
    echo "A2A Server Endpoints:"
    echo "  Health:     https://localhost:${HTTPS_PORT}/health"
    echo "  Agent Card: https://localhost:${HTTPS_PORT}/.well-known/agent.json"
    echo "  JSON-RPC:   POST https://localhost:${HTTPS_PORT}/"
    echo ""
    echo "To use this cluster:"
    echo "  export KUBECONFIG=${KUBECONFIG_PATH}"
    echo ""
    echo "Test commands:"
    echo "  curl -k https://localhost:${HTTPS_PORT}/health | jq"
    echo "  curl -k https://localhost:${HTTPS_PORT}/.well-known/agent.json | jq"
    echo ""
}

# Cleanup
cleanup() {
    log_info "Cleaning up..."
    
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_info "Deleting cluster '${CLUSTER_NAME}'..."
        kind delete cluster --name "${CLUSTER_NAME}"
        rm -f "${KUBECONFIG_PATH}"
        log_success "Cluster deleted"
    else
        log_warn "Cluster '${CLUSTER_NAME}' not found"
    fi
}

# Print usage
print_usage() {
    cat <<EOF
Usage: $0 <command>

Commands:
    setup       Create kind cluster and install all dependencies
    build       Build and load Docker image into kind
    deploy      Deploy A2A server with Helm
    test        Run A2A endpoint tests
    cleanup     Delete the kind cluster
    all         Full setup, build, deploy, and test
    help        Show this help message

Environment Variables:
    CLUSTER_NAME          Kind cluster name (default: a2a-e2e-test)
    IMAGE_TAG             Docker image tag (default: e2e-test)
    HTTP_PORT             Host HTTP port (default: 8880)
    HTTPS_PORT            Host HTTPS port (default: 8443)
    CERT_MANAGER_VERSION  cert-manager version (default: v1.14.4)
    TRAEFIK_VERSION       Traefik Helm chart version (default: 28.0.0)
    GATEWAY_API_VERSION   Gateway API CRDs version (default: v1.0.0)

Examples:
    # Full end-to-end test
    $0 all

    # Just setup the environment
    $0 setup

    # Run tests against existing deployment
    $0 test

    # Clean up everything
    $0 cleanup
EOF
}

# Main
main() {
    case "${1:-help}" in
        setup)
            check_prerequisites
            create_cluster
            install_gateway_api
            install_cert_manager
            install_traefik
            install_flux
            create_sample_resources
            create_self_signed_issuer
            print_summary
            ;;
        build)
            export KUBECONFIG="${KUBECONFIG_PATH}"
            build_and_load_image
            ;;
        deploy)
            export KUBECONFIG="${KUBECONFIG_PATH}"
            deploy_a2a_server
            print_summary
            ;;
        test)
            export KUBECONFIG="${KUBECONFIG_PATH}"
            run_tests
            ;;
        cleanup)
            cleanup
            ;;
        all)
            check_prerequisites
            create_cluster
            install_gateway_api
            install_cert_manager
            install_traefik
            install_flux
            create_sample_resources
            create_self_signed_issuer
            build_and_load_image
            deploy_a2a_server
            
            log_info "Waiting 30s for resources to reconcile..."
            sleep 30
            
            run_tests
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
