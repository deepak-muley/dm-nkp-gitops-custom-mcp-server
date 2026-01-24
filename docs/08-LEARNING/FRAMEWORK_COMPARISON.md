# MCP/A2A Framework Comparison

This document compares different approaches for building MCP (Model Context Protocol) and A2A (Agent-to-Agent) servers.

## Overview

| Framework | Language | MCP Support | A2A Support | K8s Native | Maturity |
|-----------|----------|-------------|-------------|------------|----------|
| **Custom (this project)** | Go | ✅ | ✅ | ❌ (manual) | Internal |
| **[mcp-go](https://github.com/mark3labs/mcp-go)** | Go | ✅ | ❌ | ❌ | Active |
| **[go-mcp](https://github.com/metoro-io/mcp-golang)** | Go | ✅ | ❌ | ❌ | Active |
| **[kmcp](https://kagent.dev/docs/kmcp)** | Python | ✅ | ✅ | ✅ | CNCF Sandbox |

---

## 1. Custom Implementation (This Project)

### What We Built
- `pkg/mcp/` - JSON-RPC 2.0 over stdio
- `pkg/a2a/` - HTTP server with A2A protocol
- `pkg/tools/` - Shared tool handlers

### Pros
| Advantage | Description |
|-----------|-------------|
| **Full Control** | Complete control over protocol implementation, no external dependencies |
| **Minimal Dependencies** | Only k8s client-go, no framework overhead |
| **Learning Value** | Deep understanding of MCP/A2A protocols |
| **Customization** | Easy to add custom features, modify behavior |
| **Small Binary** | ~15MB static binary with distroless |
| **No Framework Lock-in** | Can migrate to any framework later |

### Cons
| Disadvantage | Description |
|--------------|-------------|
| **Maintenance Burden** | Must maintain protocol implementation ourselves |
| **Missing Features** | No streaming, SSE, or advanced MCP features |
| **No Community Support** | Bugs/issues handled internally |
| **Protocol Updates** | Must manually update when MCP spec changes |
| **Testing** | Must write all protocol compliance tests |

### When to Use
- Learning MCP/A2A protocols
- Simple use cases with few tools
- Maximum control needed
- Minimal dependencies required

---

## 2. mcp-go (mark3labs)

**GitHub**: https://github.com/mark3labs/mcp-go

### What It Provides
```go
import "github.com/mark3labs/mcp-go/mcp"
import "github.com/mark3labs/mcp-go/server"

s := server.NewMCPServer("my-server", "1.0.0")
s.AddTool(mcp.NewTool("get_gitops_status", 
    mcp.WithDescription("Get GitOps status"),
    mcp.WithString("namespace", mcp.Description("Namespace")),
), handleGetGitOpsStatus)
s.ServeStdio()
```

### Pros
| Advantage | Description |
|-----------|-------------|
| **Protocol Compliant** | Full MCP 2024-11-05 spec implementation |
| **Streaming Support** | SSE and streaming responses |
| **Active Development** | Regular updates, bug fixes |
| **Type Safety** | Strong Go typing for MCP types |
| **Resources & Prompts** | Full MCP features (tools, resources, prompts) |
| **Multiple Transports** | stdio, SSE, HTTP |

### Cons
| Disadvantage | Description |
|--------------|-------------|
| **No A2A Support** | MCP only, would need separate A2A implementation |
| **External Dependency** | Adds dependency to your project |
| **Less Control** | Framework decisions may not match your needs |
| **Learning Curve** | Must learn framework API |

### Migration Effort from Custom
- **Medium** - Rewrite tool handlers to match mcp-go interface
- Keep `pkg/tools/` business logic, wrap with mcp-go

---

## 3. go-mcp (metoro-io)

**GitHub**: https://github.com/metoro-io/mcp-golang

### What It Provides
```go
import "github.com/metoro-io/mcp-golang"

server := mcp.NewServer()
server.RegisterTool("get_gitops_status", handler, schema)
server.Start()
```

### Pros
| Advantage | Description |
|-----------|-------------|
| **Simple API** | Minimalist, easy to understand |
| **Lightweight** | Small footprint |
| **Go Idioms** | Follows Go conventions |

### Cons
| Disadvantage | Description |
|--------------|-------------|
| **Less Mature** | Fewer features than mcp-go |
| **No A2A** | MCP only |
| **Limited Docs** | Less documentation |
| **Smaller Community** | Fewer contributors |

### Migration Effort from Custom
- **Low-Medium** - Similar architecture to our custom implementation

---

## 4. kmcp (Kagent)

**Website**: https://kagent.dev/docs/kmcp
**Part of**: CNCF Sandbox project

### What It Provides
- CLI tool for MCP server development
- Kubernetes-native deployment (CRDs)
- Integration with Kagent AI agent framework
- Python/FastMCP based

```bash
# Create MCP server project
kmcp init my-mcp-server

# Deploy to Kubernetes
kmcp deploy --namespace my-ns

# Creates MCPServer CRD
kubectl get mcpservers
```

### Kubernetes CRD Example
```yaml
apiVersion: kagent.dev/v1alpha1
kind: MCPServer
metadata:
  name: gitops-mcp-server
spec:
  image: ghcr.io/my-org/gitops-mcp:v1.0.0
  replicas: 2
  tools:
    - name: get_gitops_status
      description: Get GitOps status
  secrets:
    - name: kubeconfig
      secretRef:
        name: cluster-kubeconfig
```

### Pros
| Advantage | Description |
|-----------|-------------|
| **Kubernetes Native** | CRDs, operators, native k8s integration |
| **A2A Support** | Full A2A protocol support via Kagent |
| **CNCF Project** | Community backing, long-term support |
| **Secret Management** | Built-in secret handling for MCP servers |
| **Multi-Agent** | Orchestrate multiple agents/servers |
| **Production Ready** | Designed for enterprise deployment |
| **Observability** | Built-in metrics, tracing |

### Cons
| Disadvantage | Description |
|--------------|-------------|
| **Python Only** | MCP servers must be Python (FastMCP) |
| **Heavy Infrastructure** | Requires Kagent operator, CRDs |
| **Overkill for Simple Cases** | Complex for single MCP server |
| **Learning Curve** | Must learn Kagent ecosystem |
| **Go Not Supported** | Can't use existing Go tools directly |

### Migration Path
- **High Effort** - Would need to:
  1. Rewrite tools in Python
  2. Deploy Kagent operator
  3. Create MCPServer CRDs
  
- **Alternative**: Keep Go server, use kmcp just for deployment orchestration

---

## Comparison Matrix

| Feature | Custom | mcp-go | go-mcp | kmcp |
|---------|--------|--------|--------|------|
| **Language** | Go | Go | Go | Python |
| **MCP Protocol** | Partial | Full | Partial | Full |
| **A2A Protocol** | ✅ Custom | ❌ | ❌ | ✅ |
| **Streaming/SSE** | ❌ | ✅ | ❌ | ✅ |
| **Resources** | ❌ | ✅ | ✅ | ✅ |
| **Prompts** | ❌ | ✅ | ❌ | ✅ |
| **K8s CRDs** | ❌ | ❌ | ❌ | ✅ |
| **Secret Mgmt** | Manual | Manual | Manual | ✅ |
| **Multi-Agent** | Manual | Manual | Manual | ✅ |
| **Binary Size** | ~15MB | ~20MB | ~18MB | N/A (Python) |
| **Dependencies** | Minimal | Medium | Low | Heavy |
| **Maintenance** | Self | Community | Community | CNCF |

---

## Recommendation

### Keep Custom Implementation If:
- ✅ You want minimal dependencies
- ✅ Learning MCP/A2A protocols is valuable
- ✅ Simple use case (read-only GitOps monitoring)
- ✅ Go is required (team expertise, existing code)
- ✅ You need both MCP + A2A in single codebase

### Migrate to mcp-go If:
- ✅ You need full MCP spec compliance
- ✅ Streaming/SSE is required
- ✅ You want community support
- ✅ A2A is not needed (or handled separately)

### Migrate to kmcp/Kagent If:
- ✅ Deploying to Kubernetes at scale
- ✅ Need multi-agent orchestration
- ✅ Python is acceptable
- ✅ Want CNCF-backed solution
- ✅ Enterprise features needed (secrets, RBAC, etc.)

---

## Hybrid Approach (Recommended for NKP)

For production NKP deployment, consider:

```
┌─────────────────────────────────────────────────────────────┐
│                    Kagent Operator (kmcp)                   │
│                    (Orchestration Layer)                    │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              dm-nkp-gitops-a2a-server (Go)                  │
│              (Your custom implementation)                    │
│                                                             │
│  - Keep existing Go tools                                   │
│  - Deploy as container                                      │
│  - Expose via A2A HTTP endpoint                             │
└─────────────────────────────────────────────────────────────┘
```

This gives you:
1. **Keep Go implementation** - No rewrite needed
2. **Kagent orchestration** - Multi-agent, secrets, scaling
3. **A2A interoperability** - Works with other Kagent agents

---

---

## Why Custom Was Written (Historical Context)

When this project started:

1. **MCP was brand new** - No mature Go frameworks existed
2. **A2A was emerging** - Custom implementation needed to support both protocols
3. **Team expertise** - Go was the preferred language
4. **Learning value** - Deep understanding of protocols was valuable
5. **Minimal dependencies** - Wanted control over the codebase

## Why K8s-Native (kmcp/Kagent) is Recommended Now

| Then (Custom) | Now (K8s-Native) |
|---------------|------------------|
| No MCP frameworks | Mature frameworks (mcp-go, FastMCP) |
| No K8s-native patterns | Kagent is CNCF Sandbox |
| DIY everything | CRDs, operators, GitOps-ready |
| Manual RBAC/secrets | Declarative, auto-managed |

## Parallel Implementation Available

This repo now contains **both** implementations:

```
dm-nkp-gitops-custom-mcp-server/
├── cmd/a2a-server/           # Go (Custom) - Production ready
├── pkg/                      # Go tools and handlers
├── chart/                    # Helm chart for Go server
│
└── kmcp-server/              # Python (K8s-Native) - Learning/Future
    ├── src/gitops_mcp/       # FastMCP server + tools
    └── k8s/                  # Kagent MCPServer CRD
```

**Use Custom (Go)** for: Production today, minimal dependencies
**Use kmcp (Python)** for: Learning, multi-agent orchestration, future K8s-native

---

## References

- [MCP Specification](https://spec.modelcontextprotocol.io)
- [A2A Protocol](https://google.github.io/A2A/)
- [mcp-go GitHub](https://github.com/mark3labs/mcp-go)
- [go-mcp GitHub](https://github.com/metoro-io/mcp-golang)
- [kmcp Documentation](https://kagent.dev/docs/kmcp)
- [Kagent GitHub](https://github.com/kagent-dev/kagent)
- [Kagent Lab (Free)](https://kagent.dev/docs/kmcp#kagent-lab-discover-kagent-and-kmcp)
