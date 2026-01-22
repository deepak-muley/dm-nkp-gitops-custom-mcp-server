# Kubernetes Troubleshooting Runbook for AI Agents

This runbook provides structured troubleshooting procedures for common Kubernetes cluster issues. It's designed to be both human-readable and AI-agent friendly, with clear decision trees, step-by-step procedures, and working code examples.

## Design Principles for Agent-Friendly Runbooks

### 1. **Hierarchical Structure**
- **Problem → Symptoms → Root Cause → Solution**
- Clear headings with consistent formatting
- Numbered steps for sequential actions

### 2. **Machine-Readable Patterns**
- Use structured data formats (tables, lists, code blocks)
- Include specific tool names and parameters
- Provide exact commands/queries that agents can execute

### 3. **Context-Aware Guidance**
- Start with quick diagnostics
- Escalate to detailed investigation
- Include decision points (if X, then Y)

### 4. **Actionable Examples**
- Real queries that work with MCP tools
- Expected outputs
- Common variations

---

## Quick Diagnostic Checklist

Use this checklist to quickly identify the problem category:

```
1. Cluster connectivity
   → "What's the current Kubernetes context?"
   → Tool: get_current_context

2. GitOps health
   → "What's the GitOps status?"
   → Tool: get_gitops_status

3. Node/cluster health
   → "Get cluster status"
   → Tool: get_cluster_status

4. Application deployments
   → "Get app deployments"
   → Tool: get_app_deployments

5. Policy violations
   → "Check for policy violations"
   → Tool: check_policy_violations
```

---

## Scenario 1: GitOps Reconciliation Failures

### Problem Statement
A Flux Kustomization or HelmRelease is failing to reconcile.

### Symptoms
- Kustomization shows `Ready: False`
- Events show reconciliation errors
- Resources not deployed to cluster

### Diagnostic Workflow

#### Step 1: Get Overall Status
**Agent Query:**
```
"What's the GitOps status in <namespace>?"
```

**Tool Call:**
```json
{
  "tool": "get_gitops_status",
  "arguments": {
    "namespace": "<namespace>"
  }
}
```

**Expected Output Analysis:**
- ✅ All ready → No problem
- ❌ Failed count > 0 → Proceed to Step 2
- ⏸️ Suspended → May need manual intervention

#### Step 2: List Failing Resources
**Agent Query:**
```
"List failing Kustomizations in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "list_kustomizations",
  "arguments": {
    "namespace": "<namespace>",
    "status_filter": "failed"
  }
}
```

#### Step 3: Debug Specific Resource
**Agent Query:**
```
"Debug the <kustomization-name> Kustomization in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "debug_reconciliation",
  "arguments": {
    "resource_type": "kustomization",
    "name": "<kustomization-name>",
    "namespace": "<namespace>"
  }
}
```

#### Step 4: Check Events
**Agent Query:**
```
"Show warning events in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "get_events",
  "arguments": {
    "namespace": "<namespace>",
    "event_type": "Warning",
    "limit": "50"
  }
}
```

### Root Cause Analysis

| Symptom in Debug Output | Likely Root Cause | Solution |
|------------------------|-------------------|----------|
| "Source not found" | GitRepository missing/failing | Check GitRepository status |
| "Dependency not ready" | Dependent Kustomization failing | Fix dependency first |
| "Validation error" | Invalid YAML/manifests | Validate manifests locally |
| "Secret not found" | Missing required secret | Check secret exists |
| "Network timeout" | Cannot reach git repository | Check network/firewall |
| "ImagePullBackOff" | Cannot pull container image | Check image registry access |

### Solution Steps

#### For Source Issues:
1. Check GitRepository status:
   ```
   "List GitRepositories in <namespace>"
   ```
2. Verify repository URL and credentials
3. Check network connectivity

#### For Dependency Issues:
1. Identify dependency:
   - Check `debug_reconciliation` output for dependency names
