# A2A Learning Guide

This guide helps you learn A2A (Agent-to-Agent) protocol step by step using this repository.

## Learning Path

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         A2A Learning Path                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Phase 1: Understand the Basics (30 min)                               │
│  ├── Read: What is A2A vs MCP?                                         │
│  ├── Run: Start the A2A server                                         │
│  └── Test: Use curl to explore the API                                 │
│                                                                         │
│  Phase 2: Explore the Code (1 hour)                                    │
│  ├── Study: pkg/a2a/types.go (A2A data model)                         │
│  ├── Study: pkg/a2a/converter.go (MCP → A2A)                          │
│  └── Study: pkg/a2a/task_manager.go (stateful tasks)                  │
│                                                                         │
│  Phase 3: Hands-on Practice (1 hour)                                   │
│  ├── Build: Run the multi-agent demo                                   │
│  ├── Modify: Add a new skill                                           │
│  └── Create: Write your own A2A client                                 │
│                                                                         │
│  Phase 4: Advanced Topics (ongoing)                                    │
│  ├── Streaming responses                                               │
│  ├── Authentication                                                    │
│  └── Building real multi-agent systems                                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Understand the Basics

### 1.1 What is A2A?

**A2A (Agent-to-Agent)** is Google's protocol for AI agents to communicate with each other. Think of it as HTTP for AI agents.

| Protocol | Purpose | Creator |
|----------|---------|---------|
| MCP | AI Assistant → Tools/Data | Anthropic |
| A2A | AI Agent → AI Agent | Google |

### 1.2 MCP vs A2A Comparison

```
MCP (Current Setup):
┌────────────┐    stdio    ┌───────────────┐
│   Cursor   │◄───────────►│  MCP Server   │
│ (AI Asst)  │   JSON-RPC  │   (tools)     │
└────────────┘             └───────────────┘
     │
     └── Single request/response
     └── Stateless
     └── Local process


A2A (New Capability):
┌────────────┐    HTTP     ┌───────────────┐
│  Agent A   │◄───────────►│   Agent B     │
│(orchestr.) │   JSON-RPC  │   (worker)    │
└────────────┘             └───────────────┘
     │
     └── Task-based (stateful)
     └── Long-running operations
     └── Network communication
```

### 1.3 Quick Start: Run the A2A Server

```bash
# Build both servers
make build build-a2a

# Start the A2A server
make run-a2a

# In another terminal, test it:
curl http://localhost:8080/.well-known/agent.json | jq
```

### 1.4 Explore with curl

**Get Agent Card (Discovery):**
```bash
curl http://localhost:8080/.well-known/agent.json | jq
```

This returns the "business card" of the agent - what it can do.

**Check Health:**
```bash
curl http://localhost:8080/health | jq
```

**Create a Task:**
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tasks/create",
    "params": {
      "skill": "list-contexts",
      "input": {}
    }
  }' | jq
```

Note the `task.id` in the response - you'll need it to check status.

**Get Task Status:**
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tasks/get",
    "params": {
      "taskId": "YOUR_TASK_ID_HERE"
    }
  }' | jq
```

---

## Phase 2: Explore the Code

### 2.1 A2A Types (`pkg/a2a/types.go`)

The core types that define A2A:

```go
// AgentCard - How agents discover each other
type AgentCard struct {
    Name        string      // "dm-nkp-gitops-agent"
    Description string      // What this agent does
    Skills      []Skill     // What it can do
    URL         string      // Where to reach it
}

// Skill - What an agent can do (like MCP Tool)
type Skill struct {
    ID          string      // "get-gitops-status"
    Name        string      // "Get GitOps Status"
    Description string      // Detailed description
    InputSchema InputSchema // Parameters
}

// Task - A unit of work (KEY DIFFERENCE from MCP)
type Task struct {
    ID        string      // Unique identifier
    Skill     string      // Which skill to run
    Status    TaskStatus  // pending → running → completed/failed
    Messages  []Message   // Conversation within task
    Artifacts []Artifact  // Output files/data
}
```

**Key Insight:** In MCP, you call a tool and get a result immediately. In A2A, you create a Task that has a lifecycle.

### 2.2 MCP to A2A Converter (`pkg/a2a/converter.go`)

This bridges your existing MCP tools to A2A:

