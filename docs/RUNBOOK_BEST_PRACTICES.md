# Runbook Best Practices for AI Agents

This document outlines standard practices for creating troubleshooting runbooks that are effective for both humans and AI agents.

## Why This Matters

Well-structured runbooks enable:
- **AI agents** to autonomously troubleshoot issues
- **Humans** to quickly find solutions
- **Teams** to share knowledge consistently
- **Automation** to execute diagnostic workflows

## Core Principles

### 1. Structure for Machine Readability

AI agents need structured, parseable information:

✅ **Good:**
```markdown
## Step 1: Check Status
**Tool:** `get_gitops_status`
**Arguments:**
- namespace: "${namespace}"
**Expected Output:** Count of ready/failed resources
```

❌ **Bad:**
```markdown
First, you should check the overall status of things. 
Use the get gitops status tool with the namespace parameter.
Look at what it returns.
```

### 2. Use Decision Trees

Clear if-then logic helps agents make choices:

```
Start → Check Status
│
├─ Status = Failed
│  └─ Debug Resource → Fix Issue
│
└─ Status = Ready
   └─ No Action Needed
```

### 3. Provide Context

Include:
- **When to use** this runbook
- **Prerequisites** required
- **Expected outcomes**
- **Common variations**

### 4. Make It Actionable

Every step should:
- Specify exact tool/command to use
- Include all required parameters
- Show expected output format
- Indicate next action

### 5. Include Working Examples

Real code examples:
- Demonstrate usage
- Can be copy-pasted
- Show edge cases
- Build confidence

## Standard Runbook Structure

### Template

```markdown
# [Problem Name] Runbook

## Problem Statement
Clear description of the issue this runbook solves.

## Symptoms
How to recognize this problem:
- Symptom 1
- Symptom 2

## Quick Diagnostic
One-line check: "Use tool X with parameter Y"

## Diagnostic Workflow

### Step 1: [Action]
**Agent Query:** "Natural language query"
**Tool:** tool_name
**Parameters:**
- param1: value1
- param2: "${variable}"

**Expected Output:**
Description of what success looks like

**Next Steps:**
- If condition A → Step 2
- If condition B → Step 3

### Step 2: [Action]
...

## Root Cause Analysis

| Symptom | Likely Cause | Solution |
|---------|--------------|----------|
| X | Y | Z |

## Solution Steps

Detailed resolution procedures...

## Complete Example

Full workflow walkthrough with actual outputs...

## Related Runbooks
- Link to other relevant procedures
```

## Design Patterns

### Pattern 1: Hierarchical Investigation

**Start broad, narrow down:**

```
1. Overall Status Check
   ↓
2. List Affected Resources
   ↓
3. Debug Specific Resource
   ↓
4. Deep Dive into Error
```

**Example:**
```
1. get_gitops_status(namespace)
   → Shows 2 failed resources
   
2. list_kustomizations(status_filter: "failed")
   → Lists: infrastructure, base-resources
   
3. debug_reconciliation(infrastructure)
   → Error: "dependency not ready"
   
4. debug_reconciliation(base-resources)
   → Root cause found
```

### Pattern 2: Dependency Chain Resolution

When issues cascade:

```
Workflow:
1. Start with failing resource
2. For each dependency:
   a. Check if dependency is ready
   b. If not, recursively debug
   c. Fix dependencies leaf-to-root
3. Work back up the chain
```

### Pattern 3: Event Correlation

Match events with resource state:

```
1. Get resource status (current state)
2. Get recent events (what happened)
3. Match timestamps (when it changed)
4. Find first error event (root cause)
5. Investigate that event's context
```

## Formatting Guidelines

### Tool Calls

Always include:
- Exact tool name
- All parameters with types
- Example values

**Format:**
```json
{
  "tool": "tool_name",
  "arguments": {
    "param1": "value1",
    "param2": "${variable}"
  }
}
```

### Tables

Use tables for:
- Root cause analysis
- Comparison matrices
- Parameter reference

**Example:**
| Error | Cause | Fix |
|-------|-------|-----|
| X | Y | Z |

### Code Blocks

Use code blocks for:
- Actual commands
- JSON/XML structures
- Example outputs
- Working scripts

### Variables

Use template syntax: `${variable}`

Replace at runtime with:
- Namespace names
- Resource names
- Cluster names
- Context-specific values

## Testing Your Runbook

### Checklist

- [ ] Can an AI agent follow it without clarification?
- [ ] Are all tool names and parameters exact?
- [ ] Do examples actually work?
- [ ] Is the decision tree clear?
- [ ] Can a human understand it quickly?
- [ ] Are edge cases covered?

