# Enterprise Deployment Guide

This guide covers deploying the dm-nkp-gitops MCP/A2A server in production environments.

## Deployment Modes

The Helm chart supports two deployment modes:

| Mode | `deploymentMode` | When to Use |
|------|------------------|-------------|
| **Traditional** | `deployment` | Default, works on any K8s cluster |
| **K8s-Native** | `mcpserver` | Enterprise recommended, requires Kagent |

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DEPLOYMENT MODE COMPARISON                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  TRADITIONAL (deploymentMode: deployment)                                   │
│  ─────────────────────────────────────────                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                      │
│  │ Deployment   │  │ Service      │  │ HTTPRoute    │                      │
│  │ (manual)     │──│ (manual)     │──│ (manual)     │                      │
│  └──────────────┘  └──────────────┘  └──────────────┘                      │
│                                                                             │
│  K8S-NATIVE (deploymentMode: mcpserver)                                     │
│  ──────────────────────────────────────                                     │
│  ┌──────────────────────────────────────────────────────────┐              │
│  │                    MCPServer CRD                          │              │
│  │  (Kagent operator manages Deployment, Service, etc.)      │              │
│  └──────────────────────────────────────────────────────────┘              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Quick Start

### Option 1: Traditional Deployment (No Prerequisites)

```bash
# Deploy with traditional Kubernetes Deployment
make helm-install-std
```

### Option 2: K8s-Native with Kagent (Enterprise Recommended)

```bash
# Step 1: Install Kagent operator
make install-kagent

# Step 2: Deploy with MCPServer CRD
make helm-install-mcpserver
```

### Option 3: Full Enterprise Configuration

```bash
# Step 1: Install Kagent operator
make install-kagent

# Step 2: Deploy with enterprise values (HA, security hardening, observability)
make helm-install-enterprise
```

## Configuration Reference

### values.yaml Key Settings

```yaml
# Choose deployment mode
deploymentMode: "mcpserver"  # or "deployment"

# Replicas for HA
replicaCount: 2

# Security (always read-only in production)
a2a:
  readOnly: true
  logLevel: "info"

# MCPServer specific (only when deploymentMode: mcpserver)
mcpServer:
  metrics:
    enabled: true
  pdb:
    enabled: true
    minAvailable: 1
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5

# Network security
networkPolicy:
  enabled: true
```

### Enterprise values-enterprise.yaml

The `values-enterprise.yaml` file provides production-ready defaults:

| Feature | Setting |
|---------|---------|
| Deployment Mode | MCPServer CRD |
| Replicas | 2 (min), 5 (max) |
| HPA | Enabled (CPU 70%, Memory 80%) |
| PDB | Enabled (minAvailable: 1) |
| Network Policy | Enabled (zero-trust) |
| Security Context | Non-root, read-only FS |
| Metrics | Enabled (port 9090) |

## Why MCPServer CRD for Enterprise?

| Aspect | Traditional | MCPServer CRD |
|--------|-------------|---------------|
| **Resource Management** | Manual YAML | Auto-managed by operator |
| **Upgrades** | Helm upgrade | CRD spec change (GitOps friendly) |
| **Multi-Agent** | Manual coordination | Kagent handles agent discovery |
| **Observability** | Manual ServiceMonitor | Built-in metrics |
| **A2A Protocol** | Custom implementation | Native support |
| **CNCF Alignment** | N/A | Kagent is CNCF Sandbox |

## Architecture

### With Kagent MCPServer

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Production Cluster                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────┐                                        │
│  │      Kagent Operator            │                                        │
│  │  (watches MCPServer CRDs)       │                                        │
│  └───────────────┬─────────────────┘                                        │
│                  │ manages                                                  │
│                  ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  gitops-agent namespace                                              │   │
│  │  ┌─────────────────────────────────────────────────────────────┐    │   │
│  │  │  MCPServer: dm-nkp-gitops-a2a-server                        │    │   │
│  │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │    │   │
│  │  │  │ Pod (Go)     │  │ Pod (Go)     │  │ ...          │       │    │   │
│  │  │  │ container    │  │ container    │  │ (autoscaled) │       │    │   │
│  │  │  └──────────────┘  └──────────────┘  └──────────────┘       │    │   │
│  │  └─────────────────────────────────────────────────────────────┘    │   │
│  │                                                                      │   │
│  │  Auto-created by Kagent:                                            │   │
│  │  - Deployment                                                       │   │
│  │  - Service                                                          │   │
│  │  - ServiceMonitor (if metrics enabled)                              │   │
│  │  - HPA (if autoscaling enabled)                                     │   │
│  │  - PDB (if pdb enabled)                                             │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Security Hardening

### Pod Security

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532
  runAsGroup: 65532
  fsGroup: 65532
  seccompProfile:
    type: RuntimeDefault

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true
  capabilities:
    drop:
      - ALL
```

### Network Policy (Zero Trust)

```yaml
networkPolicy:
  enabled: true
  ingressFrom:
    - namespaceSelector:
        matchLabels:
          purpose: ai-agents
```

### RBAC (Read-Only)

The ClusterRole only grants `get`, `list`, `watch` permissions - no write access.

## High Availability

### Recommended Settings

```yaml
replicaCount: 2

mcpServer:
  pdb:
    enabled: true
    minAvailable: 1
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchLabels:
              app.kubernetes.io/name: dm-nkp-gitops-a2a-server
          topologyKey: kubernetes.io/hostname

topologySpreadConstraints:
  - maxSkew: 1
    topologyKey: topology.kubernetes.io/zone
    whenUnsatisfiable: ScheduleAnyway
```

## Observability

### Prometheus Metrics

```yaml
mcpServer:
  metrics:
    enabled: true
    port: 9090
    path: /metrics
```

### Distributed Tracing (OpenTelemetry)

```yaml
mcpServer:
  tracing:
    enabled: true
    endpoint: "http://otel-collector.monitoring:4317"
```

## Migration Path

### From Traditional to MCPServer

1. **Prepare**: Install Kagent operator
2. **Deploy**: Apply MCPServer alongside existing Deployment
3. **Verify**: Test both endpoints work
4. **Switch**: Update HTTPRoute to point to MCPServer service
5. **Cleanup**: Remove traditional Deployment

```bash
# Step 1: Install Kagent
make install-kagent

# Step 2: Deploy MCPServer (new release name to avoid conflict)
helm install dm-nkp-gitops-a2a-server-v2 chart/dm-nkp-gitops-a2a-server \
  --set deploymentMode=mcpserver \
  --namespace gitops-agent

# Step 3: Verify
kubectl get mcpserver -n gitops-agent
curl https://gitops-agent-v2.example.com/health

# Step 4: Switch traffic (update HTTPRoute hostname or DNS)

# Step 5: Remove old deployment
helm uninstall dm-nkp-gitops-a2a-server-v1 --namespace gitops-agent
```

## Troubleshooting

### Check MCPServer Status

```bash
kubectl get mcpserver -n gitops-agent
kubectl describe mcpserver dm-nkp-gitops-a2a-server -n gitops-agent
```

### Check Kagent Operator

```bash
kubectl get pods -n kagent-system
kubectl logs -n kagent-system -l app=kagent-controller
```

### View Generated Resources

```bash
# See what Kagent created from MCPServer CRD
kubectl get all -l app.kubernetes.io/name=dm-nkp-gitops-a2a-server -n gitops-agent
```