2. Recursively debug dependencies:
   ```
   "Debug the <dependency-name> Kustomization in <namespace>"
   ```
3. Fix dependencies in order (leaf to root)

#### For Validation Issues:
1. Get source repository details
2. Manually validate manifests:
   ```bash
   # If using kustomize
   kustomize build <path>
   
   # If using helm
   helm template <chart> --debug
   ```

#### For Secret Issues:
1. Check if secret exists (requires kubectl or kubernetes MCP server)
2. Verify secret is in correct namespace
3. Check sealed secrets are decrypted

### Complete Example Workflow

```
User: "Why is my infrastructure Kustomization failing?"

Agent Flow:
1. get_gitops_status(namespace: "flux-system")
   → Shows 1 failed Kustomization

2. list_kustomizations(namespace: "flux-system", status_filter: "failed")
   → infrastructure Kustomization is failing

3. debug_reconciliation(resource_type: "kustomization", 
                         name: "infrastructure", 
                         namespace: "flux-system")
   → Error: "dependency 'base-cluster-resources' not ready"

4. debug_reconciliation(resource_type: "kustomization",
                         name: "base-cluster-resources",
                         namespace: "flux-system")
   → Error: "GitRepository 'cluster-config' not found"

5. list_gitrepositories(namespace: "flux-system")
   → cluster-config GitRepository is missing

Solution: Create the missing GitRepository resource
```

---

## Scenario 2: Cluster Node Issues

### Problem Statement
Cluster nodes are unhealthy or not joining the cluster.

### Symptoms
- Nodes showing `NotReady` status
- Machines not provisioning
- Pods in `Pending` state

### Diagnostic Workflow

#### Step 1: Check Cluster Status
**Agent Query:**
```
"Get cluster status for <cluster-name>"
```

**Tool Call:**
```json
{
  "tool": "get_cluster_status",
  "arguments": {
    "cluster_name": "<cluster-name>",
    "namespace": "<namespace>"
  }
}
```

**Expected Output Fields:**
- Phase: `Provisioning`, `Running`, `Failed`
- Conditions: Various health conditions
- Infrastructure Ready: Boolean

#### Step 2: List Machines
**Agent Query:**
```
"List machines for <cluster-name>"
```

**Tool Call:**
```json
{
  "tool": "list_machines",
  "arguments": {
    "cluster_name": "<cluster-name>",
    "namespace": "<namespace>"
  }
}
```

#### Step 3: Check Events
**Agent Query:**
```
"Show warning events in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "get_events",
  "arguments": {
    "namespace": "<namespace>",
    "event_type": "Warning",
    "resource_name": "<machine-name>"
  }
}
```

### Root Cause Analysis

| Phase | Condition | Likely Cause | Solution |
|-------|-----------|--------------|----------|
| `Provisioning` | InfrastructureNotReady | Cloud provider issues | Check cloud credentials |
| `Provisioning` | ControlPlaneInitialized=False | Bootstrap failure | Check control plane logs |
| `Running` | NodeReady=False | Kubelet issues | Check node logs |
| `Failed` | BootstrapFailed | Cannot join cluster | Check network/firewall |

### Solution Steps

#### For Infrastructure Issues:
1. Check cloud provider credentials
2. Verify quotas/limits not exceeded
3. Check security groups/firewall rules

#### For Bootstrap Issues:
1. Get bootstrap logs (requires kubectl):
   ```bash
   kubectl logs -n <namespace> <bootstrap-pod>
   ```
2. Check bootstrap configuration
3. Verify network connectivity to control plane

#### For Node Ready Issues:
1. SSH to node (if possible)
2. Check kubelet status:
   ```bash
   systemctl status kubelet
   journalctl -u kubelet -n 100
   ```
3. Verify node can reach API server

### Complete Example Workflow

