# Testing and Learning Guide

This guide shows you how to build, test, and learn from the troubleshooting examples.

## Quick Start

### Build the Troubleshooter

```bash
# From project root
make build-troubleshooter
```

Or directly:
```bash
cd examples/troubleshooting
go build -o troubleshooter troubleshooter.go
```

### Run Examples

```bash
# Run a specific workflow
make run-troubleshooter WORKFLOW=gitops-failure

# Or directly
./bin/troubleshooter gitops-failure
./bin/troubleshooter cluster-node
./bin/troubleshooter app-deployment
```

### Test All Workflows

```bash
make test-troubleshooter
```

## Learning Path

### Phase 1: Understand the Structure (15 minutes)

1. **Read the code structure**
   ```bash
   # View the main file
   cat examples/troubleshooting/troubleshooter.go
   ```

2. **See what workflows are available**
   ```bash
   ./bin/troubleshooter  # Shows usage
   ```

3. **Run a workflow to see output**
   ```bash
   ./bin/troubleshooter gitops-failure
   ```

   **What to observe:**
   - How steps are numbered
   - How tool calls are structured
   - How arguments use template variables
   - How expected outputs are described

### Phase 2: Explore Workflow Definitions (30 minutes)

1. **See all workflows as JSON**
   ```bash
   make troubleshoot-workflows-json
   # or
   ./bin/troubleshooter all | jq '.'
   ```

2. **Examine the data structure**
   ```bash
   # Pretty print with jq
   ./bin/troubleshooter all | jq '.gitops-failure.steps[0]'
   ```

3. **Understand decision trees**
   ```bash
   ./bin/troubleshooter all | jq '.gitops-failure.decision_tree'
   ```

### Phase 3: Modify and Experiment (45 minutes)

1. **Create a simple test workflow**

   Edit `troubleshooter.go` and add:
   ```go
   func getTestWorkflow() TroubleshootWorkflow {
       return TroubleshootWorkflow{
           Name:        "Test Workflow",
           Description: "A simple test workflow",
           Steps: []Step{
               {
                   Number:      1,
                   Description: "First step - check context",
                   Tool:        "get_current_context",
                   Arguments:   map[string]interface{}{},
                   Expected:    "Current Kubernetes context",
               },
           },
       }
   }
   ```

   Then add it to `main()`:
   ```go
   case "test":
       workflows = map[string]TroubleshootWorkflow{
           "test": getTestWorkflow(),
       }
   ```

   Test it:
   ```bash
   go build -o troubleshooter troubleshooter.go
   ./troubleshooter test
   ```

2. **Modify an existing workflow**

   Try changing the expected output description in `getGitOpsFailureWorkflow()` and see how it affects the output.

3. **Add a new step**

   Add a step to an existing workflow and see how the flow changes.

### Phase 4: Integrate with Real Tools (1+ hours)

1. **Understand MCP protocol**

   The workflows generate MCP requests. Study the output:
   ```bash
   ./bin/troubleshooter gitops-failure | grep -A 10 "MCP Request"
   ```

2. **Connect to actual MCP server**

   To execute these workflows with a real MCP server:
   
   ```bash
   # Start your MCP server
   make run-readonly
   
   # In another terminal, send a request
   echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_gitops_status","arguments":{"namespace":"flux-system"}}}' | ./bin/dm-nkp-gitops-mcp-server serve --read-only
   ```

3. **Build a workflow executor**

   See Phase 4 in the main README for building a complete executor that:
   - Connects to MCP server
   - Executes workflow steps
   - Parses responses
   - Follows decision trees

## Interactive Testing

### Test Individual Workflows

```bash
# GitOps Failure
./bin/troubleshooter gitops-failure

# Cluster Node Issues
./bin/troubleshooter cluster-node

# App Deployment Issues
./bin/troubleshooter app-deployment
```

### See JSON Structure

