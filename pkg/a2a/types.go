// Package a2a provides the Agent-to-Agent (A2A) protocol implementation.
// A2A is Google's open protocol for AI agents to communicate and collaborate.
//
// Key Differences from MCP:
// - MCP: Human/AI → Tool (single tool invocation)
// - A2A: AI → AI (task-based collaboration with state)
//
// Core Concepts:
// - AgentCard: Discovery mechanism (like a business card for agents)
// - Skill: What an agent can do (similar to MCP Tool)
// - Task: Stateful, potentially long-running operation
// - Message: Chat-like communication within a task
// - Artifact: Structured output from a task
package a2a

import (
	"time"
)

// Protocol version for A2A
const A2AProtocolVersion = "0.1.0"

// =============================================================================
// AGENT CARD - Discovery Mechanism
// =============================================================================

// AgentCard describes an agent's capabilities, exposed at /.well-known/agent.json
// This is how agents discover each other's capabilities.
type AgentCard struct {
	// Name is the human-readable name of the agent
	Name string `json:"name"`

	// Description explains what this agent does
	Description string `json:"description"`

	// Version of this agent
	Version string `json:"version"`

	// URL is the base URL where this agent can be reached
	URL string `json:"url"`

	// DocumentationURL points to detailed documentation
	DocumentationURL string `json:"documentationUrl,omitempty"`

	// Capabilities describes what protocol features this agent supports
	Capabilities AgentCapabilities `json:"capabilities"`

	// Skills lists what this agent can do
	Skills []Skill `json:"skills"`

	// Authentication describes how to authenticate with this agent
	Authentication *AuthenticationInfo `json:"authentication,omitempty"`

	// Provider information about who created this agent
	Provider *ProviderInfo `json:"provider,omitempty"`
}

// AgentCapabilities describes what A2A features the agent supports
type AgentCapabilities struct {
	// Streaming indicates the agent can stream responses
	Streaming bool `json:"streaming"`

	// PushNotifications indicates the agent can send push notifications
	PushNotifications bool `json:"pushNotifications"`

	// StateTransitionHistory indicates the agent tracks task state history
	StateTransitionHistory bool `json:"stateTransitionHistory"`
}

// AuthenticationInfo describes authentication requirements
type AuthenticationInfo struct {
	// Type of authentication: "none", "bearer", "oauth2", "api_key"
	Type string `json:"type"`

	// Required indicates if authentication is mandatory
	Required bool `json:"required"`

	// Schemes lists supported auth schemes (for multiple options)
	Schemes []string `json:"schemes,omitempty"`
}

// ProviderInfo describes who created the agent
type ProviderInfo struct {
	Organization string `json:"organization"`
	URL          string `json:"url,omitempty"`
}

// =============================================================================
// SKILLS - What an Agent Can Do
// =============================================================================

// Skill describes a capability that an agent has.
// This is similar to MCP's Tool, but with additional metadata.
type Skill struct {
	// ID is a unique identifier for this skill (kebab-case)
	ID string `json:"id"`

	// Name is a human-readable name
	Name string `json:"name"`

	// Description explains what this skill does
	Description string `json:"description"`

	// InputSchema defines the JSON Schema for skill inputs
	InputSchema InputSchema `json:"inputSchema,omitempty"`

	// OutputSchema defines the expected output format
	OutputSchema *OutputSchema `json:"outputSchema,omitempty"`

	// Tags for categorization and discovery
	Tags []string `json:"tags,omitempty"`

	// Examples show how to use this skill
	Examples []SkillExample `json:"examples,omitempty"`
}

// InputSchema defines the JSON Schema for inputs (same as MCP)
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property defines a JSON Schema property
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default,omitempty"`
}

// OutputSchema defines the expected output format
type OutputSchema struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// SkillExample shows how to use a skill
type SkillExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Input       map[string]interface{} `json:"input"`
}

// =============================================================================
// TASKS - Stateful Operations
// =============================================================================

// Task represents a unit of work assigned to an agent.
// Unlike MCP tool calls, tasks are stateful and can be long-running.
type Task struct {
	// ID is a unique identifier for this task
	ID string `json:"id"`

	// SessionID groups related tasks together
	SessionID string `json:"sessionId,omitempty"`

	// Skill is the ID of the skill to execute
	Skill string `json:"skill"`

	// Input contains the parameters for the skill
	Input map[string]interface{} `json:"input,omitempty"`

	// Status is the current state of the task
	Status TaskStatus `json:"status"`

	// Messages contains the conversation history for this task
	Messages []Message `json:"messages,omitempty"`

	// Artifacts contains outputs produced by the task
	Artifacts []Artifact `json:"artifacts,omitempty"`

	// Metadata contains additional context
	Metadata TaskMetadata `json:"metadata,omitempty"`

	// Error contains error information if the task failed
	Error *TaskError `json:"error,omitempty"`

	// CreatedAt is when the task was created
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the task was last modified
	UpdatedAt time.Time `json:"updatedAt"`

	// CompletedAt is when the task finished (if completed)
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// TaskStatus represents the state of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // Created but not started
	TaskStatusRunning    TaskStatus = "running"    // Currently executing
	TaskStatusCompleted  TaskStatus = "completed"  // Finished successfully
	TaskStatusFailed     TaskStatus = "failed"     // Finished with error
	TaskStatusCancelled  TaskStatus = "cancelled"  // Cancelled by user/agent
	TaskStatusInputNeeded TaskStatus = "input-needed" // Waiting for user input
)

