# dm-nkp-gitops-kmcp-server

Kubernetes-native MCP/A2A server for GitOps monitoring using **kmcp** and **Kagent**.

This is a parallel implementation of the Go-based `dm-nkp-gitops-a2a-server` using the K8s-native approach.

## Architecture Comparison

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CUSTOM (Go) vs K8S-NATIVE (kmcp)                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  CUSTOM (../cmd/a2a-server/)              K8S-NATIVE (this directory)       │
│  ─────────────────────────────            ───────────────────────────       │
│                                                                             │
│  ┌─────────────────────────┐              ┌─────────────────────────┐       │
│  │  Go Binary              │              │  Python + FastMCP       │       │
│  │  - Custom JSON-RPC      │              │  - Standard MCP SDK     │       │
│  │  - Custom A2A impl      │              │  - Kagent A2A support   │       │
│  │  - Manual RBAC          │              │  - CRD-based RBAC       │       │
│  └───────────┬─────────────┘              └───────────┬─────────────┘       │
│              │                                        │                     │
│              ▼                                        ▼                     │
│  ┌─────────────────────────┐              ┌─────────────────────────┐       │
│  │  Helm Chart             │              │  Kagent MCPServer CRD   │       │
│  │  - Deployment           │              │  - Declarative config   │       │
│  │  - Service              │              │  - Auto-managed         │       │
│  │  - RBAC (manual)        │              │  - Built-in RBAC        │       │
│  │  - HTTPRoute (manual)   │              │  - Auto HTTPRoute       │       │
│  └─────────────────────────┘              └─────────────────────────┘       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Prerequisites

```bash
# Install kmcp CLI
pip install kmcp

# Install Kagent operator in your cluster
kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml

# Verify
kubectl get crd mcpservers.kagent.dev
```

## Quick Start

```bash
# 1. Install Python dependencies
cd kmcp-server
pip install -e .

# 2. Run locally (for development)
kmcp run

# 3. Deploy to Kubernetes
kubectl apply -f k8s/
```

## Project Structure

```
kmcp-server/
├── src/
│   └── gitops_mcp/
│       ├── __init__.py
│       ├── server.py           # FastMCP server with all tools
│       └── tools/
│           ├── __init__.py
│           ├── flux.py         # Flux/GitOps tools
│           ├── cluster.py      # CAPI cluster tools
│           ├── apps.py         # Kommander app tools
│           └── policy.py       # Gatekeeper/Kyverno tools
├── k8s/
│   ├── namespace.yaml          # Namespace
│   ├── mcpserver.yaml          # Kagent MCPServer CRD
│   ├── secrets.yaml            # Secret references
│   └── kustomization.yaml      # Kustomize overlay
├── tests/
│   └── test_tools.py
├── pyproject.toml
├── Dockerfile
└── README.md
```

## Why K8s-Native?

| Aspect | Custom (Go) | K8s-Native (kmcp) |
|--------|-------------|-------------------|
| **Deployment** | Helm chart + manual YAML | Single MCPServer CRD |
| **RBAC** | Manual ClusterRole/Binding | Auto-managed by Kagent |
| **Secrets** | Manual Secret mounting | Declarative in CRD |
| **Scaling** | Manual HPA | Built-in scaling |
| **Upgrades** | Helm upgrade | CRD spec change |
| **Multi-agent** | Manual coordination | Kagent orchestration |
| **Observability** | Manual Prometheus | Built-in metrics |

## Development

```bash
# Run tests
pytest tests/

# Run locally with hot reload
kmcp run --reload

# Build container
docker build -t ghcr.io/deepak-muley/dm-nkp-gitops-kmcp-server:latest .
```

## Deployment to NKP

```bash
# Deploy with Kustomize
kubectl apply -k k8s/

# Or deploy with kmcp CLI
kmcp deploy --namespace gitops-agent --image ghcr.io/deepak-muley/dm-nkp-gitops-kmcp-server:latest
```

## Migration Path

To migrate from custom Go implementation to kmcp:

1. Keep Go server running during transition
2. Deploy kmcp server alongside
3. Test both implementations
4. Switch traffic to kmcp server
5. Decommission Go server

Both servers expose the same tools with the same names, so clients don't need changes.
