# Kubernetes Security Assessment Runbook for AI Agents

This runbook provides structured procedures for assessing Kubernetes cluster security posture and identifying potential vulnerabilities. It's designed to be both human-readable and AI-agent friendly, with clear decision trees, step-by-step procedures, and working code examples.

## Design Principles

This runbook follows the same design principles as the troubleshooting runbook:
- **Hierarchical Structure** - Problem → Symptoms → Root Cause → Solution
- **Machine-Readable Patterns** - Structured data, specific tool names, exact commands
- **Context-Aware Guidance** - Start broad, narrow down
- **Actionable Examples** - Real queries with MCP tools

---

## Quick Security Checklist

Use this checklist to quickly assess cluster security:

```
1. Policy compliance
   → "Check for policy violations"
   → Tool: check_policy_violations

2. Policy enforcement
   → "List all Gatekeeper constraints"
   → Tool: list_constraints

3. GitOps security
   → "What's the GitOps status?"
   → Tool: get_gitops_status

4. Cluster health
   → "Get cluster status"
   → Tool: get_cluster_status

5. Application security
   → "Get app deployments"
   → Tool: get_app_deployments
```

---

## Scenario 1: Policy Violations and Compliance Issues

### Problem Statement
Resources in the cluster violate security policies enforced by Gatekeeper or Kyverno.

### Symptoms
- Policy violations detected
- Resources failing admission
- Non-compliant configurations deployed
- Security standards not being enforced

### Diagnostic Workflow

#### Step 1: Check for Policy Violations
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

**Expected Output Analysis:**
- ✅ No violations → Cluster is compliant
- ❌ Violations found → Proceed to Step 2
- ⚠️ Policy engine not installed → Skip to Scenario 6

**Next Steps:**
- If violations found → Step 2
- If no violations → Check other security areas

#### Step 2: Analyze Violation Details
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

**Expected Output:**
- List of all constraint types
- Enforcement status
- Constraint names

#### Step 3: Check Events for Violations
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
    "limit": "100"
  }
}
```

**Expected Output:**
- Events showing policy violations
- Denied admission attempts
- Violation messages

### Root Cause Analysis

| Violation Type | Likely Root Cause | Risk Level | Solution |
|----------------|------------------|------------|----------|
| Missing required labels | Resource deployed without required labels | **Medium** | Add required labels to resource |
| Prohibited image registries | Using untrusted image sources | **High** | Use approved registries only |
| Missing resource limits | Containers without CPU/memory limits | **Medium** | Add resource requests/limits |
| Privileged containers | Containers running with elevated privileges | **Critical** | Remove privileged flag |
| Host network/volume access | Containers accessing host resources | **High** | Use proper isolation |
| Missing security context | No securityContext defined | **Medium** | Add securityContext with restrictions |

### Solution Steps

#### For Label Violations:
1. Identify required labels from constraint
2. Update resource manifests to include labels
3. Reapply resource (via GitOps if possible)

**Example:**
```yaml
metadata:
  labels:
    app.kubernetes.io/name: my-app
    app.kubernetes.io/part-of: platform
```

#### For Image Registry Violations:
1. Verify approved registry list
2. Update image references to use approved registries
3. Ensure image scanning is enabled

#### For Resource Limit Violations:
1. Add resource requests and limits
2. Ensure requests don't exceed cluster capacity
3. Monitor for resource pressure

**Example:**
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

#### For Privileged Container Violations:
1. Identify containers running as privileged
2. Determine if privilege is necessary
3. Use more specific capabilities instead of privileged
4. Apply least-privilege principle

### Complete Example Workflow

```
User: "Is my cluster secure from policy violations?"

Agent Flow:
1. check_policy_violations(namespace: "", policy_engine: "both")
   → Found 5 violations across 3 namespaces

2. list_constraints()
   → Shows constraints: K8sRequiredLabels, K8sAllowedRegistries, 
      K8sRequiredLimits

3. check_policy_violations(namespace: "production", policy_engine: "both")
   → 3 violations in production namespace

4. get_events(namespace: "production", event_type: "Warning")
   → Shows denied admission events for resources missing labels

Solution: 
- Add required labels to all resources
- Update HelmRelease/Kustomization to include labels
- Ensure GitOps sync applies changes
```

---

## Scenario 2: Inadequate Policy Enforcement

### Problem Statement
Policy enforcement is not configured or constraints are not properly enforced.

### Symptoms
- No policy violations found, but security issues exist
- Constraints exist but violations are allowed
- Policy engine installed but not active

### Diagnostic Workflow

#### Step 1: Check Policy Engine Status
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

**Expected Output Analysis:**
- ✅ Constraints listed → Policy engine is installed
- ❌ Error: "resource not found" → Policy engine not installed
- ⚠️ Empty list → Constraints not configured

#### Step 2: Verify Policy Enforcement
**Agent Query:**
```
"Check for policy violations"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "policy_engine": "both"
  }
}
```

**Expected Output:**
- Shows enforcement status
- Lists active policies

#### Step 3: Check GitOps for Policy Configuration
**Agent Query:**
```
"What's the GitOps status?"
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