// TaskMetadata contains additional context for a task
type TaskMetadata struct {
	// RequestingAgent is the agent that created this task
	RequestingAgent string `json:"requestingAgent,omitempty"`

	// Priority of the task
	Priority string `json:"priority,omitempty"`

	// Timeout in seconds for task execution
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`

	// Custom allows arbitrary metadata
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// TaskError contains error information
type TaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// =============================================================================
// MESSAGES - Chat-like Communication
// =============================================================================

// Message represents a message exchanged during task execution
type Message struct {
	// Role indicates who sent the message: "user", "agent", "system"
	Role string `json:"role"`

	// Content contains the message content
	Content []ContentPart `json:"content"`

	// Metadata for the message
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the message was created
	Timestamp time.Time `json:"timestamp"`
}

// ContentPart is a piece of content in a message
type ContentPart struct {
	// Type of content: "text", "image", "file", "data"
	Type string `json:"type"`

	// Text content (for type="text")
	Text string `json:"text,omitempty"`

	// MimeType for non-text content
	MimeType string `json:"mimeType,omitempty"`

	// Data contains the content (base64 for binary)
	Data string `json:"data,omitempty"`

	// File contains file information (for type="file")
	File *FileContent `json:"file,omitempty"`
}

// FileContent represents a file in a message
type FileContent struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
	Data     string `json:"data,omitempty"` // Base64 encoded
}

// =============================================================================
// ARTIFACTS - Structured Outputs
// =============================================================================

// Artifact represents a structured output from a task
type Artifact struct {
	// Name identifies this artifact
	Name string `json:"name"`

	// Description explains what this artifact contains
	Description string `json:"description,omitempty"`

	// MimeType of the artifact content
	MimeType string `json:"mimeType"`

	// Data contains the artifact content
	Data interface{} `json:"data"`

	// Index for ordering multiple artifacts
	Index int `json:"index,omitempty"`

	// Metadata for the artifact
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamp when the artifact was created
	Timestamp time.Time `json:"timestamp"`
}

// =============================================================================
// JSON-RPC TYPES - Protocol Messages
// =============================================================================

// A2ARequest represents an A2A JSON-RPC request
type A2ARequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// A2AResponse represents an A2A JSON-RPC response
type A2AResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *A2AError     `json:"error,omitempty"`
}

// A2AError represents a JSON-RPC error
type A2AError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternalError  = -32603

	// A2A-specific errors
	ErrTaskNotFound    = -32000
	ErrSkillNotFound   = -32001
	ErrTaskCancelled   = -32002
	ErrUnauthorized    = -32003
)

// =============================================================================
// REQUEST/RESPONSE TYPES
// =============================================================================

// TaskCreateRequest is the request body for tasks/create
type TaskCreateRequest struct {
	ID        string                 `json:"id,omitempty"` // Optional, server generates if empty
	SessionID string                 `json:"sessionId,omitempty"`
	Skill     string                 `json:"skill"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Metadata  TaskMetadata           `json:"metadata,omitempty"`
}

// TaskCreateResponse is the response for tasks/create
type TaskCreateResponse struct {
	Task *Task `json:"task"`
}

// TaskGetRequest is the request body for tasks/get
type TaskGetRequest struct {
	TaskID string `json:"taskId"`
}

// TaskGetResponse is the response for tasks/get
type TaskGetResponse struct {
	Task *Task `json:"task"`
}

// TaskCancelRequest is the request body for tasks/cancel
type TaskCancelRequest struct {
	TaskID string `json:"taskId"`
}

// TaskCancelResponse is the response for tasks/cancel
type TaskCancelResponse struct {
	Task *Task `json:"task"`
}

// TaskMessageRequest is the request body for tasks/message
type TaskMessageRequest struct {
	TaskID  string  `json:"taskId"`
	Message Message `json:"message"`
}

// TaskMessageResponse is the response for tasks/message
type TaskMessageResponse struct {
	Task *Task `json:"task"`
}

// AgentInfoResponse is the response for agent/info
type AgentInfoResponse struct {
	Agent *AgentCard `json:"agent"`
}
