# dm-nkp-gitops-custom-mcp-server

A collection of MCP (Model Context Protocol) and A2A (Agent-to-Agent) servers for monitoring and debugging NKP GitOps infrastructure managed by Flux CD.

## Projects in This Repository

This repository contains multiple independent projects, each with its own purpose and implementation:

| Project | Description | Language | Status |
|---------|-------------|----------|--------|
| **[MCP Server](cmd/server/)** | Primary MCP server for AI assistants (Cursor, Claude Desktop) | Go | ✅ Production |
| **[A2A Server](cmd/a2a-server/)** | Agent-to-Agent server for multi-agent coordination | Go | ✅ Production |
| **[kmcp-server](kmcp-server/)** | Kubernetes-native MCP server using kmcp/Kagent | Python | 🔄 Alternative |
| **[Multi-Agent Examples](examples/multi-agent/)** | Demo code for A2A multi-agent workflows | Go | 📚 Examples |
| **[Troubleshooting Examples](examples/troubleshooting/)** | Working code examples for runbook patterns | Go | 📚 Examples |

> 📖 **Start Here**: See [Deployment and Usage Guide](docs/DEPLOYMENT_AND_USAGE.md) for complete setup instructions.

---

## Quick Start: MCP Server (Primary)

The MCP server is the main project for integrating with AI assistants.

### Build

```bash
make build
```

### Run

```bash
# With default kubeconfig
./bin/dm-nkp-gitops-mcp-server serve

# With specific kubeconfig in read-only mode
KUBECONFIG=/path/to/kubeconfig ./bin/dm-nkp-gitops-mcp-server serve --read-only
```

### Configure in Cursor

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dm-nkp-gitops": {
      "command": "/path/to/dm-nkp-gitops-mcp-server",
      "args": ["serve", "--read-only"],
      "env": {
        "KUBECONFIG": "/Users/deepak.muley/ws/nkp/dm-nkp-mgmt-1.conf"
      }
    }
  }
}
```

Restart Cursor after saving.

## Available Tools

| Category | Tool | Description |
|----------|------|-------------|
| **Context** | `list_contexts` | List available Kubernetes contexts |
| | `get_current_context` | Get current active context |
| **GitOps** | `get_gitops_status` | Overall GitOps health summary |
| | `list_kustomizations` | List Flux Kustomizations with status |
| | `get_kustomization` | Get detailed Kustomization info |
| | `list_gitrepositories` | List GitRepository sources |
| **Clusters** | `get_cluster_status` | CAPI cluster status |
| | `list_machines` | List CAPI machines |
| **Apps** | `get_app_deployments` | Kommander app deployment status |
| | `get_helmreleases` | List HelmReleases |
| **Debug** | `debug_reconciliation` | Debug failing reconciliation |
| | `get_events` | Get Kubernetes events |
| | `get_pod_logs` | Get pod logs |
| **Policy** | `check_policy_violations` | Check policy violations |
| | `list_constraints` | List Gatekeeper constraints |

📖 **[Full Tools Reference →](docs/TOOLS_REFERENCE.md)** - Complete documentation with parameters, examples, and use cases

## Example Queries

Once configured, ask your AI assistant:

```
"What's the GitOps status of my cluster?"

"Why is the clusterops-clusters Kustomization failing?"

"Show me all CAPI clusters and their health"

"Check for any policy violations"

