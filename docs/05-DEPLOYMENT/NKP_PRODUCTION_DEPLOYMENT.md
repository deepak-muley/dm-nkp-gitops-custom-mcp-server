# NKP Production Deployment Guide

Complete guide for deploying the GitOps MCP/A2A server on Nutanix Kubernetes Platform (NKP) for production use.

## Overview

This repository includes a **production-ready Helm chart** that supports two deployment modes:

| Mode | Description | Best For |
|------|-------------|----------|
| **Traditional** | Standard Kubernetes Deployment | Any K8s cluster, quick setup |
| **MCPServer CRD** | K8s-native via Kagent operator | Production, GitOps, enterprise |

**Recommendation for NKP Production**: Use **MCPServer CRD mode** with enterprise values for:
- ✅ High availability (HA)
- ✅ Auto-scaling
- ✅ Zero-downtime updates
- ✅ GitOps-friendly (CRD-based)
- ✅ Built-in observability

---

## Prerequisites and Dependencies

### Required Components

| Component | Purpose | Installation |
|-----------|---------|--------------|
| **Kagent Operator** | MCPServer CRD support | `kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml` |
| **cert-manager** | TLS certificate management | Usually pre-installed on NKP |
| **Gateway API** | HTTPRoute for ingress | Usually pre-installed on NKP (Traefik/Kong) |
| **Prometheus Operator** | Metrics collection | Optional, for ServiceMonitor |

### Optional but Recommended

| Component | Purpose | When Needed |
|-----------|---------|-------------|
| **Prometheus** | Metrics scraping | If using ServiceMonitor |
| **Grafana** | Metrics visualization | If using Prometheus |
| **OpenTelemetry Collector** | Distributed tracing | If enabling tracing |

### NKP-Specific Considerations

1. **Traefik Gateway**: NKP typically includes Traefik. Configure `httpRoute.parentRefs` accordingly.
2. **cert-manager**: Usually pre-installed. Use existing ClusterIssuer.
3. **Service Mesh**: If using Istio/Linkerd, may need additional configuration.
4. **Network Policies**: NKP may have default policies. Coordinate with platform team.

---

## Deployment Options

### Option 1: MCPServer CRD Mode (Recommended for Production)

**Best for**: Production, GitOps workflows, enterprise environments

#### Step 1: Install Kagent Operator

```bash
# Install Kagent operator (one-time, cluster-wide)
kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml

# Verify installation
kubectl get pods -n kagent-system
kubectl get crd mcpservers.kagent.dev
```

#### Step 2: Create Namespace

```bash
kubectl create namespace gitops-agent
```

#### Step 3: Deploy with Enterprise Values

```bash
# Clone the repository
git clone https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server.git
cd dm-nkp-gitops-custom-mcp-server

# Customize values for NKP
cat > nkp-values.yaml <<EOF
deploymentMode: "mcpserver"

# NKP-specific Gateway configuration
httpRoute:
  enabled: true
  hostname: "gitops-agent.your-nkp-domain.com"  # CHANGE THIS
  parentRefs:
    - name: traefik-gateway  # Or your NKP gateway name
      namespace: traefik-system  # Or your gateway namespace

# TLS with existing cert-manager ClusterIssuer
tls:
  enabled: true
  clusterIssuer: "letsencrypt-prod"  # Use your NKP ClusterIssuer
  createClusterIssuer: false

# Production settings
replicaCount: 2
a2a:
  readOnly: true
  logLevel: "info"

# High availability
mcpServer:
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
  pdb:
    enabled: true
    minAvailable: 1

# Network security
networkPolicy:
  enabled: true
  ingressFrom:
    - namespaceSelector:
        matchLabels:
          purpose: ai-agents
EOF

# Deploy
helm upgrade --install gitops-agent ./chart/dm-nkp-gitops-a2a-server \
  -f ./chart/dm-nkp-gitops-a2a-server/values-enterprise.yaml \
  -f nkp-values.yaml \
  --namespace gitops-agent \
  --create-namespace \
  --wait
```

#### Step 4: Verify Deployment

```bash
# Check MCPServer CRD
kubectl get mcpserver -n gitops-agent

# Check pods
kubectl get pods -n gitops-agent

# Check service
kubectl get svc -n gitops-agent

# Check HTTPRoute
kubectl get httproute -n gitops-agent

# Test health endpoint
curl https://gitops-agent.your-nkp-domain.com/health
```

---

### Option 2: Traditional Deployment Mode

**Best for**: Quick setup, testing, clusters without Kagent

```bash
# Deploy with traditional Deployment
helm upgrade --install gitops-agent ./chart/dm-nkp-gitops-a2a-server \
  --set deploymentMode=deployment \
  --set replicaCount=2 \
  --set a2a.readOnly=true \
  --namespace gitops-agent \
  --create-namespace \
  --wait
```

**Note**: This mode doesn't require Kagent but lacks some enterprise features.