```bash
# All workflows as JSON
./bin/troubleshooter all | jq '.'

# Specific workflow
./bin/troubleshooter all | jq '.gitops-failure'

# Just the steps
./bin/troubleshooter all | jq '.gitops-failure.steps'

# Decision tree
./bin/troubleshooter all | jq '.gitops-failure.decision_tree'
```

### Generate MCP Requests

The output includes example MCP requests. Look for the "Example MCP Request" section in the output.

## Understanding the Code

### Key Components

1. **Workflow Structure**
   - `TroubleshootWorkflow` - Contains name, description, steps, decision tree
   - `Step` - Individual troubleshooting step with tool call
   - `DecisionTree` - Branching logic

2. **Template Resolution**
   - Variables like `${namespace}` are replaced at runtime
   - See `resolveArguments()` function

3. **MCP Request Generation**
   - `GenerateMCPRequest()` creates valid JSON-RPC 2.0 requests
   - Ready to send to MCP server

### Code Flow

```
main()
  └─> Parse arguments
  └─> Get workflow(s)
  └─> ExecuteWorkflow()
       └─> For each step:
            └─> resolveArguments() - Replace templates
            └─> Print step details
  └─> GenerateMCPRequest() - Show example request
```

## Practice Exercises

### Exercise 1: Simple Status Check

Create a workflow that:
1. Gets current context
2. Gets GitOps status
3. If failed, lists failing resources

**Solution approach:**
- Start with `getGitOpsFailureWorkflow()` as template
- Simplify to just 3 steps
- Remove decision tree for now

### Exercise 2: Add Error Handling

Modify the code to:
- Handle missing context variables gracefully
- Show what would be sent vs. actual values
- Validate tool names

### Exercise 3: Build a Workflow Executor

Create a new file `executor.go` that:
1. Takes a workflow as input
2. Connects to MCP server (stdio or HTTP)
3. Executes each step sequentially
4. Parses responses
5. Follows decision tree branches

## Troubleshooting the Troubleshooter

### Common Issues

**"Unknown workflow"**
- Check that you're using a valid workflow name
- Use `./bin/troubleshooter` (no args) to see available workflows

**"Build fails"**
- Ensure you're in the project root or examples/troubleshooting directory
- Check Go version: `go version` (needs 1.22+)
- Run `go mod tidy` if dependency issues

**"jq not found"**
- Install jq: `brew install jq` (macOS) or `apt-get install jq` (Linux)
- Or skip jq and view raw JSON output

### Debugging

**Add debug output:**
```go
// In ExecuteWorkflow, add:
fmt.Printf("DEBUG: Resolved args: %+v\n", args)
```

**Check template resolution:**
```go
// In resolveArguments, add:
fmt.Printf("DEBUG: Resolving %s from %v\n", key, v)
```

## Next Steps

1. **Study the runbook documentation**
   - Read `docs/K8S_TROUBLESHOOTING_RUNBOOK.md`
   - See how workflows map to documented procedures

2. **Experiment with modifications**
   - Change workflows
   - Add new steps
   - Modify decision trees

3. **Build real integrations**
   - Connect to MCP servers
   - Execute actual tool calls
   - Parse and act on responses

4. **Create your own workflows**
   - Document a real troubleshooting scenario
   - Convert to workflow structure
   - Test with actual issues

## Related Documentation

- [Troubleshooting README](README.md) - Overview of examples
- [K8S Troubleshooting Runbook](../../docs/K8S_TROUBLESHOOTING_RUNBOOK.md) - Complete runbook
- [Runbook Best Practices](../../docs/RUNBOOK_BEST_PRACTICES.md) - How to create runbooks
- [Tools Reference](../../docs/TOOLS_REFERENCE.md) - Available MCP tools

## Quick Reference

```bash
# Build
make build-troubleshooter

# Run specific workflow
make run-troubleshooter WORKFLOW=gitops-failure
./bin/troubleshooter gitops-failure

# Test all
make test-troubleshooter

# See JSON
make troubleshoot-workflows-json
./bin/troubleshooter all | jq '.'

# Direct Go execution (no build)
go run examples/troubleshooting/troubleshooter.go gitops-failure
```
