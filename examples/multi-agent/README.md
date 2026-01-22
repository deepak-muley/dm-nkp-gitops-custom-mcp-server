# Multi-Agent Demo

This example demonstrates A2A (Agent-to-Agent) communication between multiple specialized agents.

## Architecture

```
                    ┌─────────────────────┐
                    │   Orchestrator      │
                    │   Agent             │
                    │   (coordinator)     │
                    └──────────┬──────────┘
                               │
              ┌────────────────┼────────────────┐
              │                │                │
              ▼                ▼                ▼
    ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
    │   GitOps Agent  │ │   Policy Agent  │ │   Alert Agent   │
    │   (port 8080)   │ │   (port 8081)   │ │   (port 8082)   │
    └─────────────────┘ └─────────────────┘ └─────────────────┘
           │                    │                    │
           └────────────────────┴────────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │   Kubernetes API    │
                    └─────────────────────┘
```

## What Each Agent Does

| Agent | Port | Purpose |
|-------|------|---------|
| GitOps Agent | 8080 | Monitors Flux Kustomizations, GitRepositories, HelmReleases |
| Policy Agent | 8081 | Checks Gatekeeper/Kyverno policy violations |
| Alert Agent | 8082 | Simulates an alerting agent (demo only) |
| Orchestrator | N/A | Coordinates work across agents |

## Running the Demo

### Step 1: Start the GitOps Agent

```bash
# From repo root
make run-a2a
# Or directly:
./bin/dm-nkp-gitops-a2a-server serve --port 8080
```

### Step 2: Start a Simple Policy Agent (Optional)

For this demo, you can run a second instance on a different port:

```bash
./bin/dm-nkp-gitops-a2a-server serve --port 8081
```

### Step 3: Run the Orchestrator Demo

```bash
go run examples/multi-agent/orchestrator/main.go
```

## Understanding the Demo

### 1. Agent Discovery

The orchestrator first discovers each agent's capabilities:

```go
gitopsClient := a2a.NewClient("http://localhost:8080")
agentCard, _ := gitopsClient.GetAgentCard(ctx)

fmt.Println("Available skills:")
for _, skill := range agentCard.Skills {
    fmt.Printf("  - %s: %s\n", skill.ID, skill.Description)
}
```

### 2. Task-Based Execution

Unlike MCP's direct tool calls, A2A uses tasks:

```go
// Create a task
task, _ := gitopsClient.CreateTask(ctx, "get-gitops-status", map[string]interface{}{
    "namespace": "flux-system",
})

// Task runs asynchronously
fmt.Printf("Task %s status: %s\n", task.ID, task.Status)

// Wait for completion
completedTask, _ := gitopsClient.WaitForTask(ctx, task.ID, 30*time.Second)
```

### 3. Multi-Agent Coordination

The orchestrator can call multiple agents:

```go
// Parallel calls to different agents
var wg sync.WaitGroup

wg.Add(2)
go func() {
    defer wg.Done()
    gitopsStatus, _ := gitopsClient.ExecuteSkill(ctx, "get-gitops-status", nil, 30*time.Second)
}()

go func() {
    defer wg.Done()
    policyStatus, _ := policyClient.ExecuteSkill(ctx, "check-policy-violations", nil, 30*time.Second)
}()

wg.Wait()
```

## Key Learning Points

### MCP vs A2A Comparison

| Aspect | MCP | A2A |
|--------|-----|-----|
| Transport | stdio | HTTP |
| Communication | Human/AI → Tool | Agent → Agent |
| State | Stateless | Stateful (Tasks) |
| Discovery | Configuration | Agent Card |
| Execution | Synchronous | Async (Tasks) |

### When to Use A2A

1. **Multiple specialized agents** - Each agent has deep expertise
2. **Long-running operations** - Tasks can run for minutes
3. **Distributed systems** - Agents run on different hosts
4. **Agent collaboration** - Agents coordinate complex workflows

### When to Use MCP

1. **AI assistant integration** - Cursor, Claude Desktop
2. **Simple tool invocation** - Quick request/response
3. **Local development** - Single process communication
4. **Direct tool access** - No need for agent abstraction

## Example Workflows

### Workflow 1: GitOps Health Check

```
Orchestrator
    │
    ├──► GitOps Agent: "get-gitops-status"
    │    └── Returns: summary of all Flux resources
    │
    └──► Orchestrator synthesizes report
```

### Workflow 2: Incident Investigation