```
User: "Cluster dm-nkp-workload-1 is not provisioning nodes"

Agent Flow:
1. get_cluster_status(cluster_name: "dm-nkp-workload-1")
   → Phase: "Provisioning"
   → InfrastructureReady: False

2. list_machines(cluster_name: "dm-nkp-workload-1")
   → Shows 3 machines, all in "Provisioning" phase

3. get_events(namespace: "default", event_type: "Warning")
   → Event: "Failed to create infrastructure: quota exceeded"

Solution: Increase cloud provider quota or reduce machine count
```

---

## Scenario 3: Application Deployment Failures

### Problem Statement
Applications deployed via Kommander or Helm are not running correctly.

### Symptoms
- App shows `Not Ready` status
- HelmRelease failing
- Pods crashing or not starting

### Diagnostic Workflow

#### Step 1: Check App Deployments
**Agent Query:**
```
"Get app deployments in <workspace>"
```

**Tool Call:**
```json
{
  "tool": "get_app_deployments",
  "arguments": {
    "workspace": "<workspace-name>",
    "app_name": "<app-name>"
  }
}
```

#### Step 2: Check HelmReleases
**Agent Query:**
```
"List failing HelmReleases in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "get_helmreleases",
  "arguments": {
    "namespace": "<namespace>",
    "status_filter": "failed"
  }
}
```

#### Step 3: Debug HelmRelease
**Agent Query:**
```
"Debug HelmRelease <name> in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "debug_reconciliation",
  "arguments": {
    "resource_type": "helmrelease",
    "name": "<helmrelease-name>",
    "namespace": "<namespace>"
  }
}
```

#### Step 4: Check Pod Logs
**Agent Query:**
```
"Get logs from <pod-name> in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "get_pod_logs",
  "arguments": {
    "pod_name": "<pod-name>",
    "namespace": "<namespace>",
    "tail_lines": "100"
  }
}
```

**Note:** Pod name requires exact match. Use kubernetes MCP server to list pods first.

### Root Cause Analysis

| Error Message | Likely Cause | Solution |
|---------------|--------------|----------|
| "Chart not found" | HelmRepository issue | Check HelmRepository status |
| "Values validation failed" | Invalid values | Review HelmRelease values |
| "Install failed" | Resource conflicts | Check existing resources |
| "Upgrade failed" | Breaking changes | Review chart changes |
| "Dependency missing" | Required resources not ready | Check dependencies |

### Solution Steps

#### For Chart Issues:
1. Check HelmRepository:
   ```
   "List GitRepositories in <namespace>"
   ```
2. Verify chart version exists
3. Check HelmRepository credentials

#### For Resource Conflicts:
1. Check for existing resources with same name
2. Consider using different namespace or name
3. Delete conflicting resources if safe

#### For Pod Issues:
1. Get pod status (requires kubernetes MCP server)
2. Check pod events
3. Review pod logs for errors
4. Check resource limits/requests

### Complete Example Workflow

```
User: "My traefik app is not deploying in dm-dev-workspace"

Agent Flow:
1. get_app_deployments(workspace: "dm-dev-workspace", app_name: "traefik")
   → Status: "Not Ready"
   → Cluster: "dm-nkp-workload-1"

2. get_helmreleases(namespace: "kommander", status_filter: "failed")
   → traefik-helmrelease is failing

3. debug_reconciliation(resource_type: "helmrelease",
                         name: "traefik-helmrelease",
                         namespace: "kommander")
   → Error: "Chart 'traefik/traefik' version '10.20.0' not found in repository"

4. list_gitrepositories(namespace: "kommander")
   → traefik HelmRepository status: Ready

Solution: Update HelmRelease to use available chart version
```

---

## Scenario 4: Policy Violations

### Problem Statement
Resources are being blocked or denied due to policy violations.

### Symptoms
- Resources stuck in pending
- Policy violation events
- Resources not created

### Diagnostic Workflow

#### Step 1: Check for Violations
**Agent Query:**
```
"Check for policy violations in <namespace>"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "namespace": "<namespace>",
    "policy_engine": "both"
  }
}
```