"Debug the reconciliation failure for clusterops-workspace-applications"
```

## Project Details

### 1. MCP Server (Primary)

**Location**: [`cmd/server/`](cmd/server/)  
**Binary**: `dm-nkp-gitops-mcp-server`  
**Purpose**: MCP server for AI assistants (Cursor, Claude Desktop, VS Code)

- Uses JSON-RPC 2.0 over stdio
- Exposes 15+ tools for GitOps monitoring
- Read-only mode recommended for production
- See [AGENTS.md](AGENTS.md) for AI agent integration guide

**Quick Start**: `make build && ./bin/dm-nkp-gitops-mcp-server serve --read-only`

---

### 2. A2A Server

**Location**: [`cmd/a2a-server/`](cmd/a2a-server/)  
**Binary**: `dm-nkp-gitops-a2a-server`  
**Purpose**: Agent-to-Agent server for multi-agent coordination

- HTTP-based server (port 8080)
- Task-based execution (stateful)
- Agent Card discovery at `/.well-known/agent.json`
- Shares tool implementations with MCP server via `pkg/tools/`

**Quick Start**: `make build-a2a && ./bin/dm-nkp-gitops-a2a-server serve`

**Documentation**:
- [A2A Learning Guide](docs/A2A_LEARNING_GUIDE.md) - Step-by-step A2A tutorial
- [A2A Protocol](docs/A2A_PROTOCOL.md) - Protocol explanation and MCP comparison
- [Multi-Agent Examples](examples/multi-agent/README.md) - Demo code for orchestrating multiple agents

---

### 3. kmcp-server (Kubernetes-Native)

**Location**: [`kmcp-server/`](kmcp-server/)  
**Language**: Python  
**Purpose**: Alternative implementation using kmcp/Kagent framework

- Kubernetes-native deployment via MCPServer CRD
- Auto-managed RBAC and scaling
- Uses FastMCP SDK
- Parallel implementation with same tool names

**Quick Start**: See [kmcp-server/README.md](kmcp-server/README.md)

**Documentation**:
- [kmcp-server README](kmcp-server/README.md) - Complete setup and deployment guide
- [KMCP Learning Guide](docs/KMCP_LEARNING_GUIDE.md) - Learning path for kmcp framework
- [Framework Comparison](docs/FRAMEWORK_COMPARISON.md) - Custom vs kmcp vs other frameworks

---

### 4. Examples

#### Multi-Agent Examples

**Location**: [`examples/multi-agent/`](examples/multi-agent/)  
**Purpose**: Demo code showing A2A multi-agent coordination

- Orchestrator pattern for coordinating specialized agents
- Example workflows (GitOps health check, incident investigation)
- See [examples/multi-agent/README.md](examples/multi-agent/README.md) for details

**Quick Start**: `make demo-multi-agent` (requires A2A server running)

#### Troubleshooting Examples

**Location**: [`examples/troubleshooting/`](examples/troubleshooting/)  
**Purpose**: Working code examples for programmatic runbooks

- Structured workflow definitions
- Decision tree patterns
- Tool call generation examples
- See [examples/troubleshooting/README.md](examples/troubleshooting/README.md) for details

**Quick Start**: `make build-troubleshooter && ./bin/troubleshooter gitops-failure`

---

## Documentation

### Getting Started

- [**Deployment and Usage Guide**](docs/DEPLOYMENT_AND_USAGE.md) - 🚀 **START HERE** - Complete guide on deployment, configuration, and actual usage
- [**Tools Reference**](docs/TOOLS_REFERENCE.md) - 📖 Complete documentation for all 15 tools with parameters and examples
- [**AGENTS.md**](AGENTS.md) - AI agent integration guide for this repository

### Protocol Guides

- [MCP Primer](docs/MCP_PRIMER.md) - Understanding MCP protocol
- [A2A Learning Guide](docs/A2A_LEARNING_GUIDE.md) - Step-by-step A2A tutorial
- [A2A Protocol](docs/A2A_PROTOCOL.md) - Agent-to-Agent protocol explanation and comparison with MCP
- [KMCP Learning Guide](docs/KMCP_LEARNING_GUIDE.md) - Learning path for kmcp framework
- [Framework Comparison](docs/FRAMEWORK_COMPARISON.md) - Custom vs kmcp vs other frameworks

### Runbooks and Best Practices

- [**K8S Troubleshooting Runbook**](docs/K8S_TROUBLESHOOTING_RUNBOOK.md) - 📋 Comprehensive troubleshooting guide with workflows and examples
- [**K8S Security Runbook**](docs/K8S_SECURITY_RUNBOOK.md) - 🔒 Security assessment and hardening guide for Kubernetes clusters
- [**Runbook Best Practices**](docs/RUNBOOK_BEST_PRACTICES.md) - 📚 Standard practices for creating AI-agent-friendly runbooks (template reference)
- [**Troubleshooting Examples**](examples/troubleshooting/README.md) - Working code examples for runbook patterns
- [**Testing Guide**](examples/troubleshooting/TESTING_GUIDE.md) - 📚 How to build, test, and learn with troubleshooting examples

### Architecture and Advanced Topics

- [MCP Server Architecture](docs/MCP_SERVER_ARCHITECTURE.md) - Deep dive into server flow and recommended features
- [Security](docs/SECURITY.md) - Security analysis and hardening guide
- [Enterprise Deployment](docs/ENTERPRISE_DEPLOYMENT.md) - Production deployment guide
- [**NKP Production Deployment**](docs/NKP_PRODUCTION_DEPLOYMENT.md) - 🚀 **Complete guide for Nutanix Kubernetes Platform**
- [Flux MCP Setup](docs/FLUX_MCP_SETUP.md) - Using the official Flux MCP Server

## Development

### Building All Projects

```bash
# Build MCP server
make build

