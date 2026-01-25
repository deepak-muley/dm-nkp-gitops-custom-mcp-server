package tools

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// sanitizeForLogging removes or escapes potentially dangerous characters from user input
// to prevent log injection attacks
func sanitizeForLogging(input string) string {
	// Remove control characters and newlines that could be used for log injection
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) || r == ' ' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('?')
		}
	}
	result := builder.String()
	// Limit length to prevent log flooding
	if len(result) > 500 {
		result = result[:500] + "..."
	}
	return result
}

// validateNamespace validates Kubernetes namespace names
func validateNamespace(ns string) error {
	if ns == "" {
		return nil // Empty namespace is valid (means "all namespaces")
	}
	if len(ns) > 253 {
		return fmt.Errorf("namespace too long (max 253 characters)")
	}
	// Kubernetes namespace name regex: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	namespaceRegex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !namespaceRegex.MatchString(ns) {
		return fmt.Errorf("invalid namespace format (must match Kubernetes naming rules)")
	}
	return nil
}

// validateResourceName validates Kubernetes resource names
func validateResourceName(name string) error {
	if name == "" {
		return fmt.Errorf("resource name is required")
	}
	if len(name) > 253 {
		return fmt.Errorf("resource name too long (max 253 characters)")
	}
	// Kubernetes resource name regex: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	// But also allow uppercase for some resources
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`)
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("invalid resource name format (must match Kubernetes naming rules)")
	}
	return nil
}

// redactSensitiveData redacts common secret patterns from text (e.g., pod logs)
var sensitivePatterns = []*regexp.Regexp{
	// Passwords, secrets, tokens, keys
	regexp.MustCompile(`(?i)(password|secret|token|key|credential|api[_-]?key|auth[_-]?token)[\s]*[=:]\s*([^\s\n]+)`),
	// Bearer tokens
	regexp.MustCompile(`(?i)bearer\s+([a-zA-Z0-9\-._~+/]+=*)`),
	// Base64 encoded secrets (long base64 strings)
	regexp.MustCompile(`([A-Za-z0-9+/]{40,}={0,2})`),
	// AWS access keys
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	// Private keys (RSA, EC, etc.)
	regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`),
	// JWT tokens (basic pattern)
	regexp.MustCompile(`eyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`),
}

// redactSensitiveData redacts sensitive information from text
func redactSensitiveData(text string) string {
	result := text
	for _, pattern := range sensitivePatterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}
	return result
}

// validatePath ensures the path is safe (for file operations if any)
func validatePath(path string) bool {
	// Reject paths with control characters or suspicious patterns
	if len(path) > 2048 {
		return false
	}
	for _, r := range path {
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	// Reject path traversal attempts
	if strings.Contains(path, "..") {
		return false
	}
	return true
}

// validateToolArgs validates common tool arguments (namespace, name)
func validateToolArgs(args map[string]interface{}) error {
	// Validate namespace if present
	if ns, ok := args["namespace"].(string); ok && ns != "" {
		if err := validateNamespace(ns); err != nil {
			return fmt.Errorf("invalid namespace: %w", err)
		}
	}
	
	// Validate name if present
	if name, ok := args["name"].(string); ok && name != "" {
		if err := validateResourceName(name); err != nil {
			return fmt.Errorf("invalid resource name: %w", err)
		}
	}
	
	// Validate pod_name if present
	if podName, ok := args["pod_name"].(string); ok && podName != "" {
		if err := validateResourceName(podName); err != nil {
			return fmt.Errorf("invalid pod name: %w", err)
		}
	}
	
	// Validate resource_name if present
	if resourceName, ok := args["resource_name"].(string); ok && resourceName != "" {
		if err := validateResourceName(resourceName); err != nil {
			return fmt.Errorf("invalid resource name: %w", err)
		}
	}
	
	return nil
}