**Expected Output:**
- Status of Kustomizations managing policies
- Any failures in policy deployment

### Root Cause Analysis

| Symptom | Likely Root Cause | Risk Level | Solution |
|---------|------------------|------------|----------|
| No constraints found | Policy engine not installed | **High** | Install Gatekeeper/Kyverno |
| Constraints exist but violations allowed | Enforcement mode set to "warn" | **Medium** | Change to "enforce" mode |
| Policies not syncing | GitOps reconciliation failing | **Medium** | Fix GitOps sync issues |
| Constraints suspended | Manual suspension | **High** | Review and re-enable if safe |

### Solution Steps

#### For Missing Policy Engine:
1. Install Gatekeeper or Kyverno via GitOps
2. Create Kustomization for policy engine
3. Wait for installation to complete
4. Verify constraints are created

#### For Weak Enforcement:
1. Update constraint to use "enforce" mode
2. Test in non-production first
3. Monitor for legitimate workloads affected
4. Adjust policies as needed

#### For GitOps Failures:
1. Debug Kustomization managing policies:
   ```
   "Debug the <policy-kustomization> Kustomization in <namespace>"
   ```
2. Fix underlying issues (see GitOps troubleshooting runbook)
3. Ensure policy changes are committed to Git
4. Wait for reconciliation

### Complete Example Workflow

```
User: "Are my security policies being enforced?"

Agent Flow:
1. list_constraints()
   → Error: "constraints.gatekeeper.sh not found"
   → Policy engine not installed

2. get_gitops_status(namespace: "flux-system")
   → Shows policy-engine Kustomization is suspended

3. list_kustomizations(namespace: "flux-system", status_filter: "suspended")
   → policy-engine Kustomization is suspended

Solution: Re-enable policy-engine Kustomization to install Gatekeeper
```

---

## Scenario 3: Misconfigured RBAC and Access Control

### Problem Statement
Excessive permissions or overly permissive service accounts create security risks.

### Symptoms
- Service accounts with cluster-admin privileges
- RoleBindings granting excessive permissions
- Missing least-privilege principle
- Unused or orphaned service accounts

### Diagnostic Workflow

**Note:** RBAC checks require additional tools beyond current MCP server. This workflow shows what should be checked:

#### Step 1: Check Cluster Health
**Agent Query:**
```
"Get cluster status"
```

**Tool Call:**
```json
{
  "tool": "get_cluster_status",
  "arguments": {}
}
```

**Expected Output:**
- Cluster phase and health
- Can indicate if cluster is accessible

#### Step 2: Check GitOps for RBAC Resources
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

**Expected Output:**
- Status of Kustomizations managing RBAC
- Any failures in RBAC resource deployment

#### Step 3: Review Policy Violations Related to RBAC
**Agent Query:**
```
"Check for policy violations"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "policy_engine": "both"
  }
}
```

**Expected Output:**
- May show violations related to RBAC policies (if configured)

### Root Cause Analysis

| Issue | Likely Root Cause | Risk Level | Solution |
|-------|------------------|------------|----------|
| Cluster-admin everywhere | Default or overly permissive bindings | **Critical** | Use least-privilege roles |
| Wildcard permissions | "*" in verbs or resources | **High** | Specify exact permissions needed |
| System service accounts with cluster-admin | Misconfigured system bindings | **Critical** | Remove unnecessary bindings |
| Missing namespace restrictions | ClusterRoleBindings instead of RoleBindings | **Medium** | Use RoleBinding when namespace-scoped is sufficient |

### Solution Steps

#### For Excessive Permissions:
1. Identify resources with excessive permissions (requires kubectl or additional tools)
2. Create minimal-required roles
3. Update bindings to use new roles
4. Test to ensure functionality maintained

#### For RBAC Policy Violations:
1. Use Gatekeeper/Kyverno policies to enforce RBAC standards
2. Create policies requiring:
   - No cluster-admin except for system components
   - Explicit resource/verb lists (no wildcards)
   - Namespace-scoped when possible
3. Apply via GitOps

### Recommended RBAC Policies

