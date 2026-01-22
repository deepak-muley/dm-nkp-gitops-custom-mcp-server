package a2a

import (
	"strings"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
)

// =============================================================================
// MCP to A2A CONVERTER
// =============================================================================
//
// This converter bridges MCP Tools to A2A Skills.
//
// Key differences to understand:
//
// MCP Tool:
//   - Designed for single tool invocations
//   - Uses snake_case names
//   - No state between calls
//   - Result is immediate
//
// A2A Skill:
//   - Designed for task-based execution
//   - Uses kebab-case IDs
//   - Tasks can be long-running
//   - Supports streaming, messages, artifacts
//
// The conversion allows you to expose your existing MCP tools as A2A skills,
// enabling agent-to-agent collaboration without rewriting handlers.
// =============================================================================

// Converter handles MCP to A2A conversions
type Converter struct {
	// agentName is used in converted skill descriptions
	agentName string

	// tags are default tags applied to all skills
	defaultTags []string
}

// NewConverter creates a new MCP to A2A converter
func NewConverter(agentName string) *Converter {
	return &Converter{
		agentName:   agentName,
		defaultTags: []string{"mcp-bridge"},
	}
}

// SetDefaultTags sets the default tags for converted skills
func (c *Converter) SetDefaultTags(tags []string) {
	c.defaultTags = tags
}

// ConvertTool converts a single MCP Tool to an A2A Skill
//
// Conversion rules:
// 1. Name: snake_case → kebab-case (e.g., get_gitops_status → get-gitops-status)
// 2. Description: Enhanced with agent context
// 3. InputSchema: Directly mapped (same JSON Schema format)
// 4. Tags: Auto-generated from tool name + defaults
func (c *Converter) ConvertTool(tool mcp.Tool) Skill {
	return Skill{
		ID:          c.toKebabCase(tool.Name),
		Name:        c.toHumanReadable(tool.Name),
		Description: tool.Description,
		InputSchema: c.convertInputSchema(tool.InputSchema),
		Tags:        c.generateTags(tool.Name),
		Examples:    c.generateExamples(tool),
	}
}

// ConvertTools converts multiple MCP Tools to A2A Skills
func (c *Converter) ConvertTools(tools []mcp.Tool) []Skill {
	skills := make([]Skill, len(tools))
	for i, tool := range tools {
		skills[i] = c.ConvertTool(tool)
	}
	return skills
}

// ConvertToolResult converts an MCP ToolCallResult to A2A artifacts/messages
//
// In MCP, a tool returns a single result with content.
// In A2A, we convert this to either:
// - A message (for text content)
// - An artifact (for structured data)
func (c *Converter) ConvertToolResult(result *mcp.ToolCallResult, skillID string) ([]Message, []Artifact) {
	var messages []Message
	var artifacts []Artifact

	for i, content := range result.Content {
		switch content.Type {
		case "text":
			// Text content becomes a message
			messages = append(messages, Message{
				Role: "agent",
				Content: []ContentPart{
					{Type: "text", Text: content.Text},
				},
			})

		default:
			// Other content types become artifacts
			artifacts = append(artifacts, Artifact{
				Name:     skillID + "-output",
				MimeType: content.MimeType,
				Data:     content.Data,
				Index:    i,
			})
		}
	}

	return messages, artifacts
}