#### Step 2: List Constraints
**Agent Query:**
```
"List all Gatekeeper constraints"
```

**Tool Call:**
```json
{
  "tool": "list_constraints",
  "arguments": {}
}
```

#### Step 3: Check Events
**Agent Query:**
```
"Show events in <namespace> related to policy violations"
```

**Tool Call:**
```json
{
  "tool": "get_events",
  "arguments": {
    "namespace": "<namespace>",
    "event_type": "Warning",
    "limit": "50"
  }
}
```

### Root Cause Analysis

| Policy Engine | Violation Type | Common Causes | Solution |
|---------------|----------------|---------------|----------|
| Gatekeeper | Missing labels | Resource missing required labels | Add required labels |
| Gatekeeper | Resource limits | CPU/memory exceeds limits | Adjust resource requests |
| Kyverno | Image policy | Using untrusted image | Use approved image registry |
| Kyverno | Network policy | Unallowed network access | Update network policies |

### Solution Steps

#### For Label Violations:
1. Identify required labels from constraint
2. Update resource manifest to include labels
3. Reapply resource

#### For Resource Limit Violations:
1. Check constraint limits
2. Adjust resource requests/limits in deployment
3. Verify within allowed ranges

#### For Image Policy Violations:
1. Use approved image registries
2. Update image references in manifests
3. Add exceptions if needed (after approval)

### Complete Example Workflow

```
User: "My deployment is being blocked by policies"

Agent Flow:
1. check_policy_violations(namespace: "default", policy_engine: "both")
   → Found 2 Gatekeeper violations
   → Resource: Deployment/my-app
   → Constraint: K8sRequiredLabels

2. list_constraints(constraint_kind: "K8sRequiredLabels")
   → Shows constraint requiring "app.kubernetes.io/name" label

3. get_events(namespace: "default", event_type: "Warning")
   → Event: "denied admission: missing required label"

Solution: Add required label to Deployment manifest
```

---

## Advanced Troubleshooting Patterns

### Pattern 1: Dependency Chain Resolution

When multiple resources depend on each other:

```
Workflow:
1. Start with the failing resource
2. For each dependency:
   a. Check if dependency is ready
   b. If not, recursively debug dependency
   c. Fix dependencies from leaf to root
3. Work back up the chain
```

**Example:**
```
infrastructure Kustomization
  └─ depends on: base-cluster-resources
       └─ depends on: sealed-secrets
            └─ depends on: cert-manager

Fix order: cert-manager → sealed-secrets → base-cluster-resources → infrastructure
```

### Pattern 2: Event Correlation

Correlate events with resource state:

```
1. Get resource status
2. Get recent events for resource
3. Match event timestamps with status transitions
4. Find the first Warning/Error event
5. Investigate root cause of that event
```

### Pattern 3: Cross-Namespace Investigation

When issues span namespaces:

```
1. Check GitOps status in all namespaces
2. Identify all affected resources
3. Check dependencies across namespaces
4. Verify service accounts and RBAC
```

---

## Working Code Examples

### Example 1: Automated Health Check Script

```bash
#!/bin/bash
# k8s-health-check.sh
# Automated health check using MCP tools

NAMESPACE="${1:-flux-system}"

echo "=== Kubernetes Cluster Health Check ==="
echo ""

# Check context
echo "1. Checking context..."
# Use: get_current_context

# GitOps status
echo "2. Checking GitOps status..."
# Use: get_gitops_status namespace="$NAMESPACE"

# Check for failures
echo "3. Checking for failures..."
# Use: list_kustomizations namespace="$NAMESPACE" status_filter="failed"

# Check events
echo "4. Checking recent events..."
# Use: get_events namespace="$NAMESPACE" event_type="Warning" limit="20"

echo ""
echo "=== Health Check Complete ==="
```

### Example 2: Failure Investigation Function

