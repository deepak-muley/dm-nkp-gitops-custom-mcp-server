# MCP (Model Context Protocol) Primer

This document provides a comprehensive guide to understanding the Model Context Protocol (MCP), its architecture, security considerations, and how to build MCP servers.

---

## Table of Contents

1. [What is MCP?](#what-is-mcp)
2. [Architecture Overview](#architecture-overview)
3. [Core Concepts](#core-concepts)
4. [Protocol Specification](#protocol-specification)
5. [Transport Mechanisms](#transport-mechanisms)
6. [Security Best Practices](#security-best-practices)
7. [Building an MCP Server](#building-an-mcp-server)
8. [Client Integration](#client-integration)
9. [References](#references)

---

## What is MCP?

**Model Context Protocol (MCP)** is an open standard introduced by Anthropic in November 2024. It provides a universal way to connect AI models (like Claude, GPT, etc.) to external tools, data sources, and services.

### The Problem MCP Solves

Before MCP, integrating AI assistants with external tools required custom implementations for each combination:

```
N Models × M Tools = N×M Custom Integrations
```

MCP standardizes this:

```
N Models → MCP Protocol → M Tools
         (Single Standard)
```

### Key Benefits

| Benefit | Description |
|---------|-------------|
| **Standardization** | One protocol for all AI-tool integrations |
| **Reusability** | Build once, use with any MCP-compatible AI assistant |
| **Security** | Built-in patterns for authentication, authorization, and auditing |
| **Discoverability** | AI assistants can discover available tools dynamically |
| **Composability** | Multiple MCP servers can be combined |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Architecture                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────────┐ │
│  │ AI Assistant │────▶│  MCP Client  │────▶│   MCP Server     │ │
│  │   (Claude,   │     │  (Built into │     │  (Your Custom    │ │
│  │   Cursor,    │     │   the AI     │     │   Server or      │ │
│  │   VS Code)   │     │   tool)      │     │   Flux MCP)      │ │
│  └──────────────┘     └──────────────┘     └────────┬─────────┘ │
│                                                      │           │
│                              ┌───────────────────────┼───────┐   │
│                              │                       ▼       │   │
│                              │  ┌─────────────────────────┐  │   │
│                              │  │      Tool Handlers      │  │   │
│                              │  ├─────────────────────────┤  │   │
│                              │  │ • get_cluster_status    │  │   │
│                              │  │ • get_gitops_status     │  │   │
│                              │  │ • debug_reconciliation  │  │   │
│                              │  │ • get_app_deployments   │  │   │
│                              │  └───────────┬─────────────┘  │   │
│                              │              │                │   │
│                              │              ▼                │   │
│                              │  ┌─────────────────────────┐  │   │
│                              │  │   External Systems      │  │   │
│                              │  ├─────────────────────────┤  │   │
│                              │  │ • Kubernetes API        │  │   │
│                              │  │ • Git Repositories      │  │   │
│                              │  │ • Monitoring Systems    │  │   │
│                              │  └─────────────────────────┘  │   │
│                              │         MCP Server           │   │
│                              └───────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Role | Examples |
|-----------|------|----------|
| **MCP Host** | Application that embeds an AI model | Claude Desktop, Cursor, VS Code |
| **MCP Client** | Protocol handler within the host | Built into AI tools |
| **MCP Server** | Exposes tools and resources | Flux MCP, custom servers |
| **Tools** | Functions the AI can invoke | `get_cluster_status`, `debug_pod` |
| **Resources** | Data the AI can read | Configuration files, logs |
| **Prompts** | Pre-defined prompt templates | Debugging workflows |

---

## Core Concepts

### 1. Tools

Tools are functions that the AI can invoke. Each tool has:

```json
{
  "name": "get_cluster_status",
  "description": "Get the status of a Kubernetes cluster",
  "inputSchema": {
    "type": "object",
    "properties": {
      "cluster_name": {
        "type": "string",
        "description": "Name of the cluster to check"
      },
      "namespace": {
        "type": "string",
        "description": "Optional namespace filter"
      }
    },
    "required": ["cluster_name"]
  }
}
```

**Tool Invocation Flow:**

```
1. AI decides to use a tool based on user query
2. AI sends tool call request with arguments
3. MCP Server validates input against schema
4. MCP Server executes the tool handler
5. MCP Server returns result to AI
6. AI incorporates result into response
```

### 2. Resources

Resources are data sources the AI can read:

```json
{
  "uri": "file:///path/to/config.yaml",
  "name": "Cluster Configuration",
  "description": "Current cluster configuration file",
  "mimeType": "application/yaml"
}
```

**Resource Types:**
- **Static**: Files, configurations (read once)
- **Dynamic**: Live data, logs (can change)
- **Templated**: URIs with parameters

### 3. Prompts

Pre-defined prompt templates for common workflows:

```json
{
  "name": "debug_gitops_failure",
  "description": "Debug a failing GitOps reconciliation",
  "arguments": [
    {
      "name": "kustomization_name",
      "description": "Name of the failing Kustomization",
      "required": true
    }
  ]
}
```

---

## Protocol Specification

### Message Format

MCP uses JSON-RPC 2.0 for communication:

```json
// Request
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_cluster_status",
    "arguments": {
      "cluster_name": "dm-nkp-workload-1"
    }
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Cluster dm-nkp-workload-1 is healthy..."
      }
    ]
  }
}
```

### Core Methods

| Method | Direction | Description |
|--------|-----------|-------------|
| `initialize` | Client → Server | Initialize connection, exchange capabilities |
| `initialized` | Client → Server | Confirm initialization complete |
| `tools/list` | Client → Server | List available tools |
| `tools/call` | Client → Server | Invoke a tool |
| `resources/list` | Client → Server | List available resources |
| `resources/read` | Client → Server | Read a resource |
| `prompts/list` | Client → Server | List available prompts |
| `prompts/get` | Client → Server | Get a prompt template |
| `notifications/*` | Bidirectional | Various notifications |

### Initialization Sequence

```
Client                                    Server
   │                                         │
   │──── initialize ─────────────────────────▶│
   │     {protocolVersion, capabilities}      │
   │                                         │
   │◀─── initialize response ────────────────│
   │     {protocolVersion, capabilities,      │
   │      serverInfo}                         │
   │                                         │
   │──── initialized notification ───────────▶│
   │                                         │
   │──── tools/list ─────────────────────────▶│
   │                                         │
   │◀─── tools list ─────────────────────────│
   │                                         │
```

### Capabilities

Servers and clients declare their capabilities during initialization:

**Server Capabilities:**
```json
{
  "capabilities": {
    "tools": {},
    "resources": {
      "subscribe": true,
      "listChanged": true
    },
    "prompts": {
      "listChanged": true
    },
    "logging": {}
  }
}
```

**Client Capabilities:**
```json
{
  "capabilities": {
    "roots": {
      "listChanged": true
    },
    "sampling": {}
  }
}
```

---

## Transport Mechanisms

MCP supports multiple transport mechanisms:

### 1. STDIO (Standard Input/Output)

**Best for:** Local integrations, CLI tools, desktop applications

```
┌─────────────┐    stdin     ┌─────────────┐
│  MCP Client │─────────────▶│  MCP Server │
│             │◀─────────────│             │
└─────────────┘    stdout    └─────────────┘
```

**Configuration (Cursor/Claude Desktop):**
```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/my-mcp-server",
      "args": ["--config", "/path/to/config.yaml"],
      "env": {
        "KUBECONFIG": "/path/to/kubeconfig"
      }
    }
  }
}
```

**Pros:**
- Simple to implement
- No network configuration
- Secure (no network exposure)

**Cons:**
- Local only
- One client per server instance

### 2. HTTP with Server-Sent Events (SSE)

**Best for:** Remote servers, web-based integrations

```
┌─────────────┐    HTTP POST    ┌─────────────┐
│  MCP Client │────────────────▶│  MCP Server │
│             │◀────────────────│             │
└─────────────┘    SSE Stream   └─────────────┘
```

**Endpoints:**
- `POST /message` - Send messages to server
- `GET /sse` - Receive server events

**Pros:**
- Remote access
- Multiple clients
- Works through firewalls

**Cons:**
- Requires HTTPS in production
- More complex setup

### 3. Streamable HTTP

**Best for:** High-performance, bidirectional communication

Similar to SSE but supports true bidirectional streaming.

---

## Security Best Practices

### Mandatory Requirements (MUST)

| Requirement | Description |
|-------------|-------------|
| **HTTPS Everywhere** | All HTTP endpoints must use TLS |
| **PKCE Required** | OAuth clients must use Proof Key for Code Exchange |
| **Redirect URI Validation** | Strict validation, no wildcards |
| **Audience Binding** | Tokens must be issued for the specific server |
| **Token Storage** | Secure storage, encryption at rest |

### Security Checklist

```
□ Authentication & Authorization
  □ Implement proper OAuth 2.0 flow with PKCE
  □ Validate all tokens (signature, expiry, audience)
  □ Use short-lived access tokens
  □ Rotate refresh tokens

□ Input Validation
  □ Validate all tool inputs against JSON schema
  □ Sanitize user-provided data
  □ Reject malformed requests

□ Least Privilege
  □ Request minimal Kubernetes RBAC permissions
  □ Use service account impersonation when possible
  □ Implement read-only mode by default

□ Secrets Management
  □ Never log secrets
  □ Mask sensitive data in responses
  □ Use secret managers (Vault, K8s Secrets)
  □ Rotate credentials regularly

□ Audit & Logging
  □ Log all tool invocations
  □ Include timestamp, user, tool, arguments
  □ Don't log sensitive parameters
  □ Monitor for anomalous patterns

□ Network Security
  □ Use TLS for all connections
  □ Prefer STDIO for local integrations
  □ Implement rate limiting
  □ Use network policies in Kubernetes
```

### Common Attack Vectors & Mitigations

| Attack | Description | Mitigation |
|--------|-------------|------------|
| **Prompt Injection** | Malicious input tries to manipulate AI behavior | Input validation, output sanitization |
| **Token Passthrough** | Using tokens intended for other services | Validate audience claim |
| **Tool Poisoning** | Malicious tool descriptions | Sign and version tool definitions |
| **Privilege Escalation** | Gaining unauthorized access | Least privilege, RBAC |
| **Data Exfiltration** | Leaking sensitive data | Output filtering, secret masking |

### Kubernetes-Specific Security

```yaml
# Example: Minimal RBAC for read-only GitOps monitoring
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: mcp-server-readonly
rules:
  # Flux resources
  - apiGroups: ["kustomize.toolkit.fluxcd.io"]
    resources: ["kustomizations"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["source.toolkit.fluxcd.io"]
    resources: ["gitrepositories"]
    verbs: ["get", "list", "watch"]
  # CAPI resources
  - apiGroups: ["cluster.x-k8s.io"]
    resources: ["clusters", "machines"]
    verbs: ["get", "list", "watch"]
  # Core resources
  - apiGroups: [""]
    resources: ["pods", "events", "namespaces"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
```

---

## Building an MCP Server

### Server Structure

```
my-mcp-server/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── pkg/
│   ├── mcp/
│   │   ├── server.go         # MCP protocol handling
│   │   ├── types.go          # MCP types (Tool, Resource, etc.)
│   │   └── transport.go      # STDIO/HTTP transport
│   ├── tools/
│   │   ├── registry.go       # Tool registration
│   │   ├── cluster.go        # Cluster status tools
│   │   ├── gitops.go         # GitOps status tools
│   │   └── debug.go          # Debugging tools
│   └── config/
│       └── config.go         # Configuration management
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Implementation Steps

#### 1. Define Tool Schema

```go
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
    Type       string              `json:"type"`
    Properties map[string]Property `json:"properties"`
    Required   []string            `json:"required,omitempty"`
}
```

#### 2. Implement Tool Handler

```go
func (h *ToolHandler) GetClusterStatus(args map[string]interface{}) (*ToolResult, error) {
    clusterName, ok := args["cluster_name"].(string)
    if !ok {
        return nil, fmt.Errorf("cluster_name is required")
    }
    
    // Query Kubernetes API
    cluster, err := h.client.GetCluster(clusterName)
    if err != nil {
        return nil, err
    }
    
    return &ToolResult{
        Content: []Content{
            {Type: "text", Text: formatClusterStatus(cluster)},
        },
    }, nil
}
```

#### 3. Handle JSON-RPC Messages

```go
func (s *Server) handleMessage(msg []byte) ([]byte, error) {
    var request JSONRPCRequest
    if err := json.Unmarshal(msg, &request); err != nil {
        return s.errorResponse(nil, -32700, "Parse error")
    }
    
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
```

#### 4. Implement Transport

```go
// STDIO Transport
func (s *Server) RunSTDIO() error {
    reader := bufio.NewReader(os.Stdin)
    writer := bufio.NewWriter(os.Stdout)
    
    for {
        line, err := reader.ReadBytes('\n')
        if err == io.EOF {
            return nil
        }
        
        response, err := s.handleMessage(line)
        if err != nil {
            // Handle error
        }
        
        writer.Write(response)
        writer.Write([]byte("\n"))
        writer.Flush()
    }
}
```

---

## Client Integration

### Cursor Configuration

Location: `~/.cursor/mcp.json` (or via Settings UI)

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf",
        "MCP_READ_ONLY": "true",
        "MCP_LOG_LEVEL": "info"
      }
    }
  }
}
```

### Claude Desktop Configuration

Location: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
      }
    }
  }
}
```

### VS Code with Copilot

Location: `.vscode/settings.json`

```json
{
  "mcp": {
    "servers": {
      "dm-nkp-gitops": {
        "command": "/path/to/dm-nkp-gitops-mcp-server",
        "args": ["serve"],
        "env": {
          "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
        }
      }
    }
  },
  "chat.mcp.enabled": true
}
```

### Testing Your Server

```bash
# Test with simple echo
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./dm-nkp-gitops-mcp-server serve

# Use MCP Inspector (if available)
npx @anthropic/mcp-inspector ./dm-nkp-gitops-mcp-server serve
```

---

## References

### Official Documentation

- [MCP Specification](https://modelcontextprotocol.io/specification)
- [MCP Security Best Practices](https://modelcontextprotocol.io/specification/2025-11-25/basic/security_best_practices)
- [Flux MCP Server](https://fluxoperator.dev/docs/mcp/)

### Related Projects

- [Kagent](https://kagent.dev) - CNCF Sandbox project for AI agents in Kubernetes
- [kmcp](https://kagent.dev/docs/kmcp) - Kubernetes MCP toolkit
- [Flux Operator](https://fluxoperator.dev) - Flux CD with MCP support

### SDKs and Libraries

- [MCP TypeScript SDK](https://github.com/anthropics/mcp-typescript-sdk)
- [MCP Python SDK](https://github.com/anthropics/mcp-python-sdk)
- [MCP Go SDK](https://github.com/mark3labs/mcp-go) (Community)

---

## Summary

MCP provides a standardized way to connect AI assistants to external tools and data sources. Key takeaways:

1. **Tools** are functions the AI can invoke
2. **Resources** are data sources the AI can read
3. **JSON-RPC 2.0** is the message format
4. **STDIO** is simplest for local integrations
5. **Security** requires careful attention to authentication, authorization, and input validation
6. **Kubernetes integration** requires proper RBAC and secret handling

For your GitOps use case, an MCP server provides a powerful way to give AI assistants visibility into cluster state, debugging capabilities, and operational awareness.
