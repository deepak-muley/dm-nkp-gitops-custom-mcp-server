# MCP Server Architecture & Flow

## Overview

This document explains the architecture and message flow of the dm-nkp-gitops MCP server and recommends standard features for production readiness.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              AI Client (Cursor/Claude)                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ JSON-RPC 2.0 over stdio
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              MCP Server Process                              │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                           pkg/mcp/server.go                            │  │
│  │  ┌─────────────┐    ┌──────────────────┐    ┌───────────────────────┐ │  │
│  │  │   Reader    │───►│  Message Router  │───►│   Response Writer     │ │  │
│  │  │ (stdin)     │    │  handleMessage() │    │   (stdout)            │ │  │
│  │  └─────────────┘    └────────┬─────────┘    └───────────────────────┘ │  │
│  │                              │                                         │  │
│  │         ┌────────────────────┼────────────────────┐                    │  │
│  │         ▼                    ▼                    ▼                    │  │
│  │  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐             │  │
│  │  │ initialize  │    │ tools/list   │    │ tools/call   │             │  │
│  │  └─────────────┘    └──────────────┘    └──────┬───────┘             │  │
│  └───────────────────────────────────────────────┼───────────────────────┘  │
│                                                   │                          │
│  ┌───────────────────────────────────────────────┼───────────────────────┐  │
│  │                    pkg/tools/registry.go       │                       │  │
│  │                                                ▼                       │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐  │  │
│  │  │                      Tool Handlers                               │  │  │
│  │  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐  │  │  │
│  │  │  │ context_      │ │ flux_         │ │ cluster_handlers.go   │  │  │  │
│  │  │  │ handlers.go   │ │ handlers.go   │ │                       │  │  │  │
│  │  │  └───────────────┘ └───────────────┘ └───────────────────────┘  │  │  │
│  │  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐  │  │  │
│  │  │  │ app_          │ │ debug_        │ │ policy_handlers.go    │  │  │  │
│  │  │  │ handlers.go   │ │ handlers.go   │ │                       │  │  │  │
│  │  │  └───────────────┘ └───────────────┘ └───────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                   │                          │
│  ┌───────────────────────────────────────────────┼───────────────────────┐  │
│  │                    pkg/config/config.go        │                       │  │
│  │                                                ▼                       │  │
│  │  ┌─────────────────────────────────────────────────────────────────┐  │  │
│  │  │                    K8sClients                                    │  │  │
│  │  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐  │  │  │
│  │  │  │ Clientset     │ │ Dynamic       │ │ RestConfig            │  │  │  │
│  │  │  │ (typed API)   │ │ (CRDs)        │ │                       │  │  │  │
│  │  │  └───────────────┘ └───────────────┘ └───────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ Kubernetes API calls
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Kubernetes API Server                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │ Core API     │  │ Flux CRDs    │  │ CAPI CRDs    │  │ Policy CRDs  │    │
│  │ pods, events │  │ kustomize,   │  │ clusters,    │  │ gatekeeper,  │    │
│  │ configmaps   │  │ gitrepo      │  │ machines     │  │ kyverno      │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Message Flow: Step by Step

### 1. Startup Phase

```go
// cmd/server/main.go
func runServer() {
    // 1. Parse configuration
    cfg := config.ParseFlags(os.Args[2:])
    
    // 2. Create logger
    logger := config.NewLogger(cfg.LogLevel)
    
    // 3. Load Kubernetes config
    k8sConfig, _ := config.LoadKubeConfig(cfg.Kubeconfig, cfg.Context)
    
    // 4. Create K8s clients
    clients, _ := config.NewK8sClients(k8sConfig)
    
    // 5. Register all tools
    registry := tools.NewRegistry(clients, cfg.ReadOnly, logger)
    registry.RegisterAllTools()
    
    // 6. Create and run MCP server
    server := mcp.NewServer(mcp.ServerConfig{...})
    server.Run()  // Blocks, reading from stdin
}
```

### 2. Connection Phase (Initialize Handshake)

```
Client                                    Server
   │                                         │
   │  {"jsonrpc":"2.0","id":1,              │
   │   "method":"initialize",                │
   │   "params":{                            │
   │     "protocolVersion":"2024-11-05",     │
   │     "capabilities":{},                  │
   │     "clientInfo":{"name":"Cursor"}      │
   │   }}                                    │
   │────────────────────────────────────────►│
   │                                         │ handleInitialize()
   │                                         │
   │  {"jsonrpc":"2.0","id":1,              │
   │   "result":{                            │
   │     "protocolVersion":"2024-11-05",     │
   │     "capabilities":{"tools":{}},        │
   │     "serverInfo":{"name":"..."},        │
   │     "instructions":"..."                │
   │   }}                                    │
   │◄────────────────────────────────────────│
   │                                         │
   │  {"jsonrpc":"2.0",                      │
   │   "method":"initialized"}               │ (notification, no id)
   │────────────────────────────────────────►│
   │                                         │ handleNotification()
   │                                         │ s.initialized = true
```

