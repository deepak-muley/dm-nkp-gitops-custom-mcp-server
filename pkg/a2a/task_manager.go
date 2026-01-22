package a2a

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
	"github.com/google/uuid"
)

// =============================================================================
// TASK MANAGER
// =============================================================================
//
// The TaskManager is a KEY DIFFERENCE between MCP and A2A:
//
// MCP:
//   - Stateless: Each tool call is independent
//   - Synchronous: Request → Response, done
//   - No tracking: Caller manages state
//
// A2A:
//   - Stateful: Tasks persist and have lifecycle
//   - Async-capable: Tasks can be long-running
//   - Tracked: Server manages task state, history
//
// The TaskManager bridges your existing MCP handlers to A2A's task model:
// 1. Create a task from skill invocation
// 2. Execute the underlying MCP handler
// 3. Update task state and store results
// 4. Support cancellation, progress, messages
// =============================================================================

// Logger interface for logging
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// SkillHandler is a function that executes a skill
// This wraps MCP's ToolHandler with context for cancellation
type SkillHandler func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error)

// TaskManager manages the lifecycle of A2A tasks
type TaskManager struct {
	// tasks stores all tasks by ID
	tasks map[string]*Task

	// skillHandlers maps skill IDs to their handlers
	skillHandlers map[string]SkillHandler

	// activeTasks tracks running task contexts for cancellation
	activeTasks map[string]context.CancelFunc

	// converter for MCP ↔ A2A conversion
	converter *Converter

	// logger for logging
	logger Logger

	// mu protects concurrent access
	mu sync.RWMutex

	// taskHistory for completed tasks (for learning/debugging)
	taskHistory []*Task
	historyLimit int
}

// NewTaskManager creates a new TaskManager
func NewTaskManager(logger Logger) *TaskManager {
	return &TaskManager{
		tasks:         make(map[string]*Task),
		skillHandlers: make(map[string]SkillHandler),
		activeTasks:   make(map[string]context.CancelFunc),
		converter:     NewConverter("gitops-agent"),
		logger:        logger,
		taskHistory:   make([]*Task, 0),
		historyLimit:  100, // Keep last 100 completed tasks
	}
}

// =============================================================================
// HANDLER REGISTRATION
// =============================================================================

// RegisterSkillHandler registers a handler for a skill
func (tm *TaskManager) RegisterSkillHandler(skillID string, handler SkillHandler) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.skillHandlers[skillID] = handler
}

// RegisterMCPHandler wraps an MCP ToolHandler as an A2A SkillHandler
// This is the bridge from MCP to A2A
func (tm *TaskManager) RegisterMCPHandler(skillID string, mcpHandler mcp.ToolHandler) {
	// Wrap MCP handler with context support
	skillHandler := func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error) {
		// Check for cancellation before executing
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute the MCP handler
		// Note: MCP handlers don't support context, so we can't cancel mid-execution
		// This is a limitation of the bridge - future MCP handlers could be context-aware
		return mcpHandler(args)
	}

	tm.RegisterSkillHandler(skillID, skillHandler)
}

// RegisterMCPHandlers bulk registers MCP handlers as A2A skill handlers
func (tm *TaskManager) RegisterMCPHandlers(mcpHandlers map[string]mcp.ToolHandler) {
	for toolName, handler := range mcpHandlers {
		skillID := tm.converter.toKebabCase(toolName)
		tm.RegisterMCPHandler(skillID, handler)
	}
}

// =============================================================================
// TASK LIFECYCLE
// =============================================================================

// CreateTask creates a new task from a request
func (tm *TaskManager) CreateTask(req TaskCreateRequest) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate ID if not provided
	taskID := req.ID
	if taskID == "" {
		taskID = uuid.New().String()
	}

	// Check if skill exists
	if _, exists := tm.skillHandlers[req.Skill]; !exists {
		return nil, fmt.Errorf("skill not found: %s", req.Skill)
	}

	now := time.Now()
	task := &Task{
		ID:        taskID,
		SessionID: req.SessionID,
		Skill:     req.Skill,
		Input:     req.Input,
		Status:    TaskStatusPending,
		Messages:  []Message{},
		Artifacts: []Artifact{},
		Metadata:  req.Metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tm.tasks[taskID] = task
	tm.logger.Info("Task created", "taskId", taskID, "skill", req.Skill)

	return task, nil
}

