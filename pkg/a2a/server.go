package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
)

// =============================================================================
// A2A HTTP SERVER
// =============================================================================
//
// This is the A2A HTTP server that exposes your MCP tools as A2A skills.
//
// Key Endpoints (A2A Protocol):
//
//   GET  /.well-known/agent.json  - Agent Card (discovery)
//   POST /                         - JSON-RPC 2.0 endpoint for all A2A methods
//
// JSON-RPC Methods:
//   - agent/info     - Get agent information
//   - tasks/create   - Create a new task
//   - tasks/get      - Get task status
//   - tasks/cancel   - Cancel a running task
//   - tasks/message  - Add a message to a task
//   - tasks/list     - List all tasks
//
// COMPARISON TO MCP:
//
// MCP Server (stdio):
//   - Reads JSON-RPC from stdin, writes to stdout
//   - Single process communication
//   - Designed for local AI assistant integration
//
// A2A Server (HTTP):
//   - HTTP server exposing JSON-RPC endpoint
//   - Network communication between agents
//   - Designed for distributed multi-agent systems
//
// =============================================================================

// ServerConfig contains configuration for the A2A server
type ServerConfig struct {
	// Name of the agent
	Name string

	// Version of the agent
	Version string

	// Description of what this agent does
	Description string

	// Port to listen on
	Port int

	// BaseURL for agent card (auto-generated if empty)
	BaseURL string

	// MCP Tools to expose as A2A skills
	Tools []mcp.Tool

	// MCP Handlers for tool execution
	Handlers map[string]mcp.ToolHandler

	// Logger for logging
	Logger Logger

	// ReadOnly mode (informational only)
	ReadOnly bool
}

// Server is the A2A HTTP server
type Server struct {
	config      ServerConfig
	agentCard   AgentCard
	taskManager *TaskManager
	converter   *Converter
	httpServer  *http.Server
}

// NewServer creates a new A2A server
func NewServer(config ServerConfig) *Server {
	// Default port
	if config.Port == 0 {
		config.Port = 8080
	}

	// Default base URL
	if config.BaseURL == "" {
		config.BaseURL = fmt.Sprintf("http://localhost:%d", config.Port)
	}

	// Create converter
	converter := NewConverter(config.Name)
	converter.SetDefaultTags([]string{"gitops", "kubernetes", "nkp"})

	// Create agent card
	agentCard := converter.CreateAgentCard(
		config.Name,
		config.Version,
		config.Description,
		config.BaseURL,
		config.Tools,
	)

	// Create task manager
	taskManager := NewTaskManager(config.Logger)

	// Register MCP handlers as skill handlers
	taskManager.RegisterMCPHandlers(config.Handlers)

	return &Server{
		config:      config,
		agentCard:   agentCard,
		taskManager: taskManager,
		converter:   converter,
	}
}

// =============================================================================
// HTTP SERVER
// =============================================================================

// Run starts the A2A HTTP server
func (s *Server) Run() error {
	mux := http.NewServeMux()

	// Agent Card endpoint (A2A discovery)
	mux.HandleFunc("/.well-known/agent.json", s.handleAgentCard)
	mux.HandleFunc("/agent.json", s.handleAgentCard) // Alias for convenience

	// Health check
	mux.HandleFunc("/health", s.handleHealth)

	// JSON-RPC endpoint
	mux.HandleFunc("/", s.handleJSONRPC)

	// Create server
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.corsMiddleware(s.loggingMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	s.config.Logger.Info("A2A server starting",
		"port", s.config.Port,
		"skills", len(s.agentCard.Skills),
		"agentCard", fmt.Sprintf("%s/.well-known/agent.json", s.config.BaseURL),
	)

	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.config.Logger.Info("A2A server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// =============================================================================
// MIDDLEWARE
// =============================================================================

// loggingMiddleware logs incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.config.Logger.Debug("Request received",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
		)

		next.ServeHTTP(w, r)

		s.config.Logger.Debug("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start),
		)
	})
}

// corsMiddleware adds CORS headers for browser-based agents
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// =============================================================================
// ENDPOINT HANDLERS
// =============================================================================

// handleAgentCard returns the Agent Card for discovery
// This is the A2A equivalent of MCP's initialize response
func (s *Server) handleAgentCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.agentCard)
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	stats := s.taskManager.GetStats()
	stats["status"] = "healthy"
	stats["version"] = s.config.Version

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleJSONRPC handles all JSON-RPC requests
func (s *Server) handleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		// For GET on root, redirect to agent card info
		if r.Method == "GET" && r.URL.Path == "/" {
			s.handleAgentCard(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var request A2ARequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, nil, ErrParseError, "Parse error: "+err.Error())
		return
	}

	s.config.Logger.Debug("JSON-RPC request", "method", request.Method, "id", request.ID)

	// Route to handler
	var result interface{}
	var a2aErr *A2AError

	switch request.Method {
	case "agent/info":
		result, a2aErr = s.handleAgentInfo(request)
	case "tasks/create":
		result, a2aErr = s.handleTaskCreate(request)
	case "tasks/get":
		result, a2aErr = s.handleTaskGet(request)
	case "tasks/cancel":
		result, a2aErr = s.handleTaskCancel(request)
	case "tasks/message":
		result, a2aErr = s.handleTaskMessage(request)
	case "tasks/list":
		result, a2aErr = s.handleTaskList(request)
	default:
		a2aErr = &A2AError{
			Code:    ErrMethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}

	// Write response
	if a2aErr != nil {
		s.writeError(w, request.ID, a2aErr.Code, a2aErr.Message)
		return
	}

	s.writeSuccess(w, request.ID, result)
}

