# Learning kmcp with GitOps MCP Server

This guide helps you learn [kmcp](https://kagent.dev/docs/kmcp) by building the same GitOps monitoring tools in Python, while keeping your existing Go implementation.

## Prerequisites

```bash
# Install kmcp CLI
brew install kagent-dev/tap/kmcp
# or
pip install kmcp

# Install Kagent (for K8s deployment)
kubectl apply -f https://github.com/kagent-dev/kagent/releases/latest/download/install.yaml
```

## Step 1: Create a New kmcp Project

```bash
# Create learning project alongside your Go project
cd ~/go/src/github.com/deepak-muley
mkdir dm-nkp-gitops-kmcp-server
cd dm-nkp-gitops-kmcp-server

# Initialize with kmcp
kmcp init gitops-mcp-server
```

This creates:
```
gitops-mcp-server/
├── src/
│   └── gitops_mcp_server/
│       ├── __init__.py
│       └── server.py          # FastMCP server template
├── pyproject.toml
└── README.md
```

## Step 2: Implement the Same Tools in Python

### server.py - Main MCP Server

```python
# src/gitops_mcp_server/server.py
from mcp.server.fastmcp import FastMCP
from kubernetes import client, config
from typing import Optional

# Initialize FastMCP server
mcp = FastMCP("dm-nkp-gitops-mcp")

# Load kubernetes config
try:
    config.load_incluster_config()
except:
    config.load_kube_config()

# Dynamic client for CRDs
from kubernetes.dynamic import DynamicClient
k8s_client = client.ApiClient()
dyn_client = DynamicClient(k8s_client)


@mcp.tool()
def get_gitops_status(namespace: Optional[str] = None) -> str:
    """Get overall GitOps status including Flux Kustomizations and GitRepositories.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Markdown table with GitOps health summary
    """
    # Get Kustomizations
    kust_api = dyn_client.resources.get(
        api_version="kustomize.toolkit.fluxcd.io/v1",
        kind="Kustomization"
    )
    
    if namespace:
        kustomizations = kust_api.get(namespace=namespace)
    else:
        kustomizations = kust_api.get()
    
    # Count status
    ready = failed = suspended = 0
    for k in kustomizations.items:
        conditions = k.get("status", {}).get("conditions", [])
        is_ready = any(c["type"] == "Ready" and c["status"] == "True" for c in conditions)
        is_suspended = k.get("spec", {}).get("suspend", False)
        
        if is_suspended:
            suspended += 1
        elif is_ready:
            ready += 1
        else:
            failed += 1
    
    return f"""## GitOps Status

| Resource | Ready | Failed | Suspended |
|----------|-------|--------|-----------|
| Kustomizations | {ready} | {failed} | {suspended} |

Total: {len(kustomizations.items)} Kustomizations
"""


@mcp.tool()
def list_kustomizations(
    namespace: Optional[str] = None,
    status_filter: str = "all"
) -> str:
    """List all Flux Kustomizations with their reconciliation status.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
        status_filter: Filter by status: all, ready, failed, suspended
    
    Returns:
        Markdown table of Kustomizations
    """
    kust_api = dyn_client.resources.get(
        api_version="kustomize.toolkit.fluxcd.io/v1",
        kind="Kustomization"
    )
    
    if namespace:
        items = kust_api.get(namespace=namespace).items
    else:
        items = kust_api.get().items
    
    rows = []
    for k in items:
        name = k.metadata.name
        ns = k.metadata.namespace
        conditions = k.get("status", {}).get("conditions", [])
        
        # Determine status
        is_suspended = k.get("spec", {}).get("suspend", False)
        is_ready = any(c["type"] == "Ready" and c["status"] == "True" for c in conditions)
        
        if is_suspended:
            status = "Suspended"
        elif is_ready:
            status = "Ready"
        else:
            status = "Failed"
        
        # Apply filter
        if status_filter != "all" and status.lower() != status_filter.lower():
            continue
        
        rows.append(f"| {name} | {ns} | {status} |")
    
    header = "| Name | Namespace | Status |\n|------|-----------|--------|"
    return f"## Kustomizations\n\n{header}\n" + "\n".join(rows)


@mcp.tool()
def get_cluster_status(
    cluster_name: Optional[str] = None,
    namespace: Optional[str] = None
) -> str:
    """Get status of CAPI clusters.
    
    Args:
        cluster_name: Name of specific cluster (default: all clusters)
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Markdown table of cluster status
    """
    cluster_api = dyn_client.resources.get(
        api_version="cluster.x-k8s.io/v1beta1",
        kind="Cluster"
    )
    
    try:
        if namespace:
            clusters = cluster_api.get(namespace=namespace)
        else:
            clusters = cluster_api.get()
    except Exception as e:
        return f"Error: CAPI not installed or no clusters found: {e}"
    
    rows = []
    for c in clusters.items:
        if cluster_name and c.metadata.name != cluster_name:
            continue
        
        name = c.metadata.name
        ns = c.metadata.namespace
        phase = c.get("status", {}).get("phase", "Unknown")
        
        rows.append(f"| {name} | {ns} | {phase} |")
    
    if not rows:
        return "No CAPI clusters found"
    
    header = "| Name | Namespace | Phase |\n|------|-----------|-------|"
    return f"## CAPI Clusters\n\n{header}\n" + "\n".join(rows)


@mcp.tool()
def debug_reconciliation(
    resource_type: str,
    name: str,
    namespace: str
) -> str:
    """Debug a failing Flux reconciliation.
    
    Args:
        resource_type: Type of resource: kustomization, gitrepository, helmrelease
        name: Name of the resource
        namespace: Namespace of the resource
    
    Returns:
        Detailed debug information
    """
    api_versions = {
        "kustomization": ("kustomize.toolkit.fluxcd.io/v1", "Kustomization"),
        "gitrepository": ("source.toolkit.fluxcd.io/v1", "GitRepository"),
        "helmrelease": ("helm.toolkit.fluxcd.io/v2", "HelmRelease"),
    }
    
    if resource_type.lower() not in api_versions:
        return f"Unknown resource type: {resource_type}"
    
    api_version, kind = api_versions[resource_type.lower()]
    
    resource_api = dyn_client.resources.get(api_version=api_version, kind=kind)
    
    try:
        resource = resource_api.get(name=name, namespace=namespace)
    except Exception as e:
        return f"Error: {e}"
    
    # Extract conditions
    conditions = resource.get("status", {}).get("conditions", [])
    
    output = f"## Debug: {kind}/{name}\n\n"
    output += f"**Namespace:** {namespace}\n\n"
    output += "### Conditions\n\n"
    output += "| Type | Status | Reason | Message |\n"
    output += "|------|--------|--------|--------|\n"
    
    for c in conditions:
        msg = c.get("message", "")[:50] + "..." if len(c.get("message", "")) > 50 else c.get("message", "")
        output += f"| {c['type']} | {c['status']} | {c.get('reason', '')} | {msg} |\n"
    
    return output


# Entry point
if __name__ == "__main__":
    mcp.run()
```

## Step 3: Run Locally (Same as Go version)

```bash
# Install dependencies
cd dm-nkp-gitops-kmcp-server
pip install -e .

# Run the MCP server (stdio mode - same as your Go version)
python -m gitops_mcp_server.server

# Or use kmcp
kmcp run
```

## Step 4: Test with MCP Inspector

```bash
# Install MCP inspector
npx @anthropic/mcp-inspector

# Connect to your Python server
npx @anthropic/mcp-inspector python -m gitops_mcp_server.server
```

## Step 5: Deploy to Kubernetes with kmcp

### Create MCPServer CRD

```yaml
# k8s/mcpserver.yaml
apiVersion: kagent.dev/v1alpha1
kind: MCPServer
metadata:
  name: gitops-mcp-server
  namespace: gitops-agent
spec:
  image: ghcr.io/deepak-muley/dm-nkp-gitops-kmcp-server:latest
  replicas: 1
  
  # Service account for K8s API access
  serviceAccountName: gitops-mcp-server
  
  # Environment variables
  env:
    - name: LOG_LEVEL
      value: "info"
  
  # Resource limits
  resources:
    limits:
      memory: "256Mi"
      cpu: "200m"
    requests:
      memory: "64Mi"
      cpu: "50m"
```

### Deploy with kmcp

```bash
# Deploy to Kubernetes
kmcp deploy --namespace gitops-agent

# Or using kubectl
kubectl apply -f k8s/mcpserver.yaml

# Check status
kubectl get mcpservers -n gitops-agent
```

## Comparison: Go vs Python/kmcp

| Aspect | Go (Your Implementation) | Python (kmcp) |
|--------|--------------------------|---------------|
| **Server Setup** | ~100 lines custom JSON-RPC | 5 lines with FastMCP |
| **Tool Definition** | Register in registry.go | `@mcp.tool()` decorator |
| **K8s Client** | client-go, dynamic client | kubernetes-python |
| **Deployment** | Helm chart (manual) | kmcp CRD (automated) |
| **Binary Size** | ~15MB static | ~50MB+ (Python runtime) |
| **Startup Time** | <100ms | ~1-2s |
| **Learning Curve** | Higher | Lower |

## Exercise: Add a New Tool to Both

Try implementing `list_helmreleases` in both:

### Go Version (pkg/tools/app_handlers.go)
```go
func (r *Registry) handleGetHelmReleases(args map[string]interface{}) (*mcp.ToolCallResult, error) {
    // Your existing implementation
}
```

### Python Version
```python
@mcp.tool()
def list_helmreleases(namespace: Optional[str] = None) -> str:
    """List Flux HelmReleases with their status."""
    hr_api = dyn_client.resources.get(
        api_version="helm.toolkit.fluxcd.io/v2",
        kind="HelmRelease"
    )
    # ... implementation
```

## Next Steps

1. **Complete all tools** - Port remaining tools to Python
2. **Compare performance** - Benchmark both implementations
3. **Deploy both** - Run side-by-side in your cluster
4. **Try A2A** - Use Kagent for multi-agent orchestration

## Resources

- [kmcp Documentation](https://kagent.dev/docs/kmcp)
- [FastMCP Guide](https://github.com/jlowin/fastmcp)
- [Kagent Quick Start](https://kagent.dev/docs/getting-started)
- [Kagent Lab (Free)](https://kagent.dev/docs/kmcp#kagent-lab-discover-kagent-and-kmcp)