```
Orchestrator
    │
    ├──► GitOps Agent: "list-kustomizations" (status=failed)
    │    └── Returns: 2 failing kustomizations
    │
    ├──► GitOps Agent: "debug-reconciliation" (for each failure)
    │    └── Returns: detailed error analysis
    │
    ├──► Policy Agent: "check-policy-violations"
    │    └── Returns: 0 violations (not a policy issue)
    │
    └──► Orchestrator: "Root cause is missing secret"
```

### Workflow 3: Deployment Validation

```
Orchestrator
    │
    ├──► GitOps Agent: "list-kustomizations"
    │
    ├──► GitOps Agent: "get-helmreleases"
    │
    ├──► Policy Agent: "list-constraints"
    │
    └──► Orchestrator generates deployment report
```

## Multi-Agent Patterns

This directory contains examples of three fundamental multi-agent patterns:

### 1. Orchestrator Pattern

**Location**: [`orchestrator/main.go`](orchestrator/main.go)  
**Pattern**: Central coordinator manages multiple specialized agents

**Quick Start:**
```bash
make run-a2a  # Start agent
go run examples/multi-agent/orchestrator/main.go
```

**Use When:**
- Need to query multiple unrelated domains simultaneously
- Coordinating independent agents for incident investigation
- Gathering information from multiple sources in parallel

### 2. Pipeline Pattern

**Location**: [`pipeline/main.go`](pipeline/main.go)  
**Pattern**: Sequential data transformation through chained agents

**Quick Start:**
```bash
make run-a2a  # Start agent
go run examples/multi-agent/pipeline/main.go
```

**Use When:**
- Data needs to be transformed through multiple stages
- Sequential processing (each stage depends on previous)
- ETL-like workflows (Extract → Transform → Load)

### 3. Consensus Pattern

**Location**: [`consensus/main.go`](consensus/main.go)  
**Pattern**: Multiple expert agents independently evaluate and vote

**Quick Start:**
```bash
# Start multiple agents (simulating different experts)
make run-a2a                    # Port 8080
./bin/dm-nkp-gitops-a2a-server serve --port 8081  # Port 8081
./bin/dm-nkp-gitops-a2a-server serve --port 8082  # Port 8082

go run examples/multi-agent/consensus/main.go
```

**Use When:**
- Need high confidence for critical decisions
- Multiple agents should verify the same thing
- Handling disagreement between agents
- Expert consultation scenarios

## Files in This Demo

```
examples/multi-agent/
├── README.md              # This file
├── orchestrator/
│   └── main.go            # Orchestrator pattern demo
├── pipeline/
│   └── main.go            # Pipeline pattern demo
└── consensus/
    └── main.go            # Consensus pattern demo
```

## Learning Path

1. **Start with Orchestrator** - Easiest to understand, most common pattern
2. **Try Pipeline** - Understand sequential data flow
3. **Explore Consensus** - Learn about confidence and voting
4. **Combine Patterns** - Build complex workflows using multiple patterns

For detailed explanations of each pattern, see:
- [A2A Learning Guide - Multi-Agent Patterns](../../docs/A2A_LEARNING_GUIDE.md#multi-agent-patterns-deep-dive)

## Pattern Comparison

| Pattern | Execution | Use Case | Example |
|---------|-----------|----------|---------|
| **Orchestrator** | Parallel | Multi-domain queries | Incident investigation across GitOps, security, monitoring |
| **Pipeline** | Sequential | Data transformation | Collect → Analyze → Report |
| **Consensus** | Parallel | Critical decisions | Deployment approval with multiple expert validations |

## Next Steps

1. **Run all three patterns** - Understand the differences
2. **Study the code** - See how each pattern is implemented
3. **Modify examples** - Add your own agents or stages
4. **Combine patterns** - Build complex workflows using multiple patterns
5. **Read the guide** - See [A2A Learning Guide](../../docs/A2A_LEARNING_GUIDE.md) for detailed explanations

## Troubleshooting

### Agent not responding

```bash
# Check if agent is running
curl http://localhost:8080/health

# Check agent card
curl http://localhost:8080/.well-known/agent.json | jq
```

### Task stuck

```bash
# List all tasks
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tasks/list"}' | jq

# Cancel a task
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tasks/cancel","params":{"taskId":"TASK_ID"}}' | jq
```

### Debug logging

Start agents with `--log-level=debug` to see detailed logs.
