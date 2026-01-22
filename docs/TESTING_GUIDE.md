# Testing Guide

Complete guide for testing the dm-nkp-gitops-custom-mcp-server both locally and in production.

## Table of Contents

1. [Local End-to-End Testing](#local-end-to-end-testing)
2. [Production Testing](#production-testing)
3. [Manual Testing](#manual-testing)
4. [Troubleshooting Tests](#troubleshooting-tests)

---

## Local End-to-End Testing

### Quick Start: MCP Server E2E Test

The easiest way to test the MCP server end-to-end is using the kind cluster test script:

```bash
# Full E2E test (creates cluster, installs Flux, tests all tools)
make kind-all

# Or step by step:
make kind-setup    # Create kind cluster with Flux CD
make kind-test     # Run MCP tests
make kind-cleanup  # Clean up
```

**What it does:**
- Creates a local kind cluster
- Installs Flux CD
- Creates sample GitOps resources (Kustomizations, GitRepositories, HelmReleases)
- Builds the MCP server binary
- Tests all MCP tools against the cluster
- Provides summary and cleanup

**Prerequisites:**
- `kind` - Kubernetes in Docker
- `kubectl` - Kubernetes CLI
- `flux` CLI (optional, will use kubectl if not found)
- `jq` - JSON processor

**Install prerequisites:**
```bash
# macOS
brew install kind kubectl jq
brew install fluxcd/tap/flux

# Linux
# See: https://kind.sigs.k8s.io/docs/user/quick-start/
```

### Quick Start: A2A Server E2E Test

For testing the A2A (Agent-to-Agent) server:

```bash
# Full E2E test (creates cluster, deploys A2A server, tests endpoints)
make e2e-a2a-all

# Or step by step:
make e2e-a2a-setup    # Setup kind cluster with dependencies
make e2e-a2a-test     # Run A2A endpoint tests
make e2e-a2a-cleanup  # Clean up
```

**What it does:**
- Creates a kind cluster
- Installs Gateway API, cert-manager, Traefik
- Installs Flux CD
- Builds and deploys A2A server via Helm
- Tests all A2A endpoints (health, agent card, skills)
- Provides summary

**Prerequisites:**
- Same as MCP server, plus:
- `docker` - For building images
- `helm` - For deploying the chart

### Using the Test Scripts Directly

You can also run the scripts directly for more control:

#### MCP Server Test Script

```bash
# Full test
./scripts/test-with-kind.sh all

# Individual steps
./scripts/test-with-kind.sh setup      # Create cluster
./scripts/test-with-kind.sh build      # Build binary
./scripts/test-with-kind.sh test       # Run tests
./scripts/test-with-kind.sh interactive # Interactive mode
./scripts/test-with-kind.sh cleanup    # Clean up
```

**Environment variables:**
```bash
CLUSTER_NAME=mcp-test \
KUBECONFIG_PATH=~/.kube/kind-mcp-test.conf \
./scripts/test-with-kind.sh all
```

#### A2A Server Test Script

```bash
# Full test
./scripts/e2e-a2a-test.sh all

# Individual steps
./scripts/e2e-a2a-test.sh setup    # Setup environment
./scripts/e2e-a2a-test.sh build    # Build Docker image
./scripts/e2e-a2a-test.sh deploy   # Deploy with Helm
./scripts/e2e-a2a-test.sh test     # Run tests
./scripts/e2e-a2a-test.sh cleanup  # Clean up
```

**Environment variables:**
```bash
CLUSTER_NAME=a2a-e2e-test \
IMAGE_TAG=e2e-test \
HTTP_PORT=8880 \
HTTPS_PORT=8443 \
./scripts/e2e-a2a-test.sh all
```

### What Gets Tested

#### MCP Server Tests

The `test-with-kind.sh` script tests:

1. **Initialize** - MCP protocol initialization
2. **List Tools** - All available tools
3. **Get Current Context** - Kubernetes context
4. **List Kustomizations** - Flux Kustomizations
5. **Get GitOps Status** - Overall health
6. **List GitRepositories** - Git sources
7. **Get Events** - Kubernetes events
8. **Get HelmReleases** - Helm releases

#### A2A Server Tests

The `e2e-a2a-test.sh` script tests:

1. **Health Check** - `/health` endpoint
2. **Agent Card** - `/.well-known/agent.json` discovery
3. **List Contexts Skill** - Execute `list-contexts` skill
4. **Get GitOps Status Skill** - Execute `get-gitops-status` skill
5. **List Kustomizations Skill** - Execute `list-kustomizations` skill

---

## Production Testing

Once deployed to production (NKP or any Kubernetes cluster), use these methods to verify it's working correctly.

### 1. Verify Deployment Status

```bash
# Check MCPServer CRD (if using MCPServer mode)
kubectl get mcpserver -n gitops-agent

# Check pods
kubectl get pods -n gitops-agent

# Check service
kubectl get svc -n gitops-agent

# Check HTTPRoute (if using Gateway API)
kubectl get httproute -n gitops-agent

# Check certificate (if using TLS)
kubectl get certificate -n gitops-agent
```

### 2. Test Health Endpoint

```bash
# Get the service URL
SERVICE_URL=$(kubectl get httproute -n gitops-agent -o jsonpath='{.items[0].spec.hostnames[0]}')
echo "Testing: https://${SERVICE_URL}/health"

# Test health (from within cluster)
kubectl run -it --rm test-curl --image=curlimages/curl --restart=Never -- \
  curl -k https://${SERVICE_URL}/health

# Or port-forward and test locally
kubectl port-forward -n gitops-agent svc/dm-nkp-gitops-a2a-server 8080:8080
curl http://localhost:8080/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 3. Test Agent Card Discovery

```bash
# Test agent card endpoint
curl -k https://${SERVICE_URL}/.well-known/agent.json | jq

# Or via port-forward
curl http://localhost:8080/.well-known/agent.json | jq
```

**Expected response:**
```json
{
  "name": "dm-nkp-gitops-a2a-server",
  "version": "0.2.0",
  "description": "GitOps troubleshooting agent",
  "skills": [
    {
      "name": "list-contexts",
      "description": "List Kubernetes contexts"
    },
    ...
  ]
}
```

### 4. Test A2A Skills (JSON-RPC)

```bash
# Test list-contexts skill
curl -k -X POST https://${SERVICE_URL}/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tasks/create",
    "params": {
      "skill": "list-contexts",
      "input": {}
    }
  }' | jq

# Test get-gitops-status skill
curl -k -X POST https://${SERVICE_URL}/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tasks/create",
    "params": {
      "skill": "get-gitops-status",
      "input": {}
    }
  }' | jq

# Test list-kustomizations skill
curl -k -X POST https://${SERVICE_URL}/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tasks/create",
    "params": {
      "skill": "list-kustomizations",
      "input": {
        "namespace": "flux-system"
      }
    }
  }' | jq
```

### 5. Test MCP Server (if running locally)

If you're running the MCP server locally and connecting to production cluster:

```bash
# Set kubeconfig to production cluster
export KUBECONFIG=/path/to/production-kubeconfig

# Test MCP protocol
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq

# Test a specific tool
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_gitops_status","arguments":{}}}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq
```

### 6. Check Logs

```bash
# View pod logs
kubectl logs -l app.kubernetes.io/name=dm-nkp-gitops-a2a-server -n gitops-agent -f

# View logs for specific pod
kubectl logs -n gitops-agent deployment/dm-nkp-gitops-a2a-server -f

# View logs with timestamps
kubectl logs -n gitops-agent deployment/dm-nkp-gitops-a2a-server --timestamps
```

### 7. Check Metrics (if enabled)

```bash
# Port-forward metrics endpoint
kubectl port-forward -n gitops-agent svc/dm-nkp-gitops-a2a-server 9090:9090

# View metrics
curl http://localhost:9090/metrics

# Or if using Prometheus ServiceMonitor
kubectl get servicemonitor -n gitops-agent
```

### 8. Test from Cursor/Claude Desktop

Once deployed, configure Cursor to use the production endpoint:

**Option A: Use HTTP endpoint (if exposed)**
```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "url": "https://gitops-agent.your-domain.com",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

**Option B: Use local binary with production kubeconfig**
```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "/path/to/production-kubeconfig"
      }
    }
  }
}
```

Then test in Cursor:
```
"What tools are available from the GitOps MCP server?"
"What's the GitOps status in flux-system namespace?"
```

---

## Manual Testing

### Quick MCP Protocol Test

```bash
# Build first
make build

# Test tools/list
make test-mcp

# Test all methods
make test-all-methods
```

### Interactive Testing

```bash
# Start interactive MCP server
make run-readonly

# In another terminal, send requests
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only
```

### Test Individual Tools

```bash
# List all tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq '.result.tools[].name'

# Get current context
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_current_context","arguments":{}}}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq

# Get GitOps status
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_gitops_status","arguments":{}}}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq

# List Kustomizations
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"list_kustomizations","arguments":{"namespace":"flux-system"}}}' | \
  ./bin/dm-nkp-gitops-mcp-server serve --read-only | jq
```

---

## Troubleshooting Tests

### Test Troubleshooting Workflows

The repository includes example troubleshooting workflows:

```bash
# Build troubleshooter
make build-troubleshooter

# Run specific workflow
make run-troubleshooter WORKFLOW=gitops-failure
make run-troubleshooter WORKFLOW=cluster-node
make run-troubleshooter WORKFLOW=app-deployment

# Test all workflows
make test-troubleshooter

# See workflows as JSON
make troubleshoot-workflows-json
```

### Common Test Issues

**Issue: "kind cluster already exists"**
```bash
# Delete existing cluster
kind delete cluster --name mcp-test

# Or use different name
CLUSTER_NAME=my-test ./scripts/test-with-kind.sh all
```

**Issue: "Cannot connect to Kubernetes"**
```bash
# Check kubeconfig
kubectl config current-context

# Set correct kubeconfig
export KUBECONFIG=/path/to/kubeconfig

# Verify connection
kubectl get nodes
```

**Issue: "Flux not installed"**
```bash
# Install Flux CLI
brew install fluxcd/tap/flux

# Or use kubectl method (script will handle this)
```

**Issue: "Port already in use"**
```bash
# Use different ports
HTTP_PORT=8880 HTTPS_PORT=8443 ./scripts/e2e-a2a-test.sh all
```

**Issue: "Certificate not ready"**
```bash
# Check certificate status
kubectl get certificate -n gitops-agent

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager
```

---

## Test Checklist

### Pre-Deployment Testing

- [ ] Run `make kind-all` - MCP server E2E test passes
- [ ] Run `make e2e-a2a-all` - A2A server E2E test passes
- [ ] Run `make test` - Unit tests pass
- [ ] Run `make lint` - Code quality checks pass
- [ ] Test with local kubeconfig pointing to staging cluster

### Post-Deployment Testing

- [ ] Health endpoint returns `200 OK`
- [ ] Agent card endpoint returns valid JSON
- [ ] All A2A skills execute successfully
- [ ] Pods are running and healthy
- [ ] Service is accessible
- [ ] HTTPRoute is configured correctly (if using Gateway API)
- [ ] TLS certificate is valid (if using TLS)
- [ ] Logs show no errors
- [ ] Metrics are being collected (if enabled)
- [ ] Test from Cursor/Claude Desktop

### Production Validation

- [ ] Read-only mode is enabled
- [ ] Network policies are configured
- [ ] RBAC permissions are minimal
- [ ] Resource limits are appropriate
- [ ] High availability is configured (multiple replicas)
- [ ] Monitoring and alerting are set up
- [ ] Backup/restore procedures are tested

---

## Related Documentation

- [Deployment and Usage Guide](DEPLOYMENT_AND_USAGE.md) - How to deploy
- [NKP Production Deployment](NKP_PRODUCTION_DEPLOYMENT.md) - Production deployment guide
- [Tools Reference](TOOLS_REFERENCE.md) - All available tools
- [Troubleshooting Runbook](K8S_TROUBLESHOOTING_RUNBOOK.md) - Troubleshooting procedures

---

## Quick Reference

```bash
# Local E2E Testing
make kind-all              # MCP server E2E
make e2e-a2a-all           # A2A server E2E

# Production Testing
kubectl get pods -n gitops-agent
curl https://gitops-agent.domain.com/health
curl https://gitops-agent.domain.com/.well-known/agent.json | jq

# Manual Testing
make build
make test-mcp
make test-all-methods
```