# Build A2A server
make build-a2a

# Build both servers
make build-both

# Build troubleshooting examples
make build-troubleshooter
```

### Running Tests

```bash
# Run Go unit tests
make test

# Test MCP protocol
make test-mcp

# Test A2A server
make test-a2a-card
make test-a2a-health

# Test troubleshooting workflows
make test-troubleshooter
```

### Code Quality

```bash
# Format code
make fmt

# Lint
make lint
```

See `make help` for all available targets.

## Testing with Kind Cluster

Test the MCP server locally with a kind cluster and Flux CD:

```bash
# Prerequisites: kind, kubectl, flux CLI, jq

# Full setup and test (creates cluster, installs Flux, runs tests)
make kind-all

# Or step by step:
make kind-setup      # Create kind cluster with Flux CD
make kind-test       # Run MCP tests
make kind-interactive # Interactive testing mode
make kind-cleanup    # Delete cluster

# Manual testing
./scripts/test-with-kind.sh all
```

The test script will:
1. Create a kind cluster named `mcp-test`
2. Install Flux CD controllers
3. Deploy sample GitOps resources (Kustomizations, GitRepositories, HelmReleases)
4. Run a suite of MCP protocol tests
5. Print configuration for Cursor integration

## Security

**Always use `--read-only` mode in production** to prevent unintended mutations.

See [Security Documentation](docs/SECURITY.md) for:
- Security analysis and threat model
- RBAC configuration for restricted access
- Recommendations for hardening

## Architecture

```
┌─────────────────┐     stdin/stdout      ┌─────────────────────┐
│  AI Assistant   │◄────────────────────►│  MCP Server         │
│  (Cursor/Claude)│     JSON-RPC 2.0      │                     │
└─────────────────┘                       │  ┌───────────────┐  │
                                          │  │ Tool Handlers │  │
                                          │  └───────┬───────┘  │
                                          │          │          │
                                          │          ▼          │
                                          │  ┌───────────────┐  │
                                          │  │ K8s Clients   │  │
                                          │  └───────┬───────┘  │
                                          └──────────┼──────────┘
                                                     │
                                                     ▼
                                          ┌─────────────────────┐
                                          │  Kubernetes API     │
                                          │  - Flux CRDs        │
                                          │  - CAPI CRDs        │
                                          │  - Kommander CRDs   │
                                          │  - Core resources   │
                                          └─────────────────────┘
```

## Project Independence

Each project in this repository is designed to be **independent** for maintainability:

- **MCP Server** and **A2A Server** share `pkg/tools/` but can be built/deployed separately
- **kmcp-server** is completely independent (Python, separate dependencies)
- **Examples** are standalone demo/learning code
- Each project has its own README with specific instructions

This structure allows:
- Independent versioning and releases
- Separate maintenance cycles
- Easy extraction to separate repositories if needed
- Clear separation of concerns

## Related Projects

- [dm-nkp-gitops-infra](https://github.com/deepak-muley/dm-nkp-gitops-infra) - GitOps infrastructure repo
- [Flux Operator MCP](https://fluxoperator.dev/docs/mcp/) - Official Flux MCP Server
- [Kagent](https://kagent.dev) - CNCF AI agent framework for Kubernetes

## License

MIT