// ExecuteTask starts executing a task asynchronously
func (tm *TaskManager) ExecuteTask(taskID string) error {
	tm.mu.Lock()
	task, exists := tm.tasks[taskID]
	if !exists {
		tm.mu.Unlock()
		return fmt.Errorf("task not found: %s", taskID)
	}

	handler, hasHandler := tm.skillHandlers[task.Skill]
	if !hasHandler {
		tm.mu.Unlock()
		return fmt.Errorf("no handler for skill: %s", task.Skill)
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	tm.activeTasks[taskID] = cancel

	// Update status
	task.Status = TaskStatusRunning
	task.UpdatedAt = time.Now()
	tm.mu.Unlock()

	// Execute asynchronously
	go tm.executeTaskAsync(ctx, task, handler)

	return nil
}

// executeTaskAsync runs the task handler and updates state
func (tm *TaskManager) executeTaskAsync(ctx context.Context, task *Task, handler SkillHandler) {
	tm.logger.Debug("Executing task", "taskId", task.ID, "skill", task.Skill)

	// Execute the handler
	result, err := handler(ctx, task.Input)

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Clean up active task
	delete(tm.activeTasks, task.ID)

	now := time.Now()
	task.UpdatedAt = now
	task.CompletedAt = &now

	if ctx.Err() == context.Canceled {
		// Task was cancelled
		task.Status = TaskStatusCancelled
		tm.logger.Info("Task cancelled", "taskId", task.ID)
		tm.archiveTask(task)
		return
	}

	if err != nil {
		// Task failed
		task.Status = TaskStatusFailed
		task.Error = &TaskError{
			Code:    "EXECUTION_ERROR",
			Message: err.Error(),
		}
		tm.logger.Error("Task failed", "taskId", task.ID, "error", err)
		tm.archiveTask(task)
		return
	}

	// Task succeeded - convert result to A2A format
	messages, artifacts := tm.converter.ConvertToolResult(result, task.Skill)
	task.Messages = append(task.Messages, messages...)
	task.Artifacts = append(task.Artifacts, artifacts...)
	task.Status = TaskStatusCompleted

	tm.logger.Info("Task completed", "taskId", task.ID,
		"messages", len(messages), "artifacts", len(artifacts))
	tm.archiveTask(task)
}

// CreateAndExecuteTask is a convenience method that creates and executes in one call
func (tm *TaskManager) CreateAndExecuteTask(req TaskCreateRequest) (*Task, error) {
	task, err := tm.CreateTask(req)
	if err != nil {
		return nil, err
	}

	if err := tm.ExecuteTask(task.ID); err != nil {
		return task, err // Return task even on error so caller can see status
	}

	return task, nil
}

// CreateAndExecuteTaskSync creates, executes, and waits for completion
// This is useful for simple request-response patterns
func (tm *TaskManager) CreateAndExecuteTaskSync(req TaskCreateRequest, timeout time.Duration) (*Task, error) {
	task, err := tm.CreateAndExecuteTask(req)
	if err != nil {
		return task, err
	}

	// Wait for completion with timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		currentTask, err := tm.GetTask(task.ID)
		if err != nil {
			return nil, err
		}

		switch currentTask.Status {
		case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
			return currentTask, nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	// Timeout - cancel the task
	tm.CancelTask(task.ID)
	return tm.GetTask(task.ID)
}

// GetTask retrieves a task by ID
func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Return a copy to prevent concurrent modification
	taskCopy := *task
	return &taskCopy, nil
}

// CancelTask cancels a running task
func (tm *TaskManager) CancelTask(taskID string) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Cancel if running
	if cancel, active := tm.activeTasks[taskID]; active {
		cancel()
		task.Status = TaskStatusCancelled
		task.UpdatedAt = time.Now()
		tm.logger.Info("Task cancellation requested", "taskId", taskID)
	}

	taskCopy := *task
	return &taskCopy, nil
}

// AddMessage adds a message to a task
func (tm *TaskManager) AddMessage(taskID string, message Message) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	message.Timestamp = time.Now()
	task.Messages = append(task.Messages, message)
	task.UpdatedAt = time.Now()

	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks returns all tasks (optionally filtered by status)
func (tm *TaskManager) ListTasks(statusFilter TaskStatus) []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var tasks []*Task
	for _, task := range tm.tasks {
		if statusFilter == "" || task.Status == statusFilter {
			taskCopy := *task
			tasks = append(tasks, &taskCopy)
		}
	}
	return tasks
}

// GetTaskHistory returns recently completed tasks
func (tm *TaskManager) GetTaskHistory(limit int) []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	if limit <= 0 || limit > len(tm.taskHistory) {
		limit = len(tm.taskHistory)
	}

	// Return most recent tasks
	start := len(tm.taskHistory) - limit
	if start < 0 {
		start = 0
	}

	history := make([]*Task, limit)
	copy(history, tm.taskHistory[start:])
	return history
}

// =============================================================================
// INTERNAL HELPERS
// =============================================================================

// archiveTask moves a completed task to history
func (tm *TaskManager) archiveTask(task *Task) {
	// Add to history
	tm.taskHistory = append(tm.taskHistory, task)

	// Trim history if needed
	if len(tm.taskHistory) > tm.historyLimit {
		tm.taskHistory = tm.taskHistory[1:]
	}
}

// GetStats returns task manager statistics
func (tm *TaskManager) GetStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var pending, running, completed, failed, cancelled int
	for _, task := range tm.tasks {
		switch task.Status {
		case TaskStatusPending:
			pending++
		case TaskStatusRunning:
			running++
		case TaskStatusCompleted:
			completed++
		case TaskStatusFailed:
			failed++
		case TaskStatusCancelled:
			cancelled++
		}
	}

	return map[string]interface{}{
		"totalTasks":     len(tm.tasks),
		"pending":        pending,
		"running":        running,
		"completed":      completed,
		"failed":         failed,
		"cancelled":      cancelled,
		"activeTasks":    len(tm.activeTasks),
		"historySize":    len(tm.taskHistory),
		"registeredSkills": len(tm.skillHandlers),
	}
}
