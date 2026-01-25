# Security Improvements Applied

This document outlines the security improvements made to reduce code scanning alerts and harden the MCP server.

## Issues Fixed

### 1. Input Validation

**Issue**: No validation of user-provided arguments (namespace, resource names, etc.), allowing potential injection attacks.

**Fix**: Added comprehensive input validation:
- `validateNamespace()` - Validates Kubernetes namespace names
- `validateResourceName()` - Validates Kubernetes resource names
- `validateToolArgs()` - Validates common tool arguments
- Applied to all tool handlers

**Files Changed**:
- `pkg/tools/security.go` (new file)
- All handler files in `pkg/tools/`

**Example**:
```go
// Before
namespace, _ := args["namespace"].(string)

// After
if err := validateToolArgs(args); err != nil {
    return nil, err
}
namespace, _ := args["namespace"].(string)
```

### 2. Log Injection Prevention

**Issue**: User input (namespaces, resource names, cluster names) was included in output without sanitization, allowing log injection attacks.

**Fix**: Added `sanitizeForLogging()` function that:
- Removes control characters and newlines
- Limits string length to prevent log flooding
- Escapes potentially dangerous characters

**Files Changed**:
- `pkg/tools/security.go` (new file)
- All handler files in `pkg/tools/`

**Example**:
```go
// Before
sb.WriteString(fmt.Sprintf("# Kustomization: %s/%s\n\n", namespace, name))

// After
sanitizedNamespace := sanitizeForLogging(namespace)
sanitizedName := sanitizeForLogging(name)
sb.WriteString(fmt.Sprintf("# Kustomization: %s/%s\n\n", sanitizedNamespace, sanitizedName))
```

### 3. Sensitive Data Redaction

**Issue**: Pod logs may contain sensitive information (secrets, tokens, passwords) that gets exposed to AI clients.

**Fix**: Added `redactSensitiveData()` function that redacts:
- Passwords, secrets, tokens, keys
- Bearer tokens
- Base64 encoded secrets
- AWS access keys
- Private keys
- JWT tokens

**Files Changed**:
- `pkg/tools/security.go` (new file)
- `pkg/tools/debug_handlers.go` (pod logs handler)

**Example**:
```go
// Before
sb.WriteString(buf.String())

// After
logContent := buf.String()
redactedLogs := redactSensitiveData(logContent)
sb.WriteString(redactedLogs)
```

### 4. Field Selector Injection Prevention

**Issue**: User input used directly in Kubernetes field selectors without sanitization.

**Fix**: Sanitize all user input before using in field selectors and label selectors.

**Files Changed**:
- `pkg/tools/debug_handlers.go`
- `pkg/tools/cluster_handlers.go`

**Example**:
```go
// Before
FieldSelector: fmt.Sprintf("involvedObject.name=%s", name)

// After
sanitizedName := sanitizeForLogging(name)
FieldSelector: fmt.Sprintf("involvedObject.name=%s", sanitizedName)
```

### 5. Enhanced Linting

**Issue**: Missing security-focused linters in golangci-lint configuration.

**Fix**: Added security linters:
- `gosec` - Security-focused static analysis
- `gocritic` - More opinionated checks including security
- `gofmt` - Format checking
- `goimports` - Import formatting

**Files Changed**:
- `.golangci.yml` (new file)

## Security Functions Added

### `sanitizeForLogging(input string) string`
Sanitizes user input before logging/output to prevent log injection.

### `validateNamespace(ns string) error`
Validates Kubernetes namespace names against Kubernetes naming rules.

### `validateResourceName(name string) error`
Validates Kubernetes resource names against Kubernetes naming rules.

### `validateToolArgs(args map[string]interface{}) error`
Validates common tool arguments (namespace, name, pod_name, resource_name).

### `redactSensitiveData(text string) string`
Redacts sensitive information from text (e.g., pod logs) before returning to client.

### `validatePath(path string) bool`
Validates file paths to prevent path traversal attacks (for future use).

## Testing

All security improvements are tested through:
- Unit tests (if applicable)
- Integration tests
- Security scanning workflows (CodeQL, Gosec, Trivy)

## Best Practices Applied

1. **Defense in Depth**: Multiple layers of security (validation, sanitization, redaction)
2. **Fail Secure**: Invalid inputs are rejected, not processed
3. **Least Privilege**: Minimal information exposed in outputs
4. **Input Validation**: All user input is validated and sanitized
5. **Sensitive Data Protection**: Secrets are redacted from logs before exposure

## Remaining Considerations

### For Production Deployment

1. **RBAC**: Always use `--read-only` mode and configure restricted ServiceAccount
2. **Namespace Allowlist**: Consider implementing namespace allowlist/blocklist
3. **Rate Limiting**: Consider adding rate limiting for expensive operations
4. **Audit Logging**: Consider adding audit logs for security events
5. **Resource Limits**: Already implemented (30s timeouts, event limits)

### For Further Hardening

1. **Namespace Restrictions**: Implement namespace allowlist/blocklist
2. **Resource Size Limits**: Add limits on response sizes
3. **Request Rate Limiting**: Add rate limiting per client
4. **Audit Logging**: Log all tool invocations for security auditing
5. **Error Message Sanitization**: Ensure error messages don't leak sensitive info

## MCP-Specific Security Considerations

### Why MCP is Inherently More Secure

1. **No Network Exposure**: Uses stdio (stdin/stdout), no network listener
2. **Process Isolation**: Communication is confined to parent-child process
3. **Same User Context**: Runs under the same user as the AI client
4. **No Shared Secrets**: No API keys or tokens passed between client/server

### Additional Security Measures

1. **Read-Only Mode**: Always use `--read-only` flag in production
2. **RBAC Configuration**: Use restricted ServiceAccount with minimal permissions
3. **Kubeconfig Security**: Use kubeconfig with limited cluster access
4. **Log Redaction**: Sensitive data is redacted from pod logs
5. **Input Validation**: All inputs are validated before processing

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [MCP Security Documentation](./SECURITY.md)