Create Gatekeeper/Kyverno policies to enforce:
- No service accounts with cluster-admin
- RoleBindings preferred over ClusterRoleBindings
- Explicit verbs (no wildcards)
- Resource limits in roles

---

## Scenario 4: Network Security Vulnerabilities

### Problem Statement
Network policies not properly configured, allowing unrestricted pod-to-pod communication.

### Symptoms
- No network policies defined
- Overly permissive network policies (allow-all)
- Missing ingress/egress controls
- Unrestricted external access

### Diagnostic Workflow

#### Step 1: Check Application Deployments
**Agent Query:**
```
"Get app deployments in <workspace>"
```

**Tool Call:**
```json
{
  "tool": "get_app_deployments",
  "arguments": {
    "workspace": "<workspace>"
  }
}
```

**Expected Output:**
- List of deployed applications
- Can indicate what should be network-protected

#### Step 2: Check Policy Violations
**Agent Query:**
```
"Check for policy violations"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "policy_engine": "both"
  }
}
```

**Expected Output:**
- May show violations related to network policies (if configured)

#### Step 3: Check GitOps Status
**Agent Query:**
```
"What's the GitOps status?"
```

**Tool Call:**
```json
{
  "tool": "get_gitops_status",
  "arguments": {}
}
```

**Expected Output:**
- Status of network policy deployments

### Root Cause Analysis

| Issue | Likely Root Cause | Risk Level | Solution |
|-------|------------------|------------|----------|
| No network policies | Network policies not implemented | **High** | Implement network policies |
| Allow-all policies | Overly permissive ingress/egress | **High** | Use deny-all, allow-specific |
| Missing namespace isolation | Policies don't isolate namespaces | **Medium** | Add namespace-level policies |
| Unrestricted egress | All pods can reach internet | **Medium** | Restrict egress to approved destinations |

### Solution Steps

#### For Missing Network Policies:
1. Create default deny-all network policy
2. Add allow rules for specific communication patterns
3. Implement namespace isolation
4. Deploy via GitOps

**Example Default Deny Policy:**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

#### For Overly Permissive Policies:
1. Review existing network policies
2. Replace allow-all with specific allow rules
3. Test communication between pods
4. Document required communication patterns

---

## Scenario 5: Container and Image Security Issues

### Problem Statement
Vulnerable container images, misconfigured security contexts, or insecure container settings.

### Symptoms
- Images from untrusted registries
- Containers running as root
- Missing security contexts
- Vulnerable image versions

### Diagnostic Workflow

#### Step 1: Check Policy Violations for Image Policies
**Agent Query:**
```
"Check for policy violations related to images"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "policy_engine": "both"
  }
}
```

**Expected Output:**
- Violations related to image registries
- Violations for privileged containers
- Missing security contexts

#### Step 2: Check Application Deployments
**Agent Query:**
```
"Get app deployments"
```

**Tool Call:**
```json
{
  "tool": "get_app_deployments",
  "arguments": {}
}
```

**Expected Output:**
- List of applications that may have container security issues

#### Step 3: Review HelmReleases for Image Configurations
**Agent Query:**
```
"List all HelmReleases"
```

**Tool Call:**
```json
{
  "tool": "get_helmreleases",
  "arguments": {}
}
```

**Expected Output:**
- HelmReleases that may specify images
- Can indicate image sources

### Root Cause Analysis

| Issue | Likely Root Cause | Risk Level | Solution |
|-------|------------------|------------|----------|
| Untrusted registries | Images from public registries | **High** | Use approved internal registries |
| Root containers | Containers running as root user | **High** | Use non-root user (runAsNonRoot) |
| Missing image scanning | No vulnerability scanning | **Medium** | Implement image scanning in CI/CD |
| Privileged containers | Containers with elevated privileges | **Critical** | Remove privileged flag, use specific capabilities |
| Missing security contexts | No securityContext defined | **Medium** | Add securityContext with restrictions |

### Solution Steps

#### For Image Registry Violations:
1. Identify approved registries from policy
2. Update HelmRelease/Kustomization to use approved registries
3. Ensure image scanning is enabled
4. Sync via GitOps

#### For Root Containers:
1. Update deployments to use non-root user
2. Add securityContext:
   ```yaml
   securityContext:
     runAsNonRoot: true
     runAsUser: 1000
   ```
3. Test application functionality
4. Deploy via GitOps

#### For Privileged Containers:
1. Identify why privilege is needed
2. Replace with specific capabilities
3. Use Pod Security Standards (restricted)
4. Apply via GitOps

---

## Scenario 6: Missing Security Tools and Monitoring

### Problem Statement
Security tools (policy engines, scanning, monitoring) are not installed or not functioning.