```go
// MCP Tool (what you have)
mcpTool := mcp.Tool{
    Name:        "get_gitops_status",  // snake_case
    Description: "Get GitOps status",
    InputSchema: ...
}

// A2A Skill (what it becomes)
a2aSkill := Skill{
    ID:          "get-gitops-status",  // kebab-case
    Name:        "Get Gitops Status",  // Title Case
    Description: "Get GitOps status",
    Tags:        []string{"gitops", "flux"},
}
```

### 2.3 Task Manager (`pkg/a2a/task_manager.go`)

This is the brain of A2A - managing task lifecycle:

```
Task Lifecycle:
                    ┌─────────┐
          create    │         │
       ─────────────► pending │
                    │         │
                    └────┬────┘
                         │ execute
                         ▼
                    ┌─────────┐
                    │         │
                    │ running │
                    │         │
                    └────┬────┘
                         │
            ┌────────────┼────────────┐
            │ complete   │ fail       │ cancel
            ▼            ▼            ▼
       ┌─────────┐  ┌─────────┐  ┌─────────┐
       │completed│  │ failed  │  │cancelled│
       └─────────┘  └─────────┘  └─────────┘
```

**Why Stateful Tasks?**
- Long-running operations (deployments, migrations)
- Progress tracking
- Cancellation support
- Audit trail

### 2.4 A2A Server (`pkg/a2a/server.go`)

Compare with MCP server:

| Aspect | MCP Server | A2A Server |
|--------|------------|------------|
| Transport | stdio (stdin/stdout) | HTTP |
| Entry Point | Read from stdin | HTTP handler |
| Discovery | `initialize` response | `/.well-known/agent.json` |
| Tool/Skill Call | `tools/call` (sync) | `tasks/create` (async) |

---

## Phase 3: Hands-on Practice

### 3.1 Run the Multi-Agent Demo

```bash
# Terminal 1: Start A2A server
make run-a2a

# Terminal 2: Run orchestrator demo
go run examples/multi-agent/orchestrator/main.go
```

Watch how the orchestrator:
1. Discovers the agent's capabilities
2. Creates tasks to execute skills
3. Polls for completion
4. Synthesizes results

### 3.2 Add a New Skill

**Exercise:** Add a "health-check" skill that's A2A-only (not from MCP).

1. Add to `pkg/a2a/server.go`:
```go
// In NewServer(), add the skill to agentCard.Skills
agentCard.Skills = append(agentCard.Skills, Skill{
    ID:          "health-check",
    Name:        "Health Check",
    Description: "Check overall system health",
    InputSchema: InputSchema{Type: "object"},
})
```

2. Register the handler in task_manager:
```go
taskManager.RegisterSkillHandler("health-check", func(ctx context.Context, args map[string]interface{}) (*mcp.ToolCallResult, error) {
    return &mcp.ToolCallResult{
        Content: []mcp.Content{{
            Type: "text",
            Text: "System healthy!",
        }},
    }, nil
})
```

### 3.3 Write Your Own A2A Client

**Exercise:** Create a simple client that calls the GitOps agent.

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/a2a"
)

func main() {
    client := a2a.NewClient("http://localhost:8080")
    ctx := context.Background()
    
    // Discover
    card, _ := client.GetAgentCard(ctx)
    fmt.Printf("Agent: %s with %d skills\n", card.Name, len(card.Skills))
    
    // Execute
    result, _ := client.ExecuteSkill(ctx, "list-contexts", nil, 30*time.Second)
    fmt.Printf("Result: %s\n", result.Status)
    
    for _, msg := range result.Messages {
        for _, content := range msg.Content {
            fmt.Println(content.Text)
        }
    }
}
```

---

## Phase 4: Advanced Topics

### 4.1 Streaming Responses

A2A supports streaming for long operations. Implementation guide:

```go
// Server-side: Send progress updates
func (tm *TaskManager) executeWithProgress(task *Task) {
    // Send progress notifications
    tm.sendProgress(task.ID, 0.25, "Loading resources...")
    // ... work ...
    tm.sendProgress(task.ID, 0.50, "Processing...")
    // ... more work ...
    tm.sendProgress(task.ID, 0.75, "Finalizing...")
}
```

### 4.2 Authentication

Add authentication to your A2A server:

```go
// In server.go, add auth middleware
func (s *Server) authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !validateToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Update AgentCard
card.Authentication = &AuthenticationInfo{
    Type:     "bearer",
    Required: true,
}
```

### 4.3 Building Real Multi-Agent Systems

Design patterns for multi-agent architectures:

**Pattern 1: Orchestrator**
```
         ┌─────────────┐
         │ Orchestrator│
         └──────┬──────┘
      ┌─────────┼─────────┐
      ▼         ▼         ▼
   Agent A   Agent B   Agent C
