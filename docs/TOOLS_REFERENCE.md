# MCP Tools Reference

Complete reference for all dm-nkp-gitops MCP server tools with test queries and examples.

## Verification Status

Tested against: `kind-mcp-test` cluster

| Tool | Status | Notes |
|------|:------:|-------|
| `list_contexts` | ✅ | Works |
| `get_current_context` | ✅ | Works |
| `get_gitops_status` | ✅ | Works |
| `list_kustomizations` | ✅ | Works |
| `get_kustomization` | ✅ | Works |
| `list_gitrepositories` | ✅ | Works |
| `get_helmreleases` | ✅ | Works |
| `get_cluster_status` | ⚠️ | Requires CAPI |
| `list_machines` | ⚠️ | Requires CAPI |
| `get_app_deployments` | ⚠️ | Requires Kommander |
| `debug_reconciliation` | ✅ | Works |
| `get_events` | ✅ | Works |
| `get_pod_logs` | ✅ | Works (needs exact pod name) |
| `check_policy_violations` | ⚠️ | Requires Gatekeeper/Kyverno |
| `list_constraints` | ⚠️ | Requires Gatekeeper |

---

## Context Tools

### list_contexts

List all available Kubernetes contexts from kubeconfig.

**Parameters:** None

**Test Queries:**
```
"List all Kubernetes contexts"
"Show me available k8s contexts"
"What clusters do I have configured?"
```

**Example Output:**
```markdown
# Available Kubernetes Contexts

Current context: **kind-mcp-test**

| Context | Current |
|---------|:-------:|
| kind-mcp-test | ✓ |
| dm-nkp-mgmt-1 |  |
```

---

### get_current_context

Get the currently active Kubernetes context.

**Parameters:** None

**Test Queries:**
```
"What's the current Kubernetes context?"
"Which cluster am I connected to?"
"Show current context"
```

**Example Output:**
```markdown
# Current Kubernetes Context

**Context:** kind-mcp-test

**Server:** https://127.0.0.1:56539
```

---

## GitOps Tools

### get_gitops_status

Get overall GitOps health summary including all Flux Kustomizations and GitRepositories.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | No | Namespace to filter (default: all namespaces) |

**Test Queries:**
```
"What's the GitOps status?"
"Show me GitOps health summary"
"Are there any failing Flux reconciliations?"
"Get GitOps status for flux-system namespace"
```

**Example Output:**
```markdown
# GitOps Status Summary

## Flux Kustomizations

- ✅ Ready: 1
- ❌ Failed: 0
- ⏸️ Suspended: 1
- 📊 Total: 2

## Flux GitRepositories

- ✅ Ready: 2
- ❌ Failed: 0
- ⏸️ Suspended: 0
- 📊 Total: 2
```

---

### list_kustomizations

List all Flux Kustomizations with their reconciliation status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | No | Namespace to filter |
| `status_filter` | string | No | Filter: `all`, `ready`, `failed`, `suspended` |

**Test Queries:**
```
"List all Kustomizations"
"Show me failing Kustomizations"
"List suspended Kustomizations in flux-system"
"What Kustomizations are ready?"
```

**Example Output:**
```markdown
# Flux Kustomizations

| Namespace | Name | Ready | Suspended | Last Applied | Message |
|-----------|------|:-----:|:---------:|--------------|--------|
| flux-system | infrastructure | ❌ | ⏸️ | - |  |
| flux-system | podinfo | ✅ |  | master@sha1:b6b680fe... | Applied revision... |

**Total:** 2 Kustomizations
```

---

### get_kustomization

Get detailed information about a specific Flux Kustomization.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `name` | string | **Yes** | Name of the Kustomization |
| `namespace` | string | **Yes** | Namespace of the Kustomization |

**Test Queries:**
```
"Get details for podinfo Kustomization in flux-system"
"Show me the podinfo Kustomization"
"What's the status of flux-system/infrastructure Kustomization?"
```

**Example Output:**
```markdown
# Kustomization: flux-system/podinfo

## Status

**Status:** ✅ Ready

## Spec

- **Path:** `./kustomize`
- **Interval:** 10m
- **Source:** GitRepository/podinfo
- **Prune:** true

## Conditions

| Type | Status | Reason | Message |
|------|:------:|--------|--------|
| Ready | True | ReconciliationSucceeded | Applied revision: master@sha1:b6b680fe... |
```

