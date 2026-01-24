# Security Analysis: dm-nkp-gitops-mcp-server

## Overview

This document analyzes the security posture of the MCP server and provides recommendations for hardening.

## Current Security Model

### Communication Security

| Aspect | Current State | Risk Level | Notes |
|--------|--------------|------------|-------|
| Transport | stdio (stdin/stdout) | **Low** | Local process communication, no network exposure |
| Protocol | JSON-RPC 2.0 | **Low** | Standard protocol, no custom encryption needed |
| Authentication | Implicit (process owner) | **Medium** | Relies on OS-level process security |
| Authorization | Read-only mode flag | **Medium** | Opt-in, not enforced by default |

### Why stdio is Secure

```
┌──────────────────┐     pipe (local)    ┌─────────────────┐
│  AI Assistant    │◄───────────────────►│   MCP Server    │
│  (Cursor/Claude) │    stdin/stdout     │                 │
└──────────────────┘                     └────────┬────────┘
                                                  │
                                                  ▼ kubeconfig
                                         ┌─────────────────┐
                                         │  Kubernetes API │
                                         └─────────────────┘
```

**Key Security Properties:**
1. **No network listener** - Server doesn't bind to any network port
2. **Process isolation** - Communication is confined to parent-child process
3. **Same user context** - Runs under the same user as the AI client
4. **No shared secrets** - No API keys or tokens passed between client/server

## Security Concerns & Mitigations

### 1. Kubernetes RBAC Bypass

**Risk:** Server uses the user's kubeconfig, potentially with cluster-admin privileges.

**Current Mitigation:**
- Read-only mode (`--read-only` flag)

**Recommended Mitigations:**
```yaml
# Create a restricted ServiceAccount for MCP server
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mcp-server-readonly
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mcp-server-readonly
rules:
  # Core resources (read-only)
  - apiGroups: [""]
    resources: ["pods", "pods/log", "events", "namespaces", "configmaps"]
    verbs: ["get", "list", "watch"]
  # Flux resources (read-only)
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: ["kustomizations"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["gitrepositories", "helmrepositories"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["helm.toolkit.fluxcd.io"]
    resources: ["helmreleases"]
    verbs: ["get", "list", "watch"]
  # CAPI resources (read-only)
  - apiGroups: ["cluster.x-k8s.io"]
    resources: ["clusters", "machines", "machinedeployments"]
    verbs: ["get", "list", "watch"]
  # Policy resources (read-only)
  - apiGroups: ["constraints.gatekeeper.sh"]
    resources: ["*"]
    verbs: ["get", "list"]
  - apiGroups: ["templates.gatekeeper.sh"]
    resources: ["constrainttemplates"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: mcp-server-readonly
subjects:
  - kind: ServiceAccount
    name: mcp-server-readonly
    namespace: default
roleRef:
  kind: ClusterRole
  name: mcp-server-readonly
  apiGroup: rbac.authorization.k8s.io
```

### 2. Sensitive Data Exposure

**Risk:** Pod logs and events may contain sensitive information.

**Current State:** Logs are returned as-is to the AI client.

**Recommended Mitigations:**
1. Add log filtering to redact common secret patterns
2. Implement namespace allowlist/blocklist
3. Add resource size limits to prevent large data exfiltration

```go
// Example: Add to pkg/tools/debug_handlers.go
var sensitivePatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?i)(password|secret|token|key|credential)[\s]*[=:]\s*\S+`),
    regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9\-._~+/]+=*`),
    regexp.MustCompile(`[A-Za-z0-9+/]{40,}={0,2}`), // Base64 secrets
}

func redactSensitiveData(text string) string {
    result := text
    for _, pattern := range sensitivePatterns {
        result = pattern.ReplaceAllString(result, "[REDACTED]")
    }
    return result
}
```

### 3. Input Validation

**Risk:** Malicious arguments could cause injection or DoS.

**Current State:** Basic type checking via JSON schema.

**Recommended Mitigations:**
```go
// Add validation in pkg/tools/registry.go
func validateNamespace(ns string) error {
    if ns == "" {
        return nil
    }
    if len(ns) > 253 {
        return fmt.Errorf("namespace too long")
    }
    if !regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(ns) {
        return fmt.Errorf("invalid namespace format")
    }
    return nil
}

func validateResourceName(name string) error {
    if name == "" {
        return fmt.Errorf("name is required")
    }
    if len(name) > 253 {
        return fmt.Errorf("name too long")
    }
    return nil
}
```

