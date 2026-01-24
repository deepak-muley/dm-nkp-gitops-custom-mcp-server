# Flux MCP Server Setup Guide

This guide walks you through setting up the **Flux MCP Server** for AI-assisted GitOps operations with your NKP clusters.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Testing](#testing)
5. [Integration with AI Assistants](#integration-with-ai-assistants)
6. [Available Tools](#available-tools)
7. [Example Queries](#example-queries)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before installing Flux MCP Server, ensure you have:

### 1. Flux Operator Installed

The Flux MCP Server is part of the Flux Operator project. Your clusters should have Flux installed:

```bash
# Check if Flux is installed
flux version

# Check Flux controllers
kubectl get pods -n flux-system
```

### 2. Valid Kubeconfig Files

You need kubeconfig files for the clusters you want to monitor:

| Cluster | Kubeconfig Location |
|---------|---------------------|
| Management Cluster | `/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf` |
| Workload Cluster 1 | `/Users/deepak.muley/ws/nkp/dm-nkp-workload-1.kubeconfig` |
| Workload Cluster 2 | `/Users/deepak.muley/ws/nkp/dm-nkp-workload-2.kubeconfig` |

### 3. RBAC Permissions

Your kubeconfig should have permissions to view Flux resources:

```yaml
# Minimum required permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: flux-mcp-reader
rules:
  # Flux Kustomizations
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: ["kustomizations"]
    verbs: ["get", "list", "watch"]
  # Flux GitRepositories
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["gitrepositories", "helmrepositories", "ocirepositories"]
    verbs: ["get", "list", "watch"]
  # Flux HelmReleases
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: ["helmreleases"]
    verbs: ["get", "list", "watch"]
  # Core resources for debugging
  - apiGroups: [""]
    resources: ["pods", "events", "namespaces", "configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
  # CAPI Clusters (if using Cluster API)
  - apiGroups: ["cluster.x-k8s.io"]
    resources: ["clusters", "machines", "machinedeployments"]
    verbs: ["get", "list", "watch"]
```

---

## Installation

### Option A: Homebrew (Recommended for macOS/Linux)

```bash
# Add the tap
brew tap controlplaneio-fluxcd/tap

# Install flux-operator-mcp
brew install controlplaneio-fluxcd/tap/flux-operator-mcp

# Verify installation
flux-operator-mcp --version
```

### Option B: Download Pre-built Binary

```bash
# For macOS (Apple Silicon)
curl -LO https://github.com/controlplaneio-fluxcd/flux-operator/releases/latest/download/flux-operator-mcp_darwin_arm64.tar.gz
tar -xzf flux-operator-mcp_darwin_arm64.tar.gz
sudo mv flux-operator-mcp /usr/local/bin/

# For macOS (Intel)
curl -LO https://github.com/controlplaneio-fluxcd/flux-operator/releases/latest/download/flux-operator-mcp_darwin_amd64.tar.gz
tar -xzf flux-operator-mcp_darwin_amd64.tar.gz
sudo mv flux-operator-mcp /usr/local/bin/

# For Linux (amd64)
curl -LO https://github.com/controlplaneio-fluxcd/flux-operator/releases/latest/download/flux-operator-mcp_linux_amd64.tar.gz
tar -xzf flux-operator-mcp_linux_amd64.tar.gz
sudo mv flux-operator-mcp /usr/local/bin/

# Verify
flux-operator-mcp --version
```

### Option C: Build from Source

```bash
# Clone the repository
git clone https://github.com/controlplaneio-fluxcd/flux-operator.git
cd flux-operator

# Build the MCP server
make mcp-build

# The binary is in bin/
./bin/flux-operator-mcp --version

# Optionally move to PATH
sudo mv bin/flux-operator-mcp /usr/local/bin/
```

---

## Configuration

### Basic Configuration

Create a configuration file or use environment variables:

```bash
# Set your kubeconfig
export KUBECONFIG=/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf

# Run in read-only mode (recommended initially)
flux-operator-mcp serve --read-only=true
```

### Multi-Cluster Configuration

For monitoring multiple clusters, combine kubeconfigs:

```bash
# Method 1: Combine kubeconfigs into one file
export KUBECONFIG=/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf:/Users/deepak.muley/ws/nkp/dm-nkp-workload-1.kubeconfig:/Users/deepak.muley/ws/nkp/dm-nkp-workload-2.kubeconfig

# List available contexts
kubectl config get-contexts

# Run MCP server
flux-operator-mcp serve
```

### Command-Line Options

```bash
flux-operator-mcp serve [flags]

Flags:
  --read-only          Enable read-only mode (no mutations allowed)
  --log-level string   Log level: debug, info, warn, error (default "info")
  --kubeconfig string  Path to kubeconfig file (overrides KUBECONFIG env)
  --context string     Kubernetes context to use (default: current context)
  --help               Show help
```

---

## Testing

### Test the Server Manually

```bash
# Start the server
flux-operator-mcp serve --read-only=true &

# Send a test request (tools/list)
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | flux-operator-mcp serve

# Expected response: List of available tools
```

### Test with MCP Inspector

If you have Node.js installed:

```bash
# Install MCP Inspector
npm install -g @anthropic/mcp-inspector

# Run inspector with your server
mcp-inspector flux-operator-mcp serve
```

---

## Integration with AI Assistants

### Cursor IDE

1. Open Cursor Settings (Cmd+, or Ctrl+,)
2. Search for "MCP" or navigate to Features → MCP Servers
3. Add configuration:

```json
{
  "mcpServers": {
    "flux-operator": {
      "command": "/usr/local/bin/flux-operator-mcp",
      "args": ["serve", "--read-only=true"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
      }
    }
  }
}
```

Or create/edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "flux-operator": {
      "command": "/usr/local/bin/flux-operator-mcp",
      "args": ["serve", "--read-only=true"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
      }
    }
  }
}
```

4. Restart Cursor

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "flux-operator": {
      "command": "/usr/local/bin/flux-operator-mcp",
      "args": ["serve", "--read-only=true"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
      }
    }
  }
}
```

Restart Claude Desktop after saving.

### VS Code with GitHub Copilot

Add to `.vscode/settings.json` or User Settings:

```json
{
  "mcp": {
    "servers": {
      "flux-operator": {
        "command": "/usr/local/bin/flux-operator-mcp",
        "args": ["serve", "--read-only=true"],
        "env": {
          "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
        }
      }
    }
  },
  "chat.mcp.enabled": true
}
```

---

## Available Tools

The Flux MCP Server provides these tools:

### Cluster & Context Tools

| Tool | Description |
|------|-------------|
| `list_contexts` | List available Kubernetes contexts |
| `get_current_context` | Get the current active context |
| `switch_context` | Switch to a different context |
| `get_flux_version` | Get Flux version installed in cluster |

### GitOps Status Tools

| Tool | Description |
|------|-------------|
| `list_kustomizations` | List all Flux Kustomizations |
| `get_kustomization` | Get details of a specific Kustomization |
| `list_gitrepositories` | List all GitRepository sources |
| `get_gitrepository` | Get details of a GitRepository |
| `list_helmreleases` | List all HelmReleases |
| `get_helmrelease` | Get details of a HelmRelease |

### Debugging Tools

| Tool | Description |
|------|-------------|
| `get_reconciliation_status` | Get reconciliation status of a resource |
| `get_events` | Get Kubernetes events for a resource |
| `get_pod_logs` | Get logs from a pod |
| `trace_failure` | Trace the root cause of a failure |

### Operations Tools (if not read-only)

| Tool | Description |
|------|-------------|
| `reconcile` | Trigger reconciliation of a resource |
| `suspend` | Suspend reconciliation |
| `resume` | Resume reconciliation |

---

## Example Queries

Once configured, you can ask your AI assistant questions like:

### Status Queries

```
"What's the status of all Kustomizations in the dm-nkp-gitops-infra namespace?"

"Which GitRepositories are failing to sync?"

"Show me the reconciliation status of clusterops-clusters Kustomization"
```

### Debugging Queries

```
"Why is the clusterops-workspace-applications Kustomization failing?"

"What events are related to the dm-nkp-workload-1 cluster?"

"Trace the failure of the clusterops-sealed-secrets Kustomization"
```

### Comparison Queries

```
"Compare the HelmRelease status between dm-nkp-workload-1 and dm-nkp-workload-2"

"Which Kustomizations are healthy vs unhealthy?"

"Show me all suspended resources"
```

### Operational Queries

```
"What version of Flux is running in my cluster?"

"List all available contexts in my kubeconfig"

"Switch to the dm-nkp-workload-1 context"
```

---

## Troubleshooting

### Common Issues

#### 1. Server Not Starting

```bash
# Check if binary is executable
ls -la /usr/local/bin/flux-operator-mcp

# Make executable if needed
chmod +x /usr/local/bin/flux-operator-mcp

# Check for errors
flux-operator-mcp serve 2>&1
```

#### 2. KUBECONFIG Not Found

```bash
# Verify kubeconfig exists
ls -la /Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf

# Test kubectl works
KUBECONFIG=/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf kubectl get nodes
```

#### 3. Permission Denied

```bash
# Check your RBAC permissions
kubectl auth can-i list kustomizations.kustomize.toolkit.fluxcd.io --all-namespaces

# If denied, you need to create appropriate RBAC bindings
```

#### 4. MCP Server Not Appearing in AI Assistant

1. Verify the path to the binary is correct
2. Check the JSON syntax in configuration
3. Restart the AI assistant completely
4. Check logs (if available) for connection errors

#### 5. Tools Not Working

```bash
# Test tools manually
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | flux-operator-mcp serve

# Check for specific tool
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_kustomizations","arguments":{"namespace":"dm-nkp-gitops-infra"}}}' | flux-operator-mcp serve
```

### Debug Mode

Run with debug logging:

```bash
flux-operator-mcp serve --log-level=debug
```

### Getting Help

- [Flux Operator Documentation](https://fluxoperator.dev/docs/mcp/)
- [Flux CD Slack](https://slack.cncf.io) - #flux channel
- [GitHub Issues](https://github.com/controlplaneio-fluxcd/flux-operator/issues)

---

## Security Considerations

### Read-Only Mode

Always start with read-only mode:

```bash
flux-operator-mcp serve --read-only=true
```

This prevents accidental mutations to your cluster.

### Service Account Impersonation

For production use, consider using service account impersonation to limit permissions:

```yaml
# Create a limited service account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mcp-reader
  namespace: flux-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mcp-reader-binding
subjects:
  - kind: ServiceAccount
    name: mcp-reader
    namespace: flux-system
roleRef:
  kind: ClusterRole
  name: flux-mcp-reader
  apiGroup: rbac.authorization.k8s.io
```

### Secrets Masking

The Flux MCP Server automatically masks Kubernetes Secret values in responses. However, be cautious about:

- ConfigMaps that might contain sensitive data
- Environment variables in pod specs
- Annotations/labels with sensitive information

---

## Next Steps

1. **Install Flux MCP Server** using one of the methods above
2. **Configure your AI assistant** (Cursor, Claude Desktop, or VS Code)
3. **Test with simple queries** like "What's the Flux version?"
4. **Start debugging** your GitOps pipelines with AI assistance

For a custom MCP server tailored to your specific NKP setup, see the custom MCP server in this repository.