```

**Pattern 2: Pipeline**
```
   Agent A ──► Agent B ──► Agent C
   (gather)    (process)   (output)
```

**Pattern 3: Consensus**
```
         ┌─────────────┐
         │   Agents    │
         │  vote/agree │
         └─────────────┘
```

---

## Multi-Agent Patterns: Deep Dive

This section provides detailed explanations, code examples, and guidance for each multi-agent pattern.

### Pattern 1: Orchestrator Pattern

**Architecture:**
```
                    ┌─────────────────┐
                    │  Orchestrator   │
                    │   (Coordinator) │
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
    ┌──────────┐      ┌──────────┐      ┌──────────┐
    │ GitOps   │      │  Policy  │      │  Alert   │
    │  Agent   │      │  Agent   │      │  Agent   │
    └──────────┘      └──────────┘      └──────────┘
```

**How It Works:**
- A central **Orchestrator** agent coordinates multiple specialized agents
- Orchestrator discovers agents via Agent Cards
- Orchestrator creates tasks on multiple agents (parallel or sequential)
- Orchestrator synthesizes results from all agents
- Each agent is independent and can be replaced/upgraded separately

**Key Characteristics:**
- ✅ **Centralized Control**: One orchestrator makes all decisions
- ✅ **Parallel Execution**: Can run multiple agents simultaneously
- ✅ **Result Synthesis**: Combines outputs from multiple agents
- ✅ **Agent Independence**: Agents don't know about each other
- ✅ **Easy to Scale**: Add new agents without changing existing ones

**When to Use:**
- **Incident Investigation**: Need to check GitOps status, policies, and alerts simultaneously
- **Health Checks**: Coordinating checks across multiple domains (infrastructure, security, compliance)
- **Deployment Validation**: Verifying deployment across multiple systems
- **Multi-Domain Queries**: When you need information from unrelated agents

**Example Use Case:**
```
User: "Why is my application failing?"

Orchestrator:
  1. GitOps Agent → "Check Kustomization status"
  2. Policy Agent → "Check for policy violations"
  3. Cluster Agent → "Check cluster health"
  4. Synthesize: "Kustomization failed due to missing secret, 
                  no policy violations, cluster is healthy"
```

**Code Example:**
See [`examples/multi-agent/orchestrator/main.go`](../../examples/multi-agent/orchestrator/main.go) for a complete working example.

**Quick Start:**
```bash
# Terminal 1: Start GitOps agent
make run-a2a

# Terminal 2: Start Policy agent (optional, on different port)
./bin/dm-nkp-gitops-a2a-server serve --port 8081

# Terminal 3: Run orchestrator
go run examples/multi-agent/orchestrator/main.go
```

---

### Pattern 2: Pipeline Pattern

**Architecture:**
```
    ┌──────────┐      ┌──────────┐      ┌──────────┐
    │  Agent A │ ───► │  Agent B │ ───► │  Agent C │
    │ (gather) │      │(process) │      │ (output) │
    └──────────┘      └──────────┘      └──────────┘
         │                 │                 │
         └─────────────────┴─────────────────┘
                           │
                    ┌──────▼──────┐
                    │   Result    │
                    └─────────────┘
```

**How It Works:**
- Agents are chained together in a **sequential pipeline**
- Each agent receives output from the previous agent as input
- Data flows through the pipeline: `A → B → C`
- Each stage transforms/enriches the data
- Final agent produces the end result

**Key Characteristics:**
- ✅ **Sequential Processing**: Each stage depends on the previous
- ✅ **Data Transformation**: Each agent transforms the data
- ✅ **Clear Data Flow**: Easy to trace data through the pipeline
- ✅ **Modular**: Can add/remove stages easily
- ⚠️ **Sequential Bottleneck**: Can't parallelize stages

**When to Use:**
- **Data Processing Pipelines**: Collect → Transform → Analyze → Report
- **Multi-Stage Analysis**: Gather data, then analyze, then generate report
- **Workflow Automation**: Steps that must happen in order
- **ETL-like Operations**: Extract → Transform → Load

**Example Use Case:**
```
Pipeline: Incident Report Generation

