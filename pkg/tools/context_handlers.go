package tools

import (
	"fmt"
	"strings"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
)

// handleListContexts handles the list_contexts tool.
func (r *Registry) handleListContexts(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	var sb strings.Builder
	sb.WriteString("# Available Kubernetes Contexts\n\n")

	if len(r.clients.AvailableContexts) == 0 {
		sb.WriteString("No contexts found in kubeconfig.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Current context: **%s**\n\n", r.clients.CurrentContext))
		sb.WriteString("| Context | Current |\n")
		sb.WriteString("|---------|:-------:|\n")

		for _, ctx := range r.clients.AvailableContexts {
			current := ""
			if ctx == r.clients.CurrentContext {
				current = "✓"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", ctx, current))
		}
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleGetCurrentContext handles the get_current_context tool.
func (r *Registry) handleGetCurrentContext(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	var sb strings.Builder
	sb.WriteString("# Current Kubernetes Context\n\n")
	sb.WriteString(fmt.Sprintf("**Context:** %s\n\n", r.clients.CurrentContext))
	sb.WriteString(fmt.Sprintf("**Server:** %s\n", r.clients.RestConfig.Host))

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}
