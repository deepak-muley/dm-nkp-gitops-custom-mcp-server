# A2A (Agent-to-Agent) Protocol

## Overview

**A2A (Agent-to-Agent)** is Google's open protocol for AI agents to communicate and collaborate with each other. While MCP (Model Context Protocol) focuses on connecting AI assistants to tools and data sources, A2A addresses the need for AI agents to work together on complex tasks.

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Agent Ecosystem                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────┐         A2A          ┌───────────┐              │
│  │  Agent A  │◄────────────────────►│  Agent B  │              │
│  │ (Planning)│    collaboration     │(Execution)│              │
│  └─────┬─────┘                      └─────┬─────┘              │
│        │                                  │                     │
│        │ MCP                              │ MCP                 │
│        ▼                                  ▼                     │
│  ┌───────────┐                      ┌───────────┐              │
│  │MCP Server │                      │MCP Server │              │
│  │  (Tools)  │                      │  (Data)   │              │
│  └───────────┘                      └───────────┘              │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## MCP vs A2A: Key Differences

| Aspect | MCP (Model Context Protocol) | A2A (Agent-to-Agent) |
|--------|------------------------------|---------------------|
| **Purpose** | Connect AI to tools/data | Connect AI agents to each other |
| **Communication** | Human/AI → Tool | AI → AI |
| **Primary Use** | Tool invocation, data access | Task delegation, collaboration |
| **Transport** | stdio, HTTP/SSE | HTTP with JSON-RPC |
| **Direction** | Client-Server | Peer-to-peer (client-server roles flexible) |
| **Discovery** | Static configuration | Dynamic agent discovery |
| **State** | Mostly stateless | Stateful (task tracking) |
| **Creator** | Anthropic | Google |

## A2A Architecture

### Core Concepts

#### 1. Agent Card (Discovery)

Agents expose their capabilities via an "Agent Card" at `/.well-known/agent.json`:

```json
{
  "name": "GitOps Monitor Agent",
  "description": "Monitors and troubleshoots Flux GitOps deployments",
  "version": "1.0.0",
  "url": "https://gitops-agent.example.com",
  "capabilities": {
    "streaming": true,
    "pushNotifications": true
  },
  "skills": [
    {
      "id": "check-gitops-status",
      "name": "Check GitOps Status",
      "description": "Get overall health of GitOps resources",
      "inputSchema": {
        "type": "object",
        "properties": {
          "namespace": {"type": "string"}
        }
      }
    },
    {
      "id": "debug-reconciliation",
      "name": "Debug Reconciliation Failure",
      "description": "Analyze why a Flux resource is failing",
      "inputSchema": {
        "type": "object",
        "properties": {
          "resourceType": {"type": "string"},
          "name": {"type": "string"},
          "namespace": {"type": "string"}
        },
        "required": ["resourceType", "name", "namespace"]
      }
    }
  ],
  "authentication": {
    "type": "bearer",
    "required": true
  }
}
```

#### 2. Tasks

A2A uses a **Task** abstraction for multi-step, potentially long-running operations:

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/create",
  "params": {
    "id": "task-123",
    "skill": "debug-reconciliation",
    "input": {
      "resourceType": "kustomization",
      "name": "flux-system",
      "namespace": "flux-system"
    },
    "metadata": {
      "requestingAgent": "orchestrator-agent",
      "priority": "high"
    }
  }
}
```

Task Lifecycle:
```
                    ┌─────────┐
          create    │         │
       ─────────────► pending │
                    │         │
                    └────┬────┘
                         │ start
                         ▼
                    ┌─────────┐
                    │         │
                    │ running │◄──────┐
                    │         │       │ update
                    └────┬────┘───────┘
                         │
            ┌────────────┼────────────┐
            │ complete   │ fail       │ cancel
            ▼            ▼            ▼
       ┌─────────┐  ┌─────────┐  ┌─────────┐
       │completed│  │ failed  │  │cancelled│
       └─────────┘  └─────────┘  └─────────┘
```

#### 3. Messages (Chat-like Communication)

Agents can exchange messages within a task:

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/message",
  "params": {
    "taskId": "task-123",
    "message": {
      "role": "agent",
      "content": [
        {
          "type": "text",
          "text": "I found 3 failing Kustomizations. Should I generate remediation steps?"
        }
      ]
    }
  }
}
```

#### 4. Artifacts

