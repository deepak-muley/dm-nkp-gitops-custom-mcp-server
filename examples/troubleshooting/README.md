# Troubleshooting Examples

This directory contains working code examples that demonstrate how to create and use troubleshooting runbooks programmatically.

## Overview

These examples show:
1. **Structured workflow definitions** - How to represent troubleshooting procedures as data structures
2. **Decision tree patterns** - How to encode decision logic
3. **Tool call generation** - How to generate MCP protocol requests
4. **Template resolution** - How to fill in context-specific values

## Quick Start

### Using Makefile (Recommended)

```bash
# From project root
make build-troubleshooter    # Build the binary
make run-troubleshooter WORKFLOW=gitops-failure  # Run a workflow
make test-troubleshooter     # Test all workflows
make troubleshoot-workflows-json  # See all as JSON
```

### Direct Go Execution

```bash
# From project root or examples/troubleshooting
go run examples/troubleshooting/troubleshooter.go gitops-failure
go run examples/troubleshooting/troubleshooter.go cluster-node
go run examples/troubleshooting/troubleshooter.go app-deployment

# See all workflows as JSON
go run examples/troubleshooting/troubleshooter.go all
```

### After Building

```bash
# Build first
make build-troubleshooter

# Then run
./bin/troubleshooter gitops-failure
./bin/troubleshooter cluster-node
./bin/troubleshooter app-deployment
```

📖 **[Complete Testing Guide →](TESTING_GUIDE.md)** - Detailed instructions for learning and testing

### Example Output

```
=== Starting Workflow: GitOps Reconciliation Failure ===

Description: Systematic approach to diagnosing and fixing GitOps reconciliation failures

Step 1: Get overall GitOps status to identify failing resources
  Tool: get_gitops_status
  Arguments:
  {
    "namespace": "flux-system"
  }
  Expected: Shows count of ready/failed/suspended resources
```

## Understanding the Code

### Workflow Structure

A `TroubleshootWorkflow` contains:
- **Name** - Human-readable workflow name
- **Description** - What the workflow does
- **Steps** - Sequential actions with tool calls
- **DecisionTree** - Optional branching logic

### Step Definition

Each step includes:
- **Number** - Execution order
- **Description** - What this step does
- **Tool** - MCP tool to call
- **Arguments** - Tool parameters (can use `${variable}` templates)
- **Expected** - What output to expect
- **NextStep** - Conditional branching

### Decision Trees

Decision trees encode branching logic:
- Start with root condition
- Branch based on true/false
- Leaf nodes specify actions

## Learning Path

### Phase 1: Understand the Structure (15 min)

1. Read `troubleshooter.go` - Understand data structures
2. Run each workflow - See how they execute
3. Examine JSON output - See the complete structure

### Phase 2: Modify Workflows (30 min)

1. Create a new workflow function (e.g., `getPolicyViolationWorkflow`)
2. Add steps for your specific scenario
3. Test with: `go run troubleshooter.go <your-workflow>`

### Phase 3: Integrate with MCP (1 hour)

1. Connect to your MCP server
2. Execute actual tool calls
3. Parse responses and follow decision tree
4. Handle errors and edge cases

### Phase 4: Build Your Own Runbook (ongoing)

1. Document a real troubleshooting scenario
2. Convert it to workflow structure
3. Test with actual cluster issues
4. Iterate and improve

## Example: Creating a Custom Workflow

```go
func getCustomWorkflow() TroubleshootWorkflow {
    return TroubleshootWorkflow{
        Name:        "Your Custom Issue",
        Description: "How to fix your specific problem",
        Steps: []Step{
            {
                Number:      1,
                Description: "First diagnostic step",
                Tool:        "get_gitops_status",
                Arguments: map[string]interface{}{
                    "namespace": "${namespace}",
                },
                Expected: "Expected output",
                NextStep: map[string]int{
                    "found_issue": 2,
                    "no_issue":    0, // End
                },
            },
            // Add more steps...
        },
    }
}
```

## Integration with Runbook Documentation

These workflows correspond to scenarios in:
- [`docs/K8S_TROUBLESHOOTING_RUNBOOK.md`](../../docs/K8S_TROUBLESHOOTING_RUNBOOK.md)

The code provides executable form of the documented procedures.

## Best Practices

### 1. Start Simple
Begin with linear workflows (no branching). Add decision logic later.

### 2. Use Templates
Use `${variable}` syntax for context-specific values (namespace, resource names, etc.)

### 3. Document Expected Output
Always specify what each step should produce. This helps with validation.

### 4. Handle Edge Cases
Use `NextStep` to branch based on different outcomes.

### 5. Make It Reusable
Structure workflows so they can be applied to different namespaces/resources.

## Next Steps

1. **Study the code** - Understand how workflows are structured
2. **Run examples** - See them in action
3. **Modify workflows** - Adapt to your needs
4. **Create new workflows** - Document your own troubleshooting procedures
5. **Integrate with tools** - Connect to real MCP servers

## Related Documentation

- **[Testing Guide](TESTING_GUIDE.md)** - 📚 How to build, test, and learn with these examples
- [K8S Troubleshooting Runbook](../../docs/K8S_TROUBLESHOOTING_RUNBOOK.md) - Human-readable runbook
- [Runbook Best Practices](../../docs/RUNBOOK_BEST_PRACTICES.md) - Standards for creating runbooks
- [Tools Reference](../../docs/TOOLS_REFERENCE.md) - Complete tool documentation
- [AGENTS.md](../../AGENTS.md) - AI agent integration guide