---

### list_gitrepositories

List all Flux GitRepository sources with their sync status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | No | Namespace to filter |

**Test Queries:**
```
"List all GitRepositories"
"Show me Flux sources"
"What GitRepos are configured?"
"List GitRepositories in flux-system"
```

**Example Output:**
```markdown
# Flux GitRepositories

| Namespace | Name | URL | Branch | Ready | Last Fetched |
|-----------|------|-----|--------|:-----:|--------------|
| flux-system | flux-monitoring | https://github.com/fluxcd/flux2-monit... | main | ✅ | - |
| flux-system | podinfo | https://github.com/stefanprodan/podinfo | master | ✅ | - |

**Total:** 2 GitRepositories
```

---

### get_helmreleases

List Flux HelmReleases with their status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | No | Namespace to filter |
| `status_filter` | string | No | Filter: `all`, `ready`, `failed`, `suspended` |

**Test Queries:**
```
"List all HelmReleases"
"Show me Helm deployments"
"What HelmReleases are failing?"
"List ready HelmReleases"
```

**Example Output:**
```markdown
# Flux HelmReleases

| Namespace | Name | Chart | Version | Ready | Message |
|-----------|------|-------|---------|:-----:|--------|
| flux-system | podinfo-helm | podinfo | >=6.0.0 | ✅ | Helm install succeeded... |

**Total:** 1 HelmReleases
```

---

## Cluster Tools (Requires CAPI)

### get_cluster_status

Get status of CAPI (Cluster API) clusters.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `cluster_name` | string | No | Name of specific cluster |
| `namespace` | string | No | Namespace to filter |

**Test Queries:**
```
"What's the status of CAPI clusters?"
"Show me cluster dm-nkp-workload-1 status"
"List all managed clusters"
"Get cluster health"
```

> ⚠️ **Note:** Requires Cluster API to be installed. Returns error on clusters without CAPI.

---

### list_machines

List CAPI Machines for a cluster showing node status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `cluster_name` | string | No | Filter machines for specific cluster |
| `namespace` | string | No | Namespace to filter |

**Test Queries:**
```
"List machines for dm-nkp-workload-1"
"Show me all CAPI machines"
"What nodes are in the workload cluster?"
```

> ⚠️ **Note:** Requires Cluster API to be installed.

---

## Application Tools (Requires Kommander)

### get_app_deployments

Get application deployment status across workspaces.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `app_name` | string | No | Application name to filter |
| `workspace` | string | No | Workspace name to filter |

**Test Queries:**
```
"Show app deployments"
"What apps are deployed in dm-dev-workspace?"
"Get deployment status for traefik app"
"List all ClusterApps"
```

> ⚠️ **Note:** Requires Kommander/NKP to be installed.

---

## Debug Tools

### debug_reconciliation

Debug a failing Flux reconciliation with conditions, events, and related resource status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `resource_type` | string | **Yes** | Type: `kustomization`, `gitrepository`, `helmrelease` |
| `name` | string | **Yes** | Name of the resource |
| `namespace` | string | **Yes** | Namespace of the resource |

**Test Queries:**
```
"Debug the podinfo Kustomization in flux-system"
"Why is the infrastructure Kustomization failing?"
"Debug HelmRelease podinfo-helm in flux-system"
"Investigate reconciliation failure for gitrepository/podinfo"
```

**Example Output:**
```markdown
# Debug: Kustomization flux-system/podinfo

## Status Summary

✅ **Status:** Ready

## Conditions

| Type | Status | Reason | Last Transition | Message |
|------|:------:|--------|-----------------|--------|
| Ready | ✅ | ReconciliationSucceeded | 2026-01-19 22:56:13 | Applied revision... |

## Source Reference

- **Kind:** GitRepository
- **Name:** flux-system/podinfo

✅ Source is ready

## Recent Events

| Type | Reason | Age | Message |
|------|--------|-----|--------|
| ℹ️ | GitOperationSucceeded | 3m | no changes since last reconcilation... |
| ℹ️ | ReconciliationSucceeded | 9m | Reconciliation finished in 796ms... |
```

---

### get_events

Get Kubernetes events for debugging.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | **Yes** | Namespace to get events from |
| `resource_name` | string | No | Filter for specific resource |
| `event_type` | string | No | Filter: `all`, `Normal`, `Warning` |
| `limit` | string | No | Max events to return (default: 20) |