1. Data Collector Agent → Gathers GitOps status, events, logs
2. Analyzer Agent → Analyzes patterns, identifies root causes
3. Report Generator Agent → Creates formatted incident report
```

**Code Example:**
See [`examples/multi-agent/pipeline/main.go`](../../examples/multi-agent/pipeline/main.go) for a complete working example.

**Quick Start:**
```bash
# Start the GitOps agent
make run-a2a

# Run pipeline example
go run examples/multi-agent/pipeline/main.go
```

---

### Pattern 3: Consensus Pattern

**Architecture:**
```
                    ┌─────────────┐
                    │   Request   │
                    └──────┬──────┘
                           │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
          ▼                 ▼                 ▼
    ┌──────────┐      ┌──────────┐      ┌──────────┐
    │  Agent A │      │  Agent B │      │  Agent C │
    │ (expert) │      │ (expert) │      │ (expert) │
    └─────┬────┘      └─────┬────┘      └─────┬────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Consensus        │
                    │   (Voting/Agg)    │
                    └────────────────────┘
```

**How It Works:**
- Multiple **expert agents** independently analyze the same problem
- Each agent provides its own answer/recommendation
- A **consensus mechanism** aggregates the results:
  - **Voting**: Majority wins
  - **Weighted Average**: Based on agent confidence
  - **Best Answer**: Highest confidence wins
- Final decision is based on collective intelligence

**Key Characteristics:**
- ✅ **Redundancy**: Multiple agents verify the same thing
- ✅ **Reliability**: Reduces single point of failure
- ✅ **Confidence Scoring**: Can weight agents by expertise
- ✅ **Disagreement Detection**: Identifies when agents disagree
- ⚠️ **Resource Intensive**: Multiple agents doing similar work

**When to Use:**
- **Critical Decisions**: When you need high confidence (deployment approvals)
- **Disagreement Resolution**: When agents might have different opinions
- **Quality Assurance**: Multiple agents verify the same result
- **Expert Consultation**: Like asking multiple experts for their opinion

**Example Use Case:**
```
Question: "Should we deploy this change?"

Consensus Pattern:
  1. Security Agent → "No violations found" (confidence: 0.9)
  2. GitOps Agent → "All checks passed" (confidence: 0.95)
  3. Performance Agent → "No regressions" (confidence: 0.85)
  
Consensus: Weighted average = 0.90 → APPROVE
```

**Code Example:**
See [`examples/multi-agent/consensus/main.go`](../../examples/multi-agent/consensus/main.go) for a complete working example.

**Quick Start:**
```bash
# Start multiple agent instances (simulating different experts)
make run-a2a                    # Port 8080
./bin/dm-nkp-gitops-a2a-server serve --port 8081  # Port 8081
./bin/dm-nkp-gitops-a2a-server serve --port 8082  # Port 8082

# Run consensus example
go run examples/multi-agent/consensus/main.go
```

---

## Pattern Comparison & Decision Guide

| Aspect | Orchestrator | Pipeline | Consensus |
|--------|-------------|----------|-----------|
| **Execution** | Parallel or Sequential | Sequential | Parallel |
| **Data Flow** | Independent → Synthesized | Transformed through stages | Independent → Aggregated |
| **Use Case** | Multi-domain queries | Data processing | Critical decisions |
| **Complexity** | Medium | Low | High |
| **Latency** | Fast (parallel) | Slower (sequential) | Medium (parallel) |
| **Resource Usage** | Medium | Low | High |
| **Failure Handling** | Continue with available agents | Stop pipeline | Use remaining agents |

### Decision Tree

```
Do you need information from multiple unrelated domains?
├─ YES → Use Orchestrator Pattern
│
└─ NO → Is this a sequential data transformation?
    ├─ YES → Use Pipeline Pattern
    │
    └─ NO → Do you need high confidence/redundancy?
        ├─ YES → Use Consensus Pattern
        │
        └─ NO → Use single agent (not multi-agent)