### 3. Tool Discovery Phase

```
Client                                    Server
   │                                         │
   │  {"jsonrpc":"2.0","id":2,              │
   │   "method":"tools/list"}                │
   │────────────────────────────────────────►│
   │                                         │ handleToolsList()
   │                                         │
   │  {"jsonrpc":"2.0","id":2,              │
   │   "result":{                            │
   │     "tools":[                           │
   │       {"name":"get_gitops_status",...}, │
   │       {"name":"list_kustomizations",...}│
   │     ]                                   │
   │   }}                                    │
   │◄────────────────────────────────────────│
```

### 4. Tool Execution Phase

```
Client                                    Server
   │                                         │
   │  {"jsonrpc":"2.0","id":3,              │
   │   "method":"tools/call",                │
   │   "params":{                            │
   │     "name":"get_gitops_status",         │
   │     "arguments":{"namespace":"flux..."}│
   │   }}                                    │
   │────────────────────────────────────────►│
   │                                         │ handleToolsCall()
   │                                         │   │
   │                                         │   ├─► Parse params
   │                                         │   ├─► Find handler
   │                                         │   ├─► Execute handler
   │                                         │   │     │
   │                                         │   │     ├─► K8s API call
   │                                         │   │     ├─► Format result
   │                                         │   │     └─► Return
   │                                         │   └─► Build response
   │                                         │
   │  {"jsonrpc":"2.0","id":3,              │
   │   "result":{                            │
   │     "content":[{                        │
   │       "type":"text",                    │
   │       "text":"# GitOps Status..."       │
   │     }]                                  │
   │   }}                                    │
   │◄────────────────────────────────────────│
```

### 5. Server Code Flow

```go
// pkg/mcp/server.go - Main message loop
func (s *Server) Run() error {
    for {
        // 1. Read line from stdin (blocks)
        line, err := s.reader.ReadBytes('\n')
        
        // 2. Handle message
        response, err := s.handleMessage(line)
        
        // 3. Write response to stdout
        if response != nil {
            s.writeResponse(response)
        }
    }
}

// Message routing
func (s *Server) handleMessage(data []byte) ([]byte, error) {
    var request JSONRPCRequest
    json.Unmarshal(data, &request)
    
    switch request.Method {
    case "initialize":
        return s.handleInitialize(request)
    case "tools/list":
        return s.handleToolsList(request)
    case "tools/call":
        return s.handleToolsCall(request)
    // ... other methods
    }
}

// Tool execution
func (s *Server) handleToolsCall(request JSONRPCRequest) ([]byte, error) {
    // 1. Parse tool name and arguments
    var params ToolCallParams
    json.Unmarshal(paramsBytes, &params)
    
    // 2. Find handler function
    handler := s.config.Handlers[params.Name]
    
    // 3. Execute handler
    result, err := handler(params.Arguments)
    
    // 4. Return result
    return s.successResponse(request.ID, result), nil
}
```

---

## Current Features

| Feature | Status | Implementation |
|---------|--------|----------------|
| Tools (read-only) | ✅ | 15 tools implemented |
| Resources | ❌ | Empty list returned |
| Prompts | ❌ | Empty list returned |
| Logging | ✅ | stderr logging |
| Read-only mode | ✅ | CLI flag |
| Instructions | ✅ | Returned on initialize |

---

## Recommended Standard Features to Add

### 1. 🔴 HIGH PRIORITY: Graceful Shutdown

**Why:** Prevents resource leaks and ensures clean termination.

```go
// Add to pkg/mcp/server.go

import (
    "context"
    "os"
    "os/signal"
    "syscall"
)

func (s *Server) RunWithContext(ctx context.Context) error {
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        select {
        case <-sigChan:
            s.config.Logger.Info("Shutdown signal received")
            s.shutdown()
        case <-ctx.Done():
            s.shutdown()
        }
    }()
    
    return s.Run()
}

func (s *Server) shutdown() {
    s.mu.Lock()
    defer s.mu.Unlock()
    // Clean up resources
    s.config.Logger.Info("Server shutting down")
}
```

### 2. 🔴 HIGH PRIORITY: Request Context with Timeout

**Why:** Prevents hanging requests from blocking the server.

```go
// Add to pkg/tools/registry.go

type ToolHandlerWithContext func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error)

// In handleToolsCall
func (s *Server) handleToolsCall(request JSONRPCRequest) ([]byte, error) {
    // Add timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := handler(ctx, params.Arguments)
    // ...
}
```

### 3. 🔴 HIGH PRIORITY: Request Cancellation

**Why:** MCP spec supports cancellation notifications.