---

## Production Configuration Checklist

### ✅ Security

- [ ] **Read-only mode enabled**: `a2a.readOnly: true`
- [ ] **Non-root container**: `podSecurityContext.runAsNonRoot: true`
- [ ] **Read-only filesystem**: `securityContext.readOnlyRootFilesystem: true`
- [ ] **Network policies**: `networkPolicy.enabled: true`
- [ ] **TLS enabled**: `tls.enabled: true`
- [ ] **RBAC minimal**: ServiceAccount with least-privilege

### ✅ High Availability

- [ ] **Multiple replicas**: `replicaCount: 2` (minimum)
- [ ] **Pod anti-affinity**: Spread across nodes
- [ ] **Pod Disruption Budget**: `mcpServer.pdb.enabled: true`
- [ ] **HPA enabled**: `mcpServer.autoscaling.enabled: true`

### ✅ Observability

- [ ] **Metrics enabled**: `mcpServer.metrics.enabled: true`
- [ ] **ServiceMonitor**: For Prometheus scraping
- [ ] **Health probes**: Liveness and readiness configured
- [ ] **Structured logging**: `logLevel: "info"` or `"warn"`

### ✅ Network

- [ ] **HTTPRoute configured**: For Gateway API ingress
- [ ] **TLS certificate**: Valid and auto-renewing
- [ ] **Network policies**: Zero-trust configured

---

## NKP-Specific Configuration

### Traefik Gateway (Common on NKP)

```yaml
httpRoute:
  enabled: true
  hostname: "gitops-agent.nkp.example.com"
  parentRefs:
    - name: traefik-gateway
      namespace: traefik-system
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
```

### Kong Gateway (Alternative)

```yaml
httpRoute:
  enabled: true
  hostname: "gitops-agent.nkp.example.com"
  parentRefs:
    - name: kong-gateway
      namespace: kong-system
  annotations:
    konghq.com/plugins: rate-limiting-production
```

### cert-manager ClusterIssuer

NKP typically has cert-manager pre-installed. Use existing ClusterIssuer:

```yaml
tls:
  enabled: true
  clusterIssuer: "letsencrypt-prod"  # Your NKP ClusterIssuer name
  createClusterIssuer: false  # Don't create, use existing
```

### Service Account and RBAC

The Helm chart creates a ServiceAccount with minimal RBAC. For NKP, you may need to:

1. **Review RBAC**: Check `chart/dm-nkp-gitops-a2a-server/templates/clusterrole.yaml`
2. **Adjust permissions**: Based on NKP security policies
3. **Use existing SA**: If NKP has a standard service account pattern

```yaml
serviceAccount:
  create: true
  name: ""  # Auto-generated, or specify existing SA
  annotations:
    # Add NKP-specific annotations if needed
```

---

## Resource Sizing

### Production Recommendations

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Scaling Configuration

```yaml
mcpServer:
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 70
    targetMemoryUtilizationPercentage: 80
```

**Sizing Guidelines**:
- **Small cluster** (< 10 namespaces): 1-2 replicas, 100m CPU, 128Mi memory
- **Medium cluster** (10-50 namespaces): 2-3 replicas, 200m CPU, 256Mi memory
- **Large cluster** (50+ namespaces): 3-5 replicas, 500m CPU, 512Mi memory

---

## GitOps Deployment (Flux CD)

### Option A: HelmRelease (Recommended)

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: gitops-agent
  namespace: gitops-agent
spec:
  interval: 5m
  chart:
    spec:
      chart: dm-nkp-gitops-a2a-server
      sourceRef:
        kind: HelmRepository
        name: gitops-charts
      version: "0.2.0"
  values:
    deploymentMode: "mcpserver"
    replicaCount: 2
    a2a:
      readOnly: true
    httpRoute:
      enabled: true
      hostname: "gitops-agent.nkp.example.com"
      parentRefs:
        - name: traefik-gateway
          namespace: traefik-system
    tls:
      enabled: true
      clusterIssuer: "letsencrypt-prod"
```

### Option B: Kustomization with Helm Chart

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: gitops-agent
  namespace: gitops-agent
spec:
  interval: 5m
  path: ./gitops-agent
  prune: true
  sourceRef:
    kind: GitRepository
    name: gitops-infra
```

---

## Monitoring and Alerting

### Prometheus Metrics

The chart includes ServiceMonitor for Prometheus:

```yaml
# ServiceMonitor is auto-created when metrics enabled
mcpServer:
  metrics:
    enabled: true
    port: 9090
    path: /metrics
```

**Key Metrics to Monitor**:
- `http_requests_total` - Request count
- `http_request_duration_seconds` - Response time
- `k8s_api_calls_total` - Kubernetes API call count
- `tool_calls_total` - Tool execution count

### Grafana Dashboard

Create a dashboard with:
- Request rate and latency
- Error rate
- Pod resource usage
- Kubernetes API call metrics