Agents can produce structured outputs:

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/artifact",
  "params": {
    "taskId": "task-123",
    "artifact": {
      "name": "diagnostic-report",
      "mimeType": "application/json",
      "data": {
        "summary": "Reconciliation failing due to missing secret",
        "affectedResources": [...],
        "recommendations": [...]
      }
    }
  }
}
```

## A2A Protocol Messages

### Core Methods

| Method | Description |
|--------|-------------|
| `tasks/create` | Create a new task |
| `tasks/get` | Get task status and results |
| `tasks/cancel` | Cancel a running task |
| `tasks/message` | Send a message to a task |
| `tasks/artifact` | Produce an artifact |
| `agent/info` | Get agent capabilities |

### Example Flow

```
Orchestrator Agent                    GitOps Agent
       │                                   │
       │  POST /.well-known/agent.json     │
       │──────────────────────────────────►│
       │                                   │
       │  200 OK (Agent Card)              │
       │◄──────────────────────────────────│
       │                                   │
       │  tasks/create (debug-reconciliation)
       │──────────────────────────────────►│
       │                                   │
       │  200 OK {taskId: "t-123", status: "pending"}
       │◄──────────────────────────────────│
       │                                   │
       │         ... agent works ...       │
       │                                   │
       │  tasks/message (progress update)  │
       │◄──────────────────────────────────│
       │                                   │
       │  tasks/artifact (report)          │
       │◄──────────────────────────────────│
       │                                   │
       │  tasks/get (status: completed)    │
       │──────────────────────────────────►│
       │                                   │
```

## How A2A Relates to This MCP Server

### Current: MCP-Only Architecture

```
┌─────────────────┐      MCP       ┌─────────────────────┐
│  AI Assistant   │◄──────────────►│ dm-nkp-gitops-mcp   │
│  (Cursor)       │   stdio        │ server              │
└─────────────────┘                └──────────┬──────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │   Kubernetes API    │
                                   └─────────────────────┘
```

### Future: A2A + MCP Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Multi-Agent System                        │
│                                                             │
│  ┌───────────────┐      A2A       ┌───────────────┐        │
│  │  Orchestrator │◄──────────────►│   GitOps      │        │
│  │  Agent        │                │   Agent       │        │
│  └───────┬───────┘                └───────┬───────┘        │
│          │                                │                 │
│          │ A2A                            │ MCP             │
│          ▼                                ▼                 │
│  ┌───────────────┐                ┌───────────────┐        │
│  │  Security     │                │ dm-nkp-gitops │        │
│  │  Agent        │                │ mcp-server    │        │
│  └───────┬───────┘                └───────┬───────┘        │
│          │ MCP                            │                 │
│          ▼                                ▼                 │
│  ┌───────────────┐                ┌───────────────┐        │
│  │ Policy MCP    │                │ Kubernetes    │        │
│  │ Server        │                │ API           │        │
│  └───────────────┘                └───────────────┘        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Converting MCP Tools to A2A Skills

MCP Tool → A2A Skill mapping:

```go
// MCP Tool (current)
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"inputSchema"`
}

// A2A Skill (future)
type Skill struct {
    ID          string      `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"inputSchema"`
    Tags        []string    `json:"tags,omitempty"`
    Examples    []Example   `json:"examples,omitempty"`
}
```

Example conversion:

```go
// MCP
mcpTool := mcp.Tool{
    Name:        "get_gitops_status",
    Description: "Get overall GitOps status",
    InputSchema: mcp.InputSchema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "namespace": {Type: "string", Description: "Filter by namespace"},
        },
    },
}

// A2A
a2aSkill := a2a.Skill{
    ID:          "get-gitops-status",
    Name:        "Get GitOps Status",
    Description: "Get overall GitOps status including all Flux Kustomizations and GitRepositories",
    InputSchema: a2a.InputSchema{
        Type: "object",
        Properties: map[string]a2a.Property{
            "namespace": {Type: "string", Description: "Filter by namespace"},
        },
    },
    Tags: []string{"gitops", "flux", "monitoring"},
}
```

## Implementing A2A Support

### Step 1: Add HTTP Server Mode