```go
// Add to pkg/mcp/server.go

type Server struct {
    // ... existing fields
    pendingRequests map[interface{}]context.CancelFunc
    pendingMu       sync.Mutex
}

func (s *Server) handleNotification(request JSONRPCRequest) ([]byte, error) {
    switch request.Method {
    case "notifications/cancelled":
        var params struct {
            RequestID interface{} `json:"requestId"`
        }
        json.Unmarshal(paramsBytes, &params)
        
        s.pendingMu.Lock()
        if cancel, ok := s.pendingRequests[params.RequestID]; ok {
            cancel()
            delete(s.pendingRequests, params.RequestID)
        }
        s.pendingMu.Unlock()
    }
    return nil, nil
}
```

### 4. 🟡 MEDIUM PRIORITY: Progress Notifications

**Why:** Long-running operations should report progress.

```go
// Add to pkg/mcp/types.go

type ProgressNotification struct {
    JSONRPC        string  `json:"jsonrpc"`
    Method         string  `json:"method"` // "notifications/progress"
    Params         ProgressParams `json:"params"`
}

type ProgressParams struct {
    ProgressToken interface{} `json:"progressToken"`
    Progress      float64     `json:"progress"` // 0.0 to 1.0
    Total         int         `json:"total,omitempty"`
    Message       string      `json:"message,omitempty"`
}

// Usage in handler
func (r *Registry) handleListKustomizations(args map[string]interface{}) (*mcp.ToolCallResult, error) {
    // Send progress
    r.sendProgress(progressToken, 0.5, "Fetching kustomizations...")
    
    // ... do work
    
    r.sendProgress(progressToken, 1.0, "Complete")
    return result, nil
}
```

### 5. 🟡 MEDIUM PRIORITY: Audit Logging

**Why:** Track all operations for security and debugging.

```go
// Add pkg/audit/audit.go

type AuditLogger struct {
    output io.Writer
}

type AuditEvent struct {
    Timestamp  time.Time              `json:"timestamp"`
    RequestID  interface{}            `json:"request_id"`
    Method     string                 `json:"method"`
    Tool       string                 `json:"tool,omitempty"`
    Arguments  map[string]interface{} `json:"arguments,omitempty"`
    Duration   time.Duration          `json:"duration_ms"`
    Success    bool                   `json:"success"`
    Error      string                 `json:"error,omitempty"`
    UserAgent  string                 `json:"user_agent,omitempty"`
}

func (a *AuditLogger) Log(event AuditEvent) {
    data, _ := json.Marshal(event)
    fmt.Fprintln(a.output, string(data))
}
```

### 6. 🟡 MEDIUM PRIORITY: Health Check Tool

**Why:** AI can verify server health before operations.

```go
// Add to pkg/tools/registry.go

func (r *Registry) registerHealthTools() {
    r.register(
        mcp.Tool{
            Name:        "health_check",
            Description: "Check MCP server and Kubernetes connectivity health",
            InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}},
        },
        r.handleHealthCheck,
    )
}

func (r *Registry) handleHealthCheck(args map[string]interface{}) (*mcp.ToolCallResult, error) {
    checks := map[string]bool{
        "k8s_connection": false,
        "flux_crds":      false,
        "capi_crds":      false,
    }
    
    // Check K8s connection
    _, err := r.clients.Clientset.Discovery().ServerVersion()
    checks["k8s_connection"] = err == nil
    
    // Check Flux CRDs exist
    // ... similar checks
    
    return formatHealthResult(checks), nil
}
```

### 7. 🟡 MEDIUM PRIORITY: Input Validation

**Why:** Prevent injection and improve error messages.

```go
// Add pkg/validation/validation.go

package validation

import (
    "fmt"
    "regexp"
)

var (
    namespaceRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
    nameRegex      = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

func ValidateNamespace(ns string) error {
    if ns == "" {
        return nil // Optional
    }
    if len(ns) > 63 {
        return fmt.Errorf("namespace too long (max 63 chars)")
    }
    if !namespaceRegex.MatchString(ns) {
        return fmt.Errorf("invalid namespace format")
    }
    return nil
}

func ValidateResourceName(name string) error {
    if name == "" {
        return fmt.Errorf("name is required")
    }
    if len(name) > 253 {
        return fmt.Errorf("name too long (max 253 chars)")
    }
    if !nameRegex.MatchString(name) {
        return fmt.Errorf("invalid name format")
    }
    return nil
}

func ValidateLimit(limit string, defaultVal, maxVal int) (int, error) {
    if limit == "" {
        return defaultVal, nil
    }
    val, err := strconv.Atoi(limit)
    if err != nil {
        return 0, fmt.Errorf("invalid limit: must be a number")
    }
    if val < 1 || val > maxVal {
        return 0, fmt.Errorf("limit must be between 1 and %d", maxVal)
    }
    return val, nil
}
```