### Alerting Rules

```yaml
# Example Prometheus alert
- alert: GitOpsAgentHighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 5m
  annotations:
    summary: "High error rate on GitOps agent"
```

---

## Troubleshooting

### Check MCPServer Status

```bash
kubectl get mcpserver gitops-agent -n gitops-agent -o yaml
kubectl describe mcpserver gitops-agent -n gitops-agent
```

### Check Pod Logs

```bash
kubectl logs -l app.kubernetes.io/name=dm-nkp-gitops-a2a-server -n gitops-agent -f
```

### Check Service

```bash
kubectl get svc -n gitops-agent
kubectl describe svc gitops-agent -n gitops-agent
```

### Check HTTPRoute

```bash
kubectl get httproute -n gitops-agent
kubectl describe httproute gitops-agent -n gitops-agent
```

### Check Certificate

```bash
kubectl get certificate -n gitops-agent
kubectl describe certificate gitops-agent-tls -n gitops-agent
```

### Test Endpoints

```bash
# Health check
curl https://gitops-agent.your-nkp-domain.com/health

# Agent card
curl https://gitops-agent.your-nkp-domain.com/.well-known/agent.json | jq

# Metrics
curl https://gitops-agent.your-nkp-domain.com/metrics
```

---

## Upgrade Procedure

### Standard Upgrade

```bash
# Update Helm chart
helm upgrade gitops-agent ./chart/dm-nkp-gitops-a2a-server \
  -f nkp-values.yaml \
  --namespace gitops-agent \
  --wait
```

### Rolling Update (MCPServer Mode)

MCPServer CRD mode supports zero-downtime updates:

```yaml
# Update image tag in values
image:
  tag: "0.3.0"

# Apply update
helm upgrade gitops-agent ./chart/dm-nkp-gitops-a2a-server \
  -f nkp-values.yaml \
  --namespace gitops-agent
```

The Kagent operator handles rolling updates automatically with PDB ensuring availability.

---

## Backup and Disaster Recovery

### Configuration Backup

The Helm values file is your configuration backup. Store in Git:

```bash
# Export current values
helm get values gitops-agent -n gitops-agent > nkp-values-backup.yaml

# Commit to Git
git add nkp-values-backup.yaml
git commit -m "Backup: GitOps agent configuration"
```

### Disaster Recovery

1. **Restore from Git**: Helm values are in GitOps repo
2. **Reinstall**: `helm install` with same values
3. **Verify**: Check MCPServer status and pods

---

## Security Hardening Checklist

- [ ] **Read-only mode**: `a2a.readOnly: true` (CRITICAL)
- [ ] **Non-root user**: `runAsNonRoot: true`
- [ ] **Read-only filesystem**: `readOnlyRootFilesystem: true`
- [ ] **Drop all capabilities**: `capabilities.drop: ["ALL"]`
- [ ] **Network policies**: Restrict ingress/egress
- [ ] **TLS only**: No HTTP, only HTTPS
- [ ] **RBAC minimal**: Least-privilege ServiceAccount
- [ ] **Image scanning**: Scan container images
- [ ] **Secrets management**: Use sealed-secrets or external secret operator
- [ ] **Audit logging**: Enable Kubernetes audit logs

---

## Performance Tuning

### For High Load

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi

mcpServer:
  autoscaling:
    minReplicas: 3
    maxReplicas: 10
    targetCPUUtilizationPercentage: 60  # Scale earlier
```

### Connection Pooling

The Go client uses connection pooling. For high load, consider:
- Increasing Kubernetes API client QPS/burst
- Using multiple ServiceAccounts for load distribution

---

## Related Documentation

- [Enterprise Deployment Guide](ENTERPRISE_DEPLOYMENT.md) - General enterprise deployment
- [Helm Chart Values](../chart/dm-nkp-gitops-a2a-server/values.yaml) - All available options
- [Enterprise Values](../chart/dm-nkp-gitops-a2a-server/values-enterprise.yaml) - Production defaults
- [Kagent Documentation](https://kagent.dev) - MCPServer CRD details
- [A2A Protocol](A2A_PROTOCOL.md) - Agent-to-Agent protocol

---

## Quick Reference

### Install Everything

```bash
# 1. Install Kagent
kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml

# 2. Deploy with enterprise values
helm upgrade --install gitops-agent ./chart/dm-nkp-gitops-a2a-server \
  -f ./chart/dm-nkp-gitops-a2a-server/values-enterprise.yaml \
  -f nkp-values.yaml \
  --namespace gitops-agent \
  --create-namespace \
  --wait

# 3. Verify
kubectl get mcpserver -n gitops-agent
kubectl get pods -n gitops-agent
```

### Uninstall

```bash
helm uninstall gitops-agent -n gitops-agent
# Kagent operator remains (cluster-wide)
```

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/issues
- Documentation: See `docs/` directory
- Kagent Support: https://kagent.dev