// =============================================================================
// JSON-RPC METHOD HANDLERS
// =============================================================================

// handleAgentInfo returns agent information
func (s *Server) handleAgentInfo(request A2ARequest) (interface{}, *A2AError) {
	return AgentInfoResponse{Agent: &s.agentCard}, nil
}

// handleTaskCreate creates a new task
func (s *Server) handleTaskCreate(request A2ARequest) (interface{}, *A2AError) {
	// Parse params
	paramsBytes, _ := json.Marshal(request.Params)
	var params TaskCreateRequest
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return nil, &A2AError{Code: ErrInvalidParams, Message: "Invalid params: " + err.Error()}
	}

	// Validate skill exists
	skillFound := false
	for _, skill := range s.agentCard.Skills {
		if skill.ID == params.Skill {
			skillFound = true
			break
		}
	}
	if !skillFound {
		return nil, &A2AError{Code: ErrSkillNotFound, Message: "Skill not found: " + params.Skill}
	}

	// Create and execute task
	task, err := s.taskManager.CreateAndExecuteTask(params)
	if err != nil {
		return nil, &A2AError{Code: ErrInternalError, Message: err.Error()}
	}

	return TaskCreateResponse{Task: task}, nil
}

// handleTaskGet retrieves a task by ID
func (s *Server) handleTaskGet(request A2ARequest) (interface{}, *A2AError) {
	paramsBytes, _ := json.Marshal(request.Params)
	var params TaskGetRequest
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return nil, &A2AError{Code: ErrInvalidParams, Message: "Invalid params"}
	}

	task, err := s.taskManager.GetTask(params.TaskID)
	if err != nil {
		return nil, &A2AError{Code: ErrTaskNotFound, Message: err.Error()}
	}

	return TaskGetResponse{Task: task}, nil
}

// handleTaskCancel cancels a running task
func (s *Server) handleTaskCancel(request A2ARequest) (interface{}, *A2AError) {
	paramsBytes, _ := json.Marshal(request.Params)
	var params TaskCancelRequest
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return nil, &A2AError{Code: ErrInvalidParams, Message: "Invalid params"}
	}

	task, err := s.taskManager.CancelTask(params.TaskID)
	if err != nil {
		return nil, &A2AError{Code: ErrTaskNotFound, Message: err.Error()}
	}

	return TaskCancelResponse{Task: task}, nil
}

// handleTaskMessage adds a message to a task
func (s *Server) handleTaskMessage(request A2ARequest) (interface{}, *A2AError) {
	paramsBytes, _ := json.Marshal(request.Params)
	var params TaskMessageRequest
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return nil, &A2AError{Code: ErrInvalidParams, Message: "Invalid params"}
	}

	task, err := s.taskManager.AddMessage(params.TaskID, params.Message)
	if err != nil {
		return nil, &A2AError{Code: ErrTaskNotFound, Message: err.Error()}
	}

	return TaskMessageResponse{Task: task}, nil
}

// handleTaskList returns all tasks
func (s *Server) handleTaskList(request A2ARequest) (interface{}, *A2AError) {
	// Parse optional filter
	var statusFilter TaskStatus
	if request.Params != nil {
		paramsBytes, _ := json.Marshal(request.Params)
		var params struct {
			Status string `json:"status"`
		}
		json.Unmarshal(paramsBytes, &params)
		if params.Status != "" {
			statusFilter = TaskStatus(params.Status)
		}
	}

	tasks := s.taskManager.ListTasks(statusFilter)
	return map[string]interface{}{"tasks": tasks}, nil
}

// =============================================================================
// RESPONSE HELPERS
// =============================================================================

// writeSuccess writes a successful JSON-RPC response
func (s *Server) writeSuccess(w http.ResponseWriter, id interface{}, result interface{}) {
	response := A2AResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// writeError writes an error JSON-RPC response
func (s *Server) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	response := A2AResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &A2AError{Code: code, Message: message},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// CONVENIENCE METHODS
// =============================================================================

// GetAgentCard returns the server's agent card
func (s *Server) GetAgentCard() AgentCard {
	return s.agentCard
}

// GetTaskManager returns the task manager for direct access
func (s *Server) GetTaskManager() *TaskManager {
	return s.taskManager
}

// GetSkillIDs returns the list of available skill IDs
func (s *Server) GetSkillIDs() []string {
	ids := make([]string, len(s.agentCard.Skills))
	for i, skill := range s.agentCard.Skills {
		ids[i] = skill.ID
	}
	return ids
}

// SkillIDFromToolName converts an MCP tool name to A2A skill ID
func SkillIDFromToolName(toolName string) string {
	return strings.ReplaceAll(toolName, "_", "-")
}

// ToolNameFromSkillID converts an A2A skill ID to MCP tool name
func ToolNameFromSkillID(skillID string) string {
	return strings.ReplaceAll(skillID, "-", "_")
}
