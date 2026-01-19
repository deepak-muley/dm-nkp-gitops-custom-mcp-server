# AGENTS.md - AI Agent Guide for dm-nkp-gitops-custom-mcp-server

## Project Overview

This is a **custom MCP (Model Context Protocol) server** for monitoring and debugging NKP GitOps infrastructure managed by Flux CD.

### Purpose
- Provide AI assistants (Cursor, Claude Desktop, VS Code) with tools to query GitOps status
- Debug Flux reconciliation failures
- Check CAPI cluster health
- Monitor application deployments across workspaces
- Check policy violations (Gatekeeper/Kyverno)

### Technology Stack
- **Language:** Go 1.22+
- **Protocol:** MCP (Model Context Protocol) via JSON-RPC 2.0 over stdio
- **Kubernetes:** client-go, dynamic client for CRDs
- **Target:** Flux CD, Cluster API, Kommander/NKP

---

## Project Structure

```
dm-nkp-gitops-custom-mcp-server/
├── cmd/server/main.go       # Entry point, CLI handling
├── pkg/
│   ├── mcp/
│   │   ├── server.go        # MCP protocol handling (JSON-RPC)
│   │   └── types.go         # MCP types (Tool, Resource, etc.)
│   ├── config/
│   │   ├── config.go        # Kubeconfig, CLI flags
│   │   └── logger.go        # Structured logging to stderr
│   └── tools/
│       ├── registry.go      # Tool registration
│       ├── context_handlers.go   # Kubernetes context tools
│       ├── flux_handlers.go      # Flux/GitOps tools
│       ├── cluster_handlers.go   # CAPI cluster tools
│       ├── app_handlers.go       # App deployment tools
│       ├── debug_handlers.go     # Debugging tools
│       └── policy_handlers.go    # Gatekeeper/Kyverno tools
├── docs/
│   ├── MCP_PRIMER.md        # MCP protocol documentation
│   └── FLUX_MCP_SETUP.md    # Flux MCP Server setup guide
├── go.mod
├── Makefile
└── README.md
```

---

## Key Concepts

### MCP Protocol
- Uses **JSON-RPC 2.0** over **stdio** (stdin/stdout)
- Server exposes **Tools** that AI can invoke
- Communication: Client sends JSON request → Server processes → Server returns JSON response

### Tool Structure
```go
type Tool struct {
    Name        string      // Tool identifier (e.g., "get_gitops_status")
    Description string      // What the tool does
    InputSchema InputSchema // JSON Schema for arguments
}
```

### Adding a New Tool
1. Define the tool in `pkg/tools/registry.go` in the appropriate `register*Tools()` function
2. Implement the handler function in the appropriate `*_handlers.go` file
3. Handler signature: `func(args map[string]interface{}) (*mcp.ToolCallResult, error)`

---

## Available Tools

| Tool | Description |
|------|-------------|
| `list_contexts` | List Kubernetes contexts |
| `get_current_context` | Get current context |
| `get_gitops_status` | Overall GitOps health summary |
| `list_kustomizations` | List Flux Kustomizations |
| `get_kustomization` | Get Kustomization details |
| `list_gitrepositories` | List GitRepository sources |
| `get_cluster_status` | CAPI cluster status |
| `list_machines` | List CAPI machines |
| `get_app_deployments` | Kommander App/ClusterApp status |
| `get_helmreleases` | List HelmReleases |
| `debug_reconciliation` | Debug failing reconciliation |
| `get_events` | Get Kubernetes events |
| `get_pod_logs` | Get pod logs |
| `check_policy_violations` | Check Gatekeeper/Kyverno violations |
| `list_constraints` | List Gatekeeper constraints |

---

## Development Commands

```bash
# Build
make build

# Run locally
make run

# Run in read-only mode
make run-readonly

# Test MCP protocol
make test-mcp

# Install to PATH
make install
```

---

## Kubernetes Resources Accessed

### Flux CD
- `kustomizations.kustomize.toolkit.fluxcd.io/v1`
- `gitrepositories.source.toolkit.fluxcd.io/v1`
- `helmreleases.helm.toolkit.fluxcd.io/v2`

### Cluster API
- `clusters.cluster.x-k8s.io/v1beta1`
- `machines.cluster.x-k8s.io/v1beta1`
- `machinedeployments.cluster.x-k8s.io/v1beta1`

### Kommander/NKP
- `apps.apps.kommander.d2iq.io/v1alpha2`
- `clusterapps.apps.kommander.d2iq.io/v1alpha2`

### Gatekeeper
- `constrainttemplates.templates.gatekeeper.sh/v1`
- `constraints.gatekeeper.sh/v1beta1/*`

### Kyverno
- `clusterpolicies.kyverno.io/v1`
- `clusterpolicyreports.wgpolicyk8s.io/v1alpha2`

---

## Configuration

### Environment Variables
- `KUBECONFIG` - Path to kubeconfig file
- `MCP_READ_ONLY` - Set to "true" for read-only mode
- `MCP_LOG_LEVEL` - Log level: debug, info, warn, error

### CLI Flags
```bash
dm-nkp-gitops-mcp-server serve [flags]
  --kubeconfig string   Path to kubeconfig
  --context string      Kubernetes context
  --read-only           Enable read-only mode
  --log-level string    Log level (default: info)
```

---

## Cursor/Claude Integration

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

---

## Related Projects

- **GitOps Repo:** `github.com/deepak-muley/dm-nkp-gitops-infra`
- **Flux MCP Server:** `github.com/controlplaneio-fluxcd/flux-operator`
- **Kagent:** `kagent.dev` (CNCF Sandbox)

---

## Important Notes

1. **Logging goes to stderr** - stdout is reserved for MCP protocol
2. **Read-only mode recommended** - Prevents accidental mutations
3. **Dynamic client** - Used for CRDs (Flux, CAPI, Kommander)
4. **Timeout:** All K8s API calls have 30s timeout
5. **Output format:** Markdown tables for structured data