### 4. Resource Exhaustion

**Risk:** Unbounded queries could overload the API server.

**Current Mitigations:**
- 30-second timeout on API calls
- Event limit parameter

**Recommended Mitigations:**
```go
// Add to pkg/config/config.go
type ResourceLimits struct {
    MaxLogLines     int64 `default:"500"`
    MaxEvents       int   `default:"100"`
    MaxResources    int   `default:"500"`
    RequestTimeout  time.Duration `default:"30s"`
}
```

## Security Checklist

### For Operators

- [ ] Always use `--read-only` mode in production
- [ ] Create a dedicated ServiceAccount with minimal RBAC
- [ ] Use kubeconfig with limited cluster access
- [ ] Audit which namespaces the MCP server can access
- [ ] Review logs for sensitive data exposure
- [ ] Keep the binary updated

### For Developers

- [ ] Never log request/response bodies at INFO level
- [ ] Validate all input parameters
- [ ] Use context with timeouts for all K8s calls
- [ ] Don't expose internal error details to clients
- [ ] Sanitize user-provided data in error messages
- [ ] Add rate limiting for expensive operations

## Comparison: MCP vs API Server Security

| Feature | Traditional API Server | MCP (stdio) |
|---------|----------------------|-------------|
| Network exposure | Yes (HTTP/HTTPS) | No |
| Authentication | Required (tokens, certs) | Implicit (OS process) |
| Authorization | RBAC/ABAC | Application-level |
| TLS | Required | Not applicable |
| CORS | Needed for browsers | Not applicable |
| DoS protection | Rate limiting needed | Process limits |
| Audit logging | Built-in | Application must implement |

## Threat Model

### In Scope Threats

1. **Malicious AI prompts** - User tricks AI into executing harmful queries
   - Mitigation: Read-only mode, RBAC restrictions
   
2. **Data exfiltration** - Sensitive data leaked through AI responses
   - Mitigation: Log redaction, namespace restrictions

3. **Denial of Service** - Expensive queries overload cluster
   - Mitigation: Timeouts, limits, rate limiting

### Out of Scope Threats

1. **Compromised AI client** - If Cursor/Claude is compromised, MCP is not the boundary
2. **Compromised workstation** - If the user's machine is compromised, the attacker has full access
3. **Kubernetes API compromise** - MCP relies on K8s API security

## Future Security Enhancements

### 1. Audit Logging

```go
type AuditEvent struct {
    Timestamp   time.Time
    Tool        string
    Arguments   map[string]interface{}
    User        string // From OS
    Success     bool
    Error       string
}

func (r *Registry) auditLog(event AuditEvent) {
    r.logger.Info("AUDIT",
        "timestamp", event.Timestamp,
        "tool", event.Tool,
        "arguments", event.Arguments,
        "success", event.Success,
    )
}
```

### 2. Namespace Allowlist

```go
type SecurityConfig struct {
    AllowedNamespaces []string
    DeniedNamespaces  []string
    RedactPatterns    []string
}

func (c *SecurityConfig) IsNamespaceAllowed(ns string) bool {
    if len(c.AllowedNamespaces) > 0 {
        return contains(c.AllowedNamespaces, ns)
    }
    return !contains(c.DeniedNamespaces, ns)
}
```

### 3. Request Signing (for remote MCP)

If MCP is ever extended to support network transport:

```go
type SignedRequest struct {
    Request   JSONRPCRequest
    Timestamp time.Time
    Nonce     string
    Signature string // HMAC-SHA256
}
```

## Summary

| Security Aspect | Grade | Notes |
|----------------|-------|-------|
| Transport Security | A | stdio is inherently secure |
| Authentication | B | Relies on OS process security |
| Authorization | B- | Read-only mode is opt-in |
| Input Validation | C+ | Basic, needs enhancement |
| Data Protection | C | No redaction currently |
| Audit Logging | D | Not implemented |

**Overall Assessment:** The MCP server is reasonably secure for its intended use case (local AI assistant tool). The stdio transport eliminates network-based attack vectors. The main risks are around data exposure and need for proper RBAC configuration.

**Recommendation:** Always use `--read-only` mode and configure a restricted ServiceAccount for production use.