### Validation

Test with:
1. **AI Agent** - Ask it to follow the runbook
2. **New Team Member** - Can they understand it?
3. **Automation** - Can you encode it programmatically?

## Common Pitfalls

### ❌ Too Vague

**Bad:**
"Check the status and see if anything is wrong"

**Good:**
"Call `get_gitops_status(namespace='flux-system')` and verify failed count is 0"

### ❌ Missing Context

**Bad:**
"Use tool X"

**Good:**
"Use tool X when resource Y shows status Z"

### ❌ No Examples

**Bad:**
"Debug the resource"

**Good:**
```
"Debug the infrastructure Kustomization in flux-system"
→ Tool: debug_reconciliation(resource_type: "kustomization", 
                              name: "infrastructure", 
                              namespace: "flux-system")
```

### ❌ Unclear Decision Points

**Bad:**
"If it fails, try something else"

**Good:**
```
If error contains "Source":
  → Check GitRepository status (Step 4)
Else if error contains "Dependency":
  → Debug dependency recursively (Step 5)
```

## Learning Approach

### Phase 1: Understand the Problem Domain (1-2 hours)

1. **Read existing runbooks** - See what works
2. **Study your tools** - Understand capabilities
3. **Practice manually** - Troubleshoot real issues
4. **Take notes** - Document your process

### Phase 2: Structure Your Knowledge (2-3 hours)

1. **Identify common problems** - What issues do you see repeatedly?
2. **Map diagnostic steps** - What do you check first, second, third?
3. **Document decisions** - When do you take different paths?
4. **Collect examples** - Save real error messages and solutions

### Phase 3: Create Runbook (3-4 hours)

1. **Use the template** - Follow standard structure
2. **Write step-by-step** - Be specific and clear
3. **Add examples** - Include real tool calls
4. **Create decision trees** - Encode your logic

### Phase 4: Test and Iterate (ongoing)

1. **Use with real issues** - Does it help?
2. **Get feedback** - From team and AI agents
3. **Refine** - Improve clarity and accuracy
4. **Share** - Contribute to team knowledge

## Example: Building a Runbook from Scratch

### Step 1: Document Your Process

When you encounter a problem, write down:
- What you checked first
- What you found
- What you did next
- How you fixed it

### Step 2: Generalize

Convert your specific steps to a general pattern:
- "Checked pod logs" → "Get logs from failing pod"
- "Saw image pull error" → "Check for ImagePullBackOff events"

### Step 3: Structure

Organize into:
- Quick diagnostic
- Step-by-step workflow
- Root cause analysis
- Solutions

### Step 4: Add Examples

Include:
- Real tool calls you used
- Actual error messages
- Working solutions

### Step 5: Test

Try following it:
- With a similar problem
- With an AI agent
- As a new team member

## Advanced Techniques

### 1. Automated Workflow Execution

Encode runbooks as executable workflows:

```go
type Workflow struct {
    Steps []Step
    DecisionTree DecisionTree
}
```

See: [`examples/troubleshooting/troubleshooter.go`](../../examples/troubleshooting/troubleshooter.go)

### 2. Dynamic Runbook Generation

Generate runbooks from:
- Tool capabilities
- Common error patterns
- Historical troubleshooting data

### 3. Integration with MCP

Runbooks can directly call MCP tools:
- Structured tool calls
- Parameter validation
- Response parsing

## Tools and Resources

### Documentation Tools

- **Markdown** - Easy to write and parse
- **Mermaid** - For flowcharts/diagrams
- **JSON** - For structured data
- **YAML** - For configuration

### AI Agent Considerations

- **Explicit over implicit** - Be very clear
- **Structured data** - Tables, lists, code blocks
- **Examples** - Show, don't just tell
- **Context** - Include when/why/where

## Success Metrics

A good runbook enables:
- ✅ Autonomous AI agent troubleshooting
- ✅ Faster human problem resolution
- ✅ Consistent team approach
- ✅ Knowledge preservation
- ✅ Automation opportunities

## Related Documentation

- [K8S Troubleshooting Runbook](K8S_TROUBLESHOOTING_RUNBOOK.md) - Comprehensive example
- [Tools Reference](TOOLS_REFERENCE.md) - Available tools
- [Troubleshooting Examples](../../examples/troubleshooting/) - Working code
- [AGENTS.md](../AGENTS.md) - AI agent integration

## Contributing

When adding runbooks:
1. Follow the standard structure
2. Include working examples
3. Test with AI agents
4. Get team review
5. Iterate based on feedback

---

**Remember:** The best runbook is one that someone (human or AI) can follow successfully without asking for clarification.