```go
// cmd/server/main.go
func main() {
    switch command {
    case "serve":
        runMCPServer()  // stdio mode (existing)
    case "serve-a2a":
        runA2AServer()  // HTTP mode (new)
    }
}

func runA2AServer() {
    http.HandleFunc("/.well-known/agent.json", handleAgentCard)
    http.HandleFunc("/tasks", handleTasks)
    http.HandleFunc("/tasks/", handleTaskByID)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Step 2: Implement Agent Card Endpoint

```go
func handleAgentCard(w http.ResponseWriter, r *http.Request) {
    card := AgentCard{
        Name:        "dm-nkp-gitops-agent",
        Description: "GitOps monitoring and debugging agent",
        Version:     Version,
        URL:         "http://localhost:8080",
        Skills:      convertToolsToSkills(registry.GetTools()),
    }
    json.NewEncoder(w).Encode(card)
}
```

### Step 3: Implement Task Management

```go
type Task struct {
    ID        string                 `json:"id"`
    Skill     string                 `json:"skill"`
    Input     map[string]interface{} `json:"input"`
    Status    string                 `json:"status"`
    Result    interface{}            `json:"result,omitempty"`
    Messages  []Message              `json:"messages,omitempty"`
    Artifacts []Artifact             `json:"artifacts,omitempty"`
    CreatedAt time.Time              `json:"createdAt"`
    UpdatedAt time.Time              `json:"updatedAt"`
}

var tasks = sync.Map{}

func handleTasks(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "POST":
        // Create new task
        var req TaskCreateRequest
        json.NewDecoder(r.Body).Decode(&req)
        
        task := &Task{
            ID:        uuid.New().String(),
            Skill:     req.Skill,
            Input:     req.Input,
            Status:    "pending",
            CreatedAt: time.Now(),
        }
        tasks.Store(task.ID, task)
        
        // Execute asynchronously
        go executeTask(task)
        
        json.NewEncoder(w).Encode(task)
    }
}
```

## A2A vs Other Agent Protocols

| Protocol | Creator | Focus | Transport | Status |
|----------|---------|-------|-----------|--------|
| **A2A** | Google | Agent-to-agent | HTTP/JSON-RPC | Active (2024) |
| **MCP** | Anthropic | AI-to-tools | stdio, HTTP/SSE | Active (2024) |
| **AutoGen** | Microsoft | Multi-agent framework | Python library | Active |
| **LangGraph** | LangChain | Agent orchestration | Python library | Active |
| **CrewAI** | CrewAI | Agent crews | Python library | Active |

## Benefits of A2A for GitOps Monitoring

1. **Specialized Agents**: Have dedicated agents for Flux, CAPI, Policy, etc.
2. **Parallel Execution**: Multiple agents can work on different aspects simultaneously
3. **Expertise Isolation**: Each agent maintains deep expertise in its domain
4. **Composability**: Build complex workflows from simple agent interactions
5. **Resilience**: Failure in one agent doesn't crash the entire system

## Example: Multi-Agent GitOps Workflow

```
User: "Why is my application not deploying?"

Orchestrator Agent:
  │
  ├─► GitOps Agent: "Check Kustomization status"
  │   └─► Returns: "flux-system Kustomization failed"
  │
  ├─► GitOps Agent: "Debug reconciliation failure"
  │   └─► Returns: "Source not found: missing GitRepository"
  │
  ├─► Cluster Agent: "Check cluster health"
  │   └─► Returns: "Cluster healthy, 3/3 nodes ready"
  │
  └─► Orchestrator synthesizes response:
      "Your application isn't deploying because the GitRepository
       source is missing. The cluster is healthy. Recommend
       creating the GitRepository resource."
```

## Resources

- [A2A Protocol Specification](https://google.github.io/A2A/)
- [A2A GitHub Repository](https://github.com/google/A2A)
- [MCP Specification](https://spec.modelcontextprotocol.io)
- [Kagent (CNCF)](https://kagent.dev) - Kubernetes-native AI agents

## Summary

| When to Use | MCP | A2A |
|-------------|-----|-----|
| Single AI assistant needs tools | ✅ | ❌ |
| Tool/data integration | ✅ | ❌ |
| Simple request-response | ✅ | ✅ |
| Multi-agent collaboration | ❌ | ✅ |
| Long-running tasks | ❌ | ✅ |
| Dynamic agent discovery | ❌ | ✅ |
| Peer-to-peer communication | ❌ | ✅ |

**Recommendation**: Use MCP for tool integration (current use case), consider A2A when building multi-agent systems for complex GitOps automation.
