package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// Logger interface for logging.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// ServerConfig contains configuration for the MCP server.
type ServerConfig struct {
	Name        string
	Version     string
	Description string
	Tools       []Tool
	Handlers    map[string]ToolHandler
	Logger      Logger
}

// Server is an MCP server that communicates via stdio.
type Server struct {
	config       ServerConfig
	initialized  bool
	mu           sync.Mutex
	reader       *bufio.Reader
	writer       *bufio.Writer
	instructions string
}

// NewServer creates a new MCP server.
func NewServer(config ServerConfig) *Server {
	instructions := fmt.Sprintf(`This is the %s MCP server for monitoring and debugging NKP GitOps infrastructure.

Available capabilities:
- Query Flux Kustomization and GitRepository status
- Check CAPI cluster health and status
- Get application deployment status across workspaces
- Debug reconciliation failures
- Check policy violations (Gatekeeper/Kyverno)
- Compare configurations across clusters

Key namespaces in this infrastructure:
- dm-nkp-gitops-infra: Management cluster GitOps resources
- dm-nkp-gitops-workload: Workload cluster GitOps resources
- kommander: Kommander management resources
- dm-dev-workspace: Development workspace

Important clusters:
- dm-nkp-mgmt-1: Management cluster
- dm-nkp-workload-1: Workload cluster 1
- dm-nkp-workload-2: Workload cluster 2

When debugging issues:
1. First check the Kustomization status
2. Then check events for the affected resources
3. Look at pod logs if needed
4. Check for policy violations`, config.Name)

	return &Server{
		config:       config,
		reader:       bufio.NewReader(os.Stdin),
		writer:       bufio.NewWriter(os.Stdout),
		instructions: instructions,
	}
}

// Run starts the MCP server and processes messages until EOF.
func (s *Server) Run() error {
	s.config.Logger.Info("MCP server started, waiting for messages")

	for {
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.config.Logger.Info("EOF received, shutting down")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		// Skip empty lines
		if len(line) <= 1 {
			continue
		}

		response, err := s.handleMessage(line)
		if err != nil {
			s.config.Logger.Error("Error handling message", "error", err)
			// Send error response
			response = s.errorResponse(nil, InternalError, err.Error())
		}

		if response != nil {
			if err := s.writeResponse(response); err != nil {
				s.config.Logger.Error("Error writing response", "error", err)
			}
		}
	}
}

// handleMessage processes a single JSON-RPC message.
func (s *Server) handleMessage(data []byte) ([]byte, error) {
	var request JSONRPCRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return s.errorResponse(nil, ParseError, "Parse error"), nil
	}

	s.config.Logger.Debug("Received request", "method", request.Method, "id", request.ID)

	// Handle notifications (no id)
	if request.ID == nil {
		return s.handleNotification(request)
	}

	// Handle requests
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "initialized":
		// This is a notification, no response needed
		s.mu.Lock()
		s.initialized = true
		s.mu.Unlock()
		return nil, nil
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(request)
	case "resources/list":
		return s.handleResourcesList(request)
	case "prompts/list":
		return s.handlePromptsList(request)
	case "ping":
		return s.handlePing(request)
	default:
		return s.errorResponse(request.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", request.Method)), nil
	}
}

// handleNotification handles notifications (messages without id).
func (s *Server) handleNotification(request JSONRPCRequest) ([]byte, error) {
	switch request.Method {
	case "initialized":
		s.mu.Lock()
		s.initialized = true
		s.mu.Unlock()
		s.config.Logger.Info("Client initialized")
	case "notifications/cancelled":
		s.config.Logger.Debug("Request cancelled")
	default:
		s.config.Logger.Debug("Unknown notification", "method", request.Method)
	}
	return nil, nil
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(request JSONRPCRequest) ([]byte, error) {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Logging: &LoggingCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    s.config.Name,
			Version: s.config.Version,
		},
		Instructions: s.instructions,
	}

	return s.successResponse(request.ID, result), nil
}

// handleToolsList handles the tools/list request.
func (s *Server) handleToolsList(request JSONRPCRequest) ([]byte, error) {
	result := ToolsListResult{
		Tools: s.config.Tools,
	}

	return s.successResponse(request.ID, result), nil
}

// handleToolsCall handles the tools/call request.
func (s *Server) handleToolsCall(request JSONRPCRequest) ([]byte, error) {
	// Parse params
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return s.errorResponse(request.ID, InvalidParams, "Invalid params"), nil
	}

	var params ToolCallParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return s.errorResponse(request.ID, InvalidParams, "Invalid params"), nil
	}

	s.config.Logger.Debug("Tool call", "name", params.Name, "arguments", params.Arguments)

	// Find and execute handler
	handler, ok := s.config.Handlers[params.Name]
	if !ok {
		return s.errorResponse(request.ID, InvalidParams, fmt.Sprintf("Unknown tool: %s", params.Name)), nil
	}

	result, err := handler(params.Arguments)
	if err != nil {
		s.config.Logger.Error("Tool execution error", "tool", params.Name, "error", err)
		// Return error as tool result, not JSON-RPC error
		result = &ToolCallResult{
			Content: []Content{
				{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())},
			},
			IsError: true,
		}
	}

	return s.successResponse(request.ID, result), nil
}

// handleResourcesList handles the resources/list request.
func (s *Server) handleResourcesList(request JSONRPCRequest) ([]byte, error) {
	// Currently no resources implemented
	result := ResourcesListResult{
		Resources: []Resource{},
	}

	return s.successResponse(request.ID, result), nil
}

// handlePromptsList handles the prompts/list request.
func (s *Server) handlePromptsList(request JSONRPCRequest) ([]byte, error) {
	// Currently no prompts implemented
	result := PromptsListResult{
		Prompts: []Prompt{},
	}

	return s.successResponse(request.ID, result), nil
}

// handlePing handles the ping request.
func (s *Server) handlePing(request JSONRPCRequest) ([]byte, error) {
	return s.successResponse(request.ID, map[string]string{}), nil
}

// successResponse creates a successful JSON-RPC response.
func (s *Server) successResponse(id interface{}, result interface{}) []byte {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, _ := json.Marshal(response)
	return data
}

// errorResponse creates an error JSON-RPC response.
func (s *Server) errorResponse(id interface{}, code int, message string) []byte {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}

	data, _ := json.Marshal(response)
	return data
}

// writeResponse writes a response to stdout.
func (s *Server) writeResponse(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.writer.Write(data); err != nil {
		return err
	}
	if err := s.writer.WriteByte('\n'); err != nil {
		return err
	}
	return s.writer.Flush()
}
