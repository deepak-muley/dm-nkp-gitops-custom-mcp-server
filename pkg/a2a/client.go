package a2a

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// =============================================================================
// A2A CLIENT
// =============================================================================
//
// The A2A Client enables agent-to-agent communication.
//
// This is how one agent calls another agent's skills. In a multi-agent system:
//
//   Orchestrator Agent
//         │
//         ├── A2A Client ──► GitOps Agent (this server)
//         │
//         ├── A2A Client ──► Security Agent
//         │
//         └── A2A Client ──► Monitoring Agent
//
// Usage:
//
//   client := a2a.NewClient("http://gitops-agent:8080")
//
//   // Discover agent capabilities
//   card, _ := client.GetAgentCard(ctx)
//
//   // Execute a skill
//   task, _ := client.CreateTask(ctx, "get-gitops-status", map[string]interface{}{
//       "namespace": "flux-system",
//   })
//
//   // Wait for completion
//   result, _ := client.WaitForTask(ctx, task.ID, 30*time.Second)
//
// =============================================================================

// Client is an A2A client for communicating with other agents
type Client struct {
	// BaseURL is the base URL of the target agent
	BaseURL string

	// HTTPClient is the underlying HTTP client
	HTTPClient *http.Client

	// AgentCard caches the target agent's capabilities
	AgentCard *AgentCard

	// requestID counter for JSON-RPC
	requestID int
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.HTTPClient.Timeout = timeout
	}
}

// NewClient creates a new A2A client for the given agent URL
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		requestID: 0,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// =============================================================================
// DISCOVERY
// =============================================================================

// GetAgentCard fetches the agent's capabilities (discovery)
// This is typically the first call to understand what skills are available
func (c *Client) GetAgentCard(ctx context.Context) (*AgentCard, error) {
	url := c.BaseURL + "/.well-known/agent.json"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent card: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var card AgentCard
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("failed to decode agent card: %w", err)
	}

	// Cache for later use
	c.AgentCard = &card
	return &card, nil
}

// GetAgentInfo fetches agent info via JSON-RPC
func (c *Client) GetAgentInfo(ctx context.Context) (*AgentCard, error) {
	var result AgentInfoResponse
	if err := c.call(ctx, "agent/info", nil, &result); err != nil {
		return nil, err
	}
	return result.Agent, nil
}

// HasSkill checks if the agent has a specific skill
func (c *Client) HasSkill(ctx context.Context, skillID string) (bool, error) {
	if c.AgentCard == nil {
		if _, err := c.GetAgentCard(ctx); err != nil {
			return false, err
		}
	}

	for _, skill := range c.AgentCard.Skills {
		if skill.ID == skillID {
			return true, nil
		}
	}
	return false, nil
}

// GetSkill returns information about a specific skill
func (c *Client) GetSkill(ctx context.Context, skillID string) (*Skill, error) {
	if c.AgentCard == nil {
		if _, err := c.GetAgentCard(ctx); err != nil {
			return nil, err
		}
	}

	for _, skill := range c.AgentCard.Skills {
		if skill.ID == skillID {
			return &skill, nil
		}
	}
	return nil, fmt.Errorf("skill not found: %s", skillID)
}

// =============================================================================
// TASK OPERATIONS
// =============================================================================

// CreateTask creates a new task on the remote agent
func (c *Client) CreateTask(ctx context.Context, skillID string, input map[string]interface{}) (*Task, error) {
	params := TaskCreateRequest{
		Skill: skillID,
		Input: input,
	}

	var result TaskCreateResponse
	if err := c.call(ctx, "tasks/create", params, &result); err != nil {
		return nil, err
	}

	return result.Task, nil
}

// CreateTaskWithMetadata creates a task with additional metadata
func (c *Client) CreateTaskWithMetadata(ctx context.Context, req TaskCreateRequest) (*Task, error) {
	var result TaskCreateResponse
	if err := c.call(ctx, "tasks/create", req, &result); err != nil {
		return nil, err
	}

	return result.Task, nil
}