```python
# Python example for automated investigation
def investigate_gitops_failure(namespace: str, resource_name: str):
    """
    Automated investigation workflow for GitOps failures.
    """
    steps = [
        {
            "step": "Get overall status",
            "tool": "get_gitops_status",
            "args": {"namespace": namespace}
        },
        {
            "step": "List failing resources",
            "tool": "list_kustomizations",
            "args": {"namespace": namespace, "status_filter": "failed"}
        },
        {
            "step": "Debug specific resource",
            "tool": "debug_reconciliation",
            "args": {
                "resource_type": "kustomization",
                "name": resource_name,
                "namespace": namespace
            }
        },
        {
            "step": "Check events",
            "tool": "get_events",
            "args": {"namespace": namespace, "event_type": "Warning"}
        }
    ]
    
    results = []
    for step in steps:
        result = execute_mcp_tool(step["tool"], step["args"])
        results.append({
            "step": step["step"],
            "result": result
        })
    
    return analyze_results(results)
```

### Example 3: MCP Tool Call Template

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "debug_reconciliation",
    "arguments": {
      "resource_type": "kustomization",
      "name": "infrastructure",
      "namespace": "flux-system"
    }
  }
}
```

---

## Decision Trees

### Decision Tree: GitOps Failure

```
Start: Kustomization failing
│
├─ Error mentions "Source"
│  ├─ Check GitRepository status
│  ├─ Verify URL/credentials
│  └─ Check network connectivity
│
├─ Error mentions "Dependency"
│  ├─ Identify dependency
│  ├─ Recursively debug dependency
│  └─ Fix in dependency order
│
├─ Error mentions "Validation"
│  ├─ Validate YAML locally
│  ├─ Check syntax errors
│  └─ Verify resource references
│
└─ Error mentions "Secret"
   ├─ Verify secret exists
   ├─ Check namespace
   └─ Verify sealed secrets decrypted
```

### Decision Tree: Cluster Node Issues

```
Start: Node not ready
│
├─ Node in Provisioning phase
│  ├─ Check infrastructure readiness
│  ├─ Verify cloud credentials
│  └─ Check quotas/limits
│
├─ Node joined but NotReady
│  ├─ Check kubelet logs
│  ├─ Verify API server connectivity
│  └─ Check node resources
│
└─ Node in Failed phase
   ├─ Check bootstrap logs
   ├─ Verify network/firewall
   └─ Review infrastructure provider logs
```

---

## Best Practices Summary

### For AI Agents

1. **Start Broad, Narrow Down**
   - Begin with overall status checks
   - Progressively drill into specific resources

2. **Follow Dependency Chains**
   - Always check dependencies first
   - Work from leaf to root

3. **Correlate Multiple Signals**
   - Combine status, events, and logs
   - Look for temporal correlations

4. **Provide Context**
   - Include namespace, resource name, cluster
   - Reference previous investigation steps

5. **Suggest Next Steps**
   - Based on findings, suggest specific actions
   - Provide tool calls with correct parameters

### For Humans Learning

1. **Understand the Flow**
   - Each tool provides specific information
   - Combine tools for complete picture

2. **Practice with Examples**
   - Use the provided workflows
   - Modify to fit your scenarios

3. **Build Your Own Runbooks**
   - Document common issues you encounter
   - Add to this structure

4. **Share Learnings**
   - Contribute improvements
   - Document new patterns

---

## References

- [Tools Reference](TOOLS_REFERENCE.md) - Complete tool documentation
- [MCP Primer](MCP_PRIMER.md) - Understanding MCP protocol
- [AGENTS.md](../AGENTS.md) - AI agent integration guide

---

## Contributing

To add new troubleshooting scenarios:

1. Follow the structure: Problem → Symptoms → Workflow → Root Cause → Solution
2. Include working examples with exact tool calls
3. Provide decision trees or flowcharts
4. Test all examples before committing