### Symptoms
- Policy violations check returns "not found" errors
- No security policies enforced
- Missing security tooling
- No security event monitoring

### Diagnostic Workflow

#### Step 1: Check Policy Engine Status
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

**Expected Output Analysis:**
- ✅ Lists constraints → Policy engine installed
- ❌ Error: resource not found → Policy engine not installed

#### Step 2: Try Policy Violation Check
**Agent Query:**
```
"Check for policy violations"
```

**Tool Call:**
```json
{
  "tool": "check_policy_violations",
  "arguments": {
    "policy_engine": "both"
  }
}
```

**Expected Output:**
- Shows if policy engines are available
- Indicates which engines are missing

#### Step 3: Check GitOps for Security Tools
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

**Expected Output:**
- Status of security tool deployments
- Any failures in security tool installation

### Root Cause Analysis

| Issue | Likely Root Cause | Risk Level | Solution |
|-------|------------------|------------|----------|
| Gatekeeper not installed | Security tooling not deployed | **High** | Install via GitOps |
| Kyverno not installed | Alternative policy engine missing | **Medium** | Install if using Kyverno |
| Security tools failing | Installation or configuration issues | **High** | Debug and fix tool deployments |
| No security policies defined | Tools installed but no policies | **Medium** | Deploy security policies |

### Solution Steps

#### For Missing Policy Engine:
1. Create Kustomization for policy engine (Gatekeeper/Kyverno)
2. Add to GitOps repository
3. Ensure proper RBAC for policy engine
4. Wait for reconciliation and verify installation

#### For Failing Tools:
1. Debug the Kustomization managing security tools:
   ```
   "Debug the <tool-kustomization> Kustomization in <namespace>"
   ```
2. Check events for errors:
   ```
   "Show warning events in <namespace>"
   ```
3. Review tool logs (requires kubectl)
4. Fix configuration and re-sync

---

## Complete Security Assessment Workflow

### Recommended Assessment Order

```
1. Check policy engine status
   → "List all Gatekeeper constraints"
   
2. Check for policy violations
   → "Check for policy violations" (both engines)
   
3. Review enforcement status
   → Analyze constraint enforcement modes
   
4. Check GitOps security
   → "What's the GitOps status?"
   → Verify security-related Kustomizations are healthy
   
5. Review application deployments
   → "Get app deployments"
   → Check for security misconfigurations
   
6. Check cluster health
   → "Get cluster status"
   → Ensure cluster is in good state
```

### Security Checklist

Use this comprehensive checklist:

- [ ] **Policy Engine Installed**
  - Gatekeeper or Kyverno installed and running
  - Verified via `list_constraints()`

- [ ] **No Policy Violations**
  - Zero violations across all namespaces
  - Verified via `check_policy_violations()`

- [ ] **Policies Enforced**
  - Constraints set to "enforce" mode (not "warn")
  - No critical constraints suspended

- [ ] **GitOps Secure**
  - All security-related Kustomizations healthy
  - No failed reconciliations

- [ ] **RBAC Minimal**
  - No excessive permissions
  - Least-privilege principle followed

- [ ] **Network Policies**
  - Network policies implemented
  - Namespace isolation configured

- [ ] **Container Security**
  - No privileged containers
  - Non-root containers
  - Approved image registries only

- [ ] **Monitoring Active**
  - Security events monitored
  - Policy violations alerted

---

## Advanced Security Patterns

### Pattern 1: Defense in Depth

Layer multiple security controls:
1. **Policy Engine** - Prevents bad configurations
2. **Network Policies** - Restricts communication
3. **RBAC** - Limits access
4. **Pod Security Standards** - Enforces container security

### Pattern 2: Shift Left Security

Catch issues early:
1. **GitOps** - Security policies in Git
2. **Admission Control** - Block at deployment time
3. **Policy Enforcement** - Continuous compliance
4. **Monitoring** - Detect runtime issues

### Pattern 3: Zero Trust

Assume breach, verify everything:
1. No implicit trust
2. Explicit allow lists
3. Continuous verification
4. Minimal permissions

---

## Related Documentation

- [K8S Troubleshooting Runbook](K8S_TROUBLESHOOTING_RUNBOOK.md) - General troubleshooting
- [Security Documentation](SECURITY.md) - MCP server security analysis
- [Runbook Best Practices](RUNBOOK_BEST_PRACTICES.md) - How to create runbooks
- [Tools Reference](TOOLS_REFERENCE.md) - Complete tool documentation

---

## Contributing

To add new security scenarios:
1. Follow the structure: Problem → Symptoms → Workflow → Root Cause → Solution
2. Include working examples with exact tool calls
3. Provide risk assessments
4. Test all examples before committing