// CreateAgentCard creates an A2A AgentCard from MCP server info
func (c *Converter) CreateAgentCard(
	name string,
	version string,
	description string,
	baseURL string,
	tools []mcp.Tool,
) AgentCard {
	return AgentCard{
		Name:        name,
		Description: description,
		Version:     version,
		URL:         baseURL,
		Capabilities: AgentCapabilities{
			Streaming:              false, // Can be enabled later
			PushNotifications:      false,
			StateTransitionHistory: true,
		},
		Skills: c.ConvertTools(tools),
		Authentication: &AuthenticationInfo{
			Type:     "none", // Start simple, add auth later
			Required: false,
		},
	}
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// toKebabCase converts snake_case to kebab-case
// Example: get_gitops_status → get-gitops-status
func (c *Converter) toKebabCase(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

// toHumanReadable converts snake_case to Title Case
// Example: get_gitops_status → Get Gitops Status
func (c *Converter) toHumanReadable(name string) string {
	words := strings.Split(name, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// convertInputSchema converts MCP InputSchema to A2A InputSchema
// Currently these are identical, but this allows for future divergence
func (c *Converter) convertInputSchema(mcpSchema mcp.InputSchema) InputSchema {
	a2aSchema := InputSchema{
		Type:     mcpSchema.Type,
		Required: mcpSchema.Required,
	}

	if len(mcpSchema.Properties) > 0 {
		a2aSchema.Properties = make(map[string]Property)
		for name, prop := range mcpSchema.Properties {
			a2aSchema.Properties[name] = Property{
				Type:        prop.Type,
				Description: prop.Description,
				Enum:        prop.Enum,
				Default:     prop.Default,
			}
		}
	}

	return a2aSchema
}

// generateTags creates tags from the tool name
func (c *Converter) generateTags(toolName string) []string {
	tags := make([]string, 0, len(c.defaultTags)+2)

	// Add default tags
	tags = append(tags, c.defaultTags...)

	// Extract domain from tool name
	parts := strings.Split(toolName, "_")
	if len(parts) > 1 {
		// First part often indicates the domain
		domain := parts[0]
		switch domain {
		case "get", "list", "check":
			tags = append(tags, "read-only")
		case "debug":
			tags = append(tags, "debugging")
		}

		// Second part often indicates the resource type
		if len(parts) > 1 {
			resource := parts[1]
			switch resource {
			case "gitops", "kustomizations", "gitrepositories":
				tags = append(tags, "gitops", "flux")
			case "cluster", "machines":
				tags = append(tags, "capi", "cluster")
			case "app", "helmreleases":
				tags = append(tags, "apps", "deployment")
			case "events", "pod":
				tags = append(tags, "debugging", "kubernetes")
			case "policy", "constraints":
				tags = append(tags, "policy", "security")
			case "contexts":
				tags = append(tags, "kubernetes", "config")
			}
		}
	}

	return tags
}

// generateExamples creates example invocations for a skill
func (c *Converter) generateExamples(tool mcp.Tool) []SkillExample {
	// Generate basic examples based on the tool's properties
	examples := []SkillExample{}

	// Example 1: Basic invocation (no params or defaults)
	examples = append(examples, SkillExample{
		Name:        "Basic usage",
		Description: "Invoke with default parameters",
		Input:       map[string]interface{}{},
	})

	// Example 2: With common parameters if they exist
	if len(tool.InputSchema.Properties) > 0 {
		exampleInput := map[string]interface{}{}
		for name, prop := range tool.InputSchema.Properties {
			switch name {
			case "namespace":
				exampleInput["namespace"] = "flux-system"
			case "name":
				exampleInput["name"] = "example-resource"
			case "cluster_name":
				exampleInput["cluster_name"] = "workload-cluster-1"
			case "status_filter":
				if len(prop.Enum) > 0 {
					exampleInput["status_filter"] = prop.Enum[0]
				}
			}
		}
		if len(exampleInput) > 0 {
			examples = append(examples, SkillExample{
				Name:        "With parameters",
				Description: "Invoke with specific parameters",
				Input:       exampleInput,
			})
		}
	}

	return examples
}

// =============================================================================
// REVERSE CONVERSION (A2A to MCP) - For interoperability
// =============================================================================

// ConvertSkillToTool converts an A2A Skill back to an MCP Tool
// This is useful if you want to use A2A skills as MCP tools
func (c *Converter) ConvertSkillToTool(skill Skill) mcp.Tool {
	return mcp.Tool{
		Name:        c.toSnakeCase(skill.ID),
		Description: skill.Description,
		InputSchema: c.convertToMCPInputSchema(skill.InputSchema),
	}
}

// toSnakeCase converts kebab-case to snake_case
func (c *Converter) toSnakeCase(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}

// convertToMCPInputSchema converts A2A InputSchema to MCP InputSchema
func (c *Converter) convertToMCPInputSchema(a2aSchema InputSchema) mcp.InputSchema {
	mcpSchema := mcp.InputSchema{
		Type:     a2aSchema.Type,
		Required: a2aSchema.Required,
	}

	if len(a2aSchema.Properties) > 0 {
		mcpSchema.Properties = make(map[string]mcp.Property)
		for name, prop := range a2aSchema.Properties {
			mcpSchema.Properties[name] = mcp.Property{
				Type:        prop.Type,
				Description: prop.Description,
				Enum:        prop.Enum,
				Default:     prop.Default,
			}
		}
	}

	return mcpSchema
}