**Test Queries:**
```
"Show events in flux-system namespace"
"Get warning events in default namespace"
"Show last 10 events for podinfo"
"What's happening in kommander namespace?"
```

**Example Output:**
```markdown
# Events in flux-system

| Type | Object | Reason | Age | Count | Message |
|------|--------|--------|-----|:-----:|--------|
| ℹ️ | GitRepository/flux-monitoring | GitOperationSucceeded | 12s | 3 | no changes... |
| ℹ️ | HelmChart/flux-system-podinfo | ArtifactUpToDate | 2m | 28 | artifact up-to-date... |
| ℹ️ | Kustomization/podinfo | ReconciliationSucceeded | 8m | 1 | Reconciliation finished... |

**Showing:** 5 events
```

---

### get_pod_logs

Get logs from a pod for debugging.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `pod_name` | string | **Yes** | Exact name of the pod |
| `namespace` | string | **Yes** | Namespace of the pod |
| `container` | string | No | Container name (uses first if not specified) |
| `tail_lines` | string | No | Lines from end (default: 100) |

**Test Queries:**
```
"Get logs from source-controller-xxx pod in flux-system"
"Show last 50 lines of logs from kustomize-controller pod"
"Get logs for helm-controller in flux-system"
```

> **Note:** Requires exact pod name. Use kubectl or kubernetes MCP server to get pod names first.

---

## Policy Tools (Requires Gatekeeper/Kyverno)

### check_policy_violations

Check for Gatekeeper or Kyverno policy violations.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `namespace` | string | No | Namespace to filter |
| `policy_engine` | string | No | Engine: `gatekeeper`, `kyverno`, `both` |

**Test Queries:**
```
"Check for policy violations"
"Are there any Gatekeeper violations?"
"Show Kyverno policy violations in default namespace"
"Check policy compliance"
```

> ⚠️ **Note:** Requires Gatekeeper or Kyverno to be installed.

---

### list_constraints

List Gatekeeper constraints and their enforcement status.

**Parameters:**
| Name | Type | Required | Description |
|------|------|:--------:|-------------|
| `constraint_kind` | string | No | Filter by constraint kind |

**Test Queries:**
```
"List all Gatekeeper constraints"
"Show K8sRequiredLabels constraints"
"What policy constraints are enforced?"
```

> ⚠️ **Note:** Requires Gatekeeper to be installed.

---

## Common Debugging Workflows

### 1. Investigating GitOps Failures

```
Step 1: "What's the GitOps status?"
Step 2: "List failing Kustomizations"
Step 3: "Debug the <name> Kustomization in <namespace>"
Step 4: "Show events in <namespace>"
```

### 2. Checking Cluster Health

```
Step 1: "Get cluster status"
Step 2: "List machines for <cluster-name>"
Step 3: "Show warning events in <namespace>"
```

### 3. Application Deployment Issues

```
Step 1: "Get app deployments"
Step 2: "List failing HelmReleases"
Step 3: "Debug HelmRelease <name> in <namespace>"
Step 4: "Get pod logs for <controller-pod>"
```

### 4. Policy Compliance Check

```
Step 1: "Check for policy violations"
Step 2: "List Gatekeeper constraints"
Step 3: "Show events for constraint violations"
```

---

## Configuration Reference

### Cursor MCP Configuration

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "/path/to/kubeconfig"
      }
    },
    "kubernetes": {
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-server-kubernetes"],
      "env": {
        "KUBECONFIG": "/path/to/kubeconfig"
      }
    }
  }
}
```

### Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "/path/to/kubeconfig"
      }
    }
  }
}
```

---

## Troubleshooting

### "Resource not found" errors

These indicate the CRD is not installed:
- **CAPI errors**: Cluster API not installed (expected on kind/minikube)
- **Kommander errors**: NKP/Kommander not installed
- **Gatekeeper/Kyverno errors**: Policy engines not installed

### Pod logs "not found"

The `get_pod_logs` tool requires the **exact** pod name including the random suffix. Use the kubernetes MCP server's `kubectl get pods` to find exact names.

### MCP server not responding

1. Check the binary path is correct
2. Verify KUBECONFIG path exists
3. Restart Cursor/Claude Desktop after config changes
4. Check stderr logs: `./bin/dm-nkp-gitops-mcp-server serve 2>&1`
