# Deployment and Usage Guide

This guide clarifies what needs to be deployed and how to actually use the MCP server for troubleshooting.

## Key Concepts - What is What?

### 1. **MCP Server (The Actual Tool)**
The **MCP server** is the deployable component that provides troubleshooting capabilities.

- **What it is**: A Go binary (`dm-nkp-gitops-mcp-server`)
- **What it does**: Provides 15+ tools that AI assistants can call
- **Where it runs**: On your local machine (not in Kubernetes)
- **How it connects**: Uses your kubeconfig to talk to Kubernetes API

### 2. **Runbooks (Documentation/Examples)**
The **runbooks** are documentation that shows how to USE the MCP server.

- **What they are**: Markdown documentation files
- **What they do**: Guide you/AI on what questions to ask
- **Where they are**: In `docs/` directory
- **Do they deploy?**: NO - they're just documentation

### 3. **Troubleshooter Examples (Learning Code)**
The **troubleshooter.go** is example code for learning.

- **What it is**: Example Go code showing workflow patterns
- **What it does**: Demonstrates how runbooks could be automated
- **Where it is**: In `examples/troubleshooting/`
- **Do you deploy it?**: NO - it's just for learning

---

## Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    YOU (The User)                                   │
│  Ask questions in Cursor/Claude Desktop chat                        │
│  Example: "What's the GitOps status?"                              │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                │ Natural Language Question
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│              AI Assistant (Claude in Cursor)                        │
│  - Understands your question                                       │
│  - Decides which tool to use                                       │
│  - Formats request as JSON-RPC                                     │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                │ JSON-RPC 2.0 over stdio
                                │ Method: "tools/call"
                                │ Tool: "get_gitops_status"
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│              MCP Server Process                                     │
│  (dm-nkp-gitops-mcp-server)                                        │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 1. Receives tool call request                                │  │
│  │ 2. Finds handler for "get_gitops_status"                     │  │
│  │ 3. Executes handler (pkg/tools/flux_handlers.go)            │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
│                              │                                       │
│                              │ Calls Kubernetes API                 │
│                              ▼                                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ 4. Uses kubeconfig to authenticate                           │  │
│  │ 5. Queries Flux CRDs (Kustomizations, GitRepositories)      │  │
│  │ 6. Formats response as markdown                              │  │
│  └───────────────────────────┬──────────────────────────────────┘  │
└──────────────────────────────┼───────────────────────────────────────┘
                               │
                               │ JSON-RPC Response
                               │ Result: Markdown formatted status
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│              AI Assistant                                            │
│  - Receives tool result                                             │
│  - Formats response for you                                         │
│  - Answers your question                                            │
└───────────────────────────────┬─────────────────────────────────────┘
                                │
                                │ Natural Language Answer
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    YOU (The User)                                   │
│  See answer: "GitOps status shows 3 ready, 1 failed..."            │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Step-by-Step Deployment

### Step 1: Build the MCP Server

```bash
# Clone the repo (if you haven't)
cd /path/to/workspace

# Build the binary
make build

# This creates: ./bin/dm-nkp-gitops-mcp-server
```

### Step 2: Install/Deploy the Binary

You have two options:

#### Option A: Local Installation (Recommended for Development)

```bash
# Install to your PATH
make install

# Or manually copy
cp ./bin/dm-nkp-gitops-mcp-server ~/bin/
# or
cp ./bin/dm-nkp-gitops-mcp-server /usr/local/bin/
```

#### Option B: Use Full Path

No installation needed - just use the full path to the binary.

### Step 3: Configure Cursor/Claude Desktop

Edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/full/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "/path/to/your/kubeconfig"
      }
    }
  }
}
```

**Important**: Replace:
- `/full/path/to/dm-nkp-gitops-mcp-server` with actual path (e.g., `/Users/deepak/go/src/github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/bin/dm-nkp-gitops-mcp-server`)
- `/path/to/your/kubeconfig` with your actual kubeconfig path

### Step 4: Restart Cursor

Restart Cursor/Claude Desktop for the MCP server to be loaded.

### Step 5: Verify It's Working

In Cursor/Claude Desktop, ask:
```
"What tools are available from the GitOps MCP server?"
```

The AI should list all available tools.

---

## Actual Troubleshooting Usage

### Example 1: Check GitOps Status

**You ask in Cursor:**
```
"What's the GitOps status in flux-system namespace?"
```

**What happens:**
1. AI decides to use `get_gitops_status` tool
2. Calls MCP server with: `{"tool": "get_gitops_status", "arguments": {"namespace": "flux-system"}}`
3. MCP server queries Kubernetes API
4. Returns formatted status
5. AI presents answer to you

**You see:**
```
GitOps Status in flux-system:
- Ready Kustomizations: 5
- Failed Kustomizations: 1
- Ready GitRepositories: 3
```

### Example 2: Debug a Failing Resource

**You ask:**
```
"Why is the infrastructure Kustomization failing in flux-system?"
```

**What happens:**
1. AI uses `debug_reconciliation` tool
2. MCP server queries Flux CRDs and events
3. Returns detailed debug information
4. AI explains the issue

**You see:**
```
The infrastructure Kustomization is failing because:
- Error: Dependency 'base-cluster-resources' is not ready
- Source: GitRepository 'cluster-config' is missing
```

### Example 3: Check Security

**You ask:**
```
"Are there any policy violations in my cluster?"
```

**What happens:**
1. AI uses `check_policy_violations` tool
2. MCP server checks Gatekeeper/Kyverno
3. Returns list of violations
4. AI summarizes security issues

---

## What About the Runbooks?

The **runbooks** (`docs/K8S_TROUBLESHOOTING_RUNBOOK.md`, etc.) are:
- **Documentation** for you and the AI
- **Guide** on what questions to ask
- **Reference** for common troubleshooting procedures
- **NOT deployed** - they're just files in your repo

The AI can read these runbooks to:
- Understand what tools to use for different scenarios
- Follow structured troubleshooting workflows
- Provide better, more structured answers

**How AI uses runbooks:**
1. You ask a question
2. AI may reference runbooks for guidance
3. AI follows the workflow described in runbook
4. AI executes the tool calls
5. AI provides structured answer

---

## What About the Troubleshooter Example Code?

The `examples/troubleshooting/troubleshooter.go` is:
- **Learning code** - shows how workflows work
- **Example implementation** - demonstrates patterns
- **NOT needed for deployment** - it's educational

You can run it to learn:
```bash
make build-troubleshooter
./bin/troubleshooter gitops-failure
```

But it's **NOT** part of the actual MCP server deployment.

---

## Deployment Summary

### ✅ What You Deploy:

1. **MCP Server Binary**
   ```bash
   make build
   # Creates: ./bin/dm-nkp-gitops-mcp-server
   ```

2. **Configuration** (in Cursor config file)
   ```json
   {
     "mcpServers": {
       "dm-nkp-gitops": {
         "command": "/path/to/binary",
         "args": ["serve", "--read-only"],
         "env": {"KUBECONFIG": "/path/to/kubeconfig"}
       }
     }
   }
   ```

### ❌ What You DON'T Deploy:

1. **Runbooks** - They're just documentation files
2. **Troubleshooter example** - It's learning code
3. **Source code** - Only the binary is needed

---

## Quick Start Checklist

- [ ] Build the binary: `make build`
- [ ] Find the binary path: `./bin/dm-nkp-gitops-mcp-server`
- [ ] Edit `~/.cursor/mcp.json` with correct paths
- [ ] Restart Cursor
- [ ] Test: Ask "What tools are available?"
- [ ] Start troubleshooting: Ask about your cluster

---

## Common Questions

### Q: Do I need to deploy the runbooks somewhere?
**A:** No! They're just documentation. The AI can read them from your repo if needed, but they're not required for the MCP server to work.

### Q: Do I need to run the troubleshooter.go code?
**A:** No! It's just example code for learning. The actual troubleshooting happens through the AI assistant using the MCP server tools.

### Q: Where does the MCP server run?
**A:** On your local machine. Cursor spawns it as a child process and communicates via stdin/stdout (stdio).

### Q: Does it need to be in Kubernetes?
**A:** No! The MCP server runs locally and connects TO your Kubernetes cluster using your kubeconfig. It doesn't run IN Kubernetes.

### Q: Can I use it from multiple machines?
**A:** Yes, but you need to:
1. Build/install the binary on each machine
2. Configure Cursor on each machine
3. Have access to kubeconfig on each machine

---

## Troubleshooting the Deployment

### Problem: "MCP server not found"
**Solution:** Check the path in `~/.cursor/mcp.json` is correct and the binary exists.

### Problem: "Cannot connect to Kubernetes"
**Solution:** Verify your kubeconfig path is correct and you can run `kubectl` successfully.

### Problem: "No tools available"
**Solution:** Restart Cursor after configuration changes. Check logs in Cursor's developer console.

### Problem: "Permission denied"
**Solution:** Make sure the binary is executable: `chmod +x /path/to/dm-nkp-gitops-mcp-server`

---

## Next Steps

1. ✅ Deploy the MCP server (build + configure)
2. ✅ Test it works (ask AI a question)
3. 📖 Read the runbooks to understand workflows
4. 🔧 Start troubleshooting your cluster!

---

## Related Documentation

- [MCP Server Architecture](MCP_SERVER_ARCHITECTURE.md) - Technical details
- [Tools Reference](TOOLS_REFERENCE.md) - All available tools
- [K8S Troubleshooting Runbook](K8S_TROUBLESHOOTING_RUNBOOK.md) - How to troubleshoot
- [Quick Start](../README.md#quick-start) - Basic setup