```

### Combining Patterns

You can combine patterns for complex workflows:

**Example: Hybrid Orchestrator + Pipeline**
```
Orchestrator
    │
    ├─► Pipeline A: Data → Analysis → Report
    │
    └─► Pipeline B: Collect → Validate → Store
```

**Example: Consensus + Orchestrator**
```
Orchestrator coordinates:
    ├─► Consensus Group 1: Security experts
    └─► Consensus Group 2: Performance experts
```

---

## Learning Path for Multi-Agent Patterns

1. **Start with Orchestrator** (easiest to understand)
   - Run: `go run examples/multi-agent/orchestrator/main.go`
   - Study: How it discovers and coordinates agents
   - Modify: Add a third agent to the workflow

2. **Try Pipeline** (understand data flow)
   - Run: `go run examples/multi-agent/pipeline/main.go`
   - Study: How data transforms through stages
   - Modify: Add a new stage to the pipeline

3. **Explore Consensus** (advanced)
   - Run: `go run examples/multi-agent/consensus/main.go`
   - Study: How consensus is calculated
   - Modify: Change the consensus algorithm (voting vs weighted)

4. **Build Your Own**
   - Combine patterns for your use case
   - Add error handling and retries
   - Implement progress tracking

---

## Exercises

### Beginner
1. [ ] Start the A2A server and get the agent card
2. [ ] Create a task using curl
3. [ ] Poll for task completion

### Intermediate
4. [ ] Run the multi-agent demo
5. [ ] Add a custom skill
6. [ ] Write a simple A2A client

### Advanced
7. [ ] Add authentication
8. [ ] Implement progress notifications
9. [ ] Build a multi-agent workflow

---

## Quick Reference

### A2A Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/.well-known/agent.json` | GET | Agent discovery |
| `/health` | GET | Health check |
| `/` | POST | JSON-RPC endpoint |

### JSON-RPC Methods

| Method | Purpose |
|--------|---------|
| `agent/info` | Get agent info |
| `tasks/create` | Create new task |
| `tasks/get` | Get task status |
| `tasks/cancel` | Cancel task |
| `tasks/list` | List all tasks |
| `tasks/message` | Send message to task |

### Task Statuses

| Status | Meaning |
|--------|---------|
| `pending` | Created, not started |
| `running` | Currently executing |
| `completed` | Finished successfully |
| `failed` | Finished with error |
| `cancelled` | User/system cancelled |

---

## Resources

- [A2A Protocol Spec](https://google.github.io/A2A/)
- [A2A GitHub](https://github.com/google/A2A)
- [MCP Protocol](https://spec.modelcontextprotocol.io)
- [Kagent (CNCF)](https://kagent.dev)

---

## Quick Reference: Multi-Agent Patterns

| Pattern | Example | When to Use | Quick Start |
|---------|---------|-------------|-------------|
| **Orchestrator** | [`examples/multi-agent/orchestrator/main.go`](../../examples/multi-agent/orchestrator/main.go) | Multi-domain queries, incident investigation | `make run-a2a && go run examples/multi-agent/orchestrator/main.go` |
| **Pipeline** | [`examples/multi-agent/pipeline/main.go`](../../examples/multi-agent/pipeline/main.go) | Sequential data transformation | `make run-a2a && go run examples/multi-agent/pipeline/main.go` |
| **Consensus** | [`examples/multi-agent/consensus/main.go`](../../examples/multi-agent/consensus/main.go) | Critical decisions, expert consultation | Start 3 agents on ports 8080-8082, then run the example |

## What's Next?

After completing this guide, you can:

1. **Run all three pattern examples** - See [`examples/multi-agent/`](../../examples/multi-agent/) for working code
2. **Extend the GitOps agent** with more sophisticated skills
3. **Build specialized agents** (security, monitoring, cost)
4. **Create multi-agent workflows** combining multiple patterns
5. **Integrate with other A2A agents** in the ecosystem

### Recommended Learning Order

1. ✅ **Phase 1-3**: Understand A2A basics and run the orchestrator example
2. ✅ **Pattern Deep Dive**: Read the detailed pattern explanations above
3. ✅ **Run Examples**: Execute all three pattern examples
4. ✅ **Modify Examples**: Add your own agents or stages
5. ✅ **Build Your Own**: Create a multi-agent system for your use case

Happy learning! 🚀