### 8. 🟢 LOW PRIORITY: Resource Subscriptions

**Why:** MCP spec supports subscribing to resource changes.

```go
// Add to pkg/mcp/server.go

func (s *Server) handleResourcesSubscribe(request JSONRPCRequest) ([]byte, error) {
    var params struct {
        URI string `json:"uri"`
    }
    // ... parse params
    
    // Start watching resource
    go s.watchResource(params.URI)
    
    return s.successResponse(request.ID, map[string]bool{"subscribed": true}), nil
}

func (s *Server) watchResource(uri string) {
    // Use K8s watch API
    watcher, _ := r.clients.Dynamic.Resource(gvr).Watch(ctx, metav1.ListOptions{})
    
    for event := range watcher.ResultChan() {
        // Send notification
        s.sendNotification("notifications/resources/updated", map[string]interface{}{
            "uri": uri,
        })
    }
}
```

### 9. 🟢 LOW PRIORITY: Prompts (Pre-defined Queries)

**Why:** Provide common workflows as templates.

```go
// Add to pkg/mcp/server.go and pkg/tools/prompts.go

var defaultPrompts = []Prompt{
    {
        Name:        "debug-failing-kustomization",
        Description: "Debug a failing Flux Kustomization step by step",
        Arguments: []PromptArgument{
            {Name: "name", Description: "Kustomization name", Required: true},
            {Name: "namespace", Description: "Namespace", Required: true},
        },
    },
    {
        Name:        "cluster-health-report",
        Description: "Generate comprehensive cluster health report",
        Arguments: []PromptArgument{
            {Name: "cluster_name", Description: "CAPI cluster name", Required: false},
        },
    },
}

func (s *Server) handlePromptsGet(request JSONRPCRequest) ([]byte, error) {
    // Return prompt messages that guide the AI
    messages := []PromptMessage{
        {
            Role: "user",
            Content: []Content{{
                Type: "text",
                Text: "Please debug the Kustomization...",
            }},
        },
    }
    return s.successResponse(request.ID, PromptGetResult{Messages: messages}), nil
}
```

### 10. 🟢 LOW PRIORITY: Metrics Endpoint

**Why:** Monitor server performance.

```go
// Add pkg/metrics/metrics.go

type Metrics struct {
    RequestsTotal    int64
    RequestsDuration map[string]time.Duration
    ErrorsTotal      int64
    ActiveRequests   int64
    mu               sync.Mutex
}

func (m *Metrics) RecordRequest(method string, duration time.Duration, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.RequestsTotal++
    m.RequestsDuration[method] = duration
    if err != nil {
        m.ErrorsTotal++
    }
}

// Add metrics tool
func (r *Registry) handleGetMetrics(args map[string]interface{}) (*mcp.ToolCallResult, error) {
    return formatMetrics(r.metrics), nil
}
```

---

## Implementation Priority Matrix

| Feature | Priority | Effort | Impact | Recommended |
|---------|----------|--------|--------|-------------|
| Graceful Shutdown | 🔴 High | Low | High | ✅ Yes |
| Request Context/Timeout | 🔴 High | Medium | High | ✅ Yes |
| Request Cancellation | 🔴 High | Medium | Medium | ✅ Yes |
| Input Validation | 🟡 Medium | Low | High | ✅ Yes |
| Audit Logging | 🟡 Medium | Medium | High | ✅ Yes |
| Health Check Tool | 🟡 Medium | Low | Medium | ✅ Yes |
| Progress Notifications | 🟡 Medium | Medium | Medium | Optional |
| Resource Subscriptions | 🟢 Low | High | Low | Optional |
| Prompts | 🟢 Low | Medium | Medium | Optional |
| Metrics | 🟢 Low | Medium | Low | Optional |

---

## Recommended File Structure (After Improvements)

```
pkg/
├── mcp/
│   ├── server.go         # Main server logic
│   ├── types.go          # MCP types
│   ├── handlers.go       # Method handlers (refactored from server.go)
│   └── notifications.go  # Progress/resource notifications
├── config/
│   ├── config.go         # Configuration
│   └── logger.go         # Logging
├── tools/
│   ├── registry.go       # Tool registration
│   ├── *_handlers.go     # Tool implementations
│   └── health.go         # Health check tool
├── validation/
│   └── validation.go     # Input validation
├── audit/
│   └── audit.go          # Audit logging
└── metrics/
    └── metrics.go        # Metrics collection
```

---

## Summary

Your MCP server has a solid foundation. The recommended improvements in order of priority:

1. **Graceful shutdown** - Essential for production
2. **Request timeouts** - Prevent hanging operations
3. **Input validation** - Security and UX
4. **Audit logging** - Security compliance
5. **Health check tool** - Operational visibility

Would you like me to implement any of these features?