// GetTask retrieves the current state of a task
func (c *Client) GetTask(ctx context.Context, taskID string) (*Task, error) {
	params := TaskGetRequest{TaskID: taskID}

	var result TaskGetResponse
	if err := c.call(ctx, "tasks/get", params, &result); err != nil {
		return nil, err
	}

	return result.Task, nil
}

// CancelTask cancels a running task
func (c *Client) CancelTask(ctx context.Context, taskID string) (*Task, error) {
	params := TaskCancelRequest{TaskID: taskID}

	var result TaskCancelResponse
	if err := c.call(ctx, "tasks/cancel", params, &result); err != nil {
		return nil, err
	}

	return result.Task, nil
}

// SendMessage sends a message to a task
func (c *Client) SendMessage(ctx context.Context, taskID string, message Message) (*Task, error) {
	params := TaskMessageRequest{
		TaskID:  taskID,
		Message: message,
	}

	var result TaskMessageResponse
	if err := c.call(ctx, "tasks/message", params, &result); err != nil {
		return nil, err
	}

	return result.Task, nil
}

// ListTasks lists all tasks on the remote agent
func (c *Client) ListTasks(ctx context.Context, statusFilter string) ([]*Task, error) {
	params := map[string]string{}
	if statusFilter != "" {
		params["status"] = statusFilter
	}

	var result struct {
		Tasks []*Task `json:"tasks"`
	}
	if err := c.call(ctx, "tasks/list", params, &result); err != nil {
		return nil, err
	}

	return result.Tasks, nil
}

// =============================================================================
// CONVENIENCE METHODS
// =============================================================================

// ExecuteSkill creates a task and waits for it to complete
// This is the simplest way to call a skill synchronously
func (c *Client) ExecuteSkill(ctx context.Context, skillID string, input map[string]interface{}, timeout time.Duration) (*Task, error) {
	// Create the task
	task, err := c.CreateTask(ctx, skillID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Wait for completion
	return c.WaitForTask(ctx, task.ID, timeout)
}

// WaitForTask polls a task until it reaches a terminal state
func (c *Client) WaitForTask(ctx context.Context, taskID string, timeout time.Duration) (*Task, error) {
	deadline := time.Now().Add(timeout)
	pollInterval := 100 * time.Millisecond

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		task, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, err
		}

		switch task.Status {
		case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
			return task, nil
		case TaskStatusRunning, TaskStatusPending:
			// Continue polling
		}

		time.Sleep(pollInterval)
		// Exponential backoff up to 2 seconds
		if pollInterval < 2*time.Second {
			pollInterval = pollInterval * 2
		}
	}

	// Timeout - try to cancel
	c.CancelTask(ctx, taskID)
	return c.GetTask(ctx, taskID)
}

// ExecuteSkillAndGetText executes a skill and returns the text result
func (c *Client) ExecuteSkillAndGetText(ctx context.Context, skillID string, input map[string]interface{}, timeout time.Duration) (string, error) {
	task, err := c.ExecuteSkill(ctx, skillID, input, timeout)
	if err != nil {
		return "", err
	}

	if task.Status == TaskStatusFailed {
		return "", fmt.Errorf("task failed: %s", task.Error.Message)
	}

	if task.Status == TaskStatusCancelled {
		return "", fmt.Errorf("task was cancelled")
	}

	// Extract text from messages
	var text string
	for _, msg := range task.Messages {
		for _, content := range msg.Content {
			if content.Type == "text" {
				text += content.Text
			}
		}
	}

	return text, nil
}

// =============================================================================
// JSON-RPC COMMUNICATION
// =============================================================================

// call makes a JSON-RPC call to the remote agent
func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	c.requestID++

	request := A2ARequest{
		JSONRPC: "2.0",
		ID:      c.requestID,
		Method:  method,
		Params:  params,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var response A2AResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	// Unmarshal result into the provided type
	if result != nil {
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Health checks if the agent is healthy
func (c *Client) Health(ctx context.Context) (map[string]interface{}, error) {
	url := c.BaseURL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return result, nil
}
