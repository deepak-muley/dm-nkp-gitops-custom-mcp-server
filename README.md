# dm-nkp-gitops-custom-mcp-server

A custom MCP (Model Context Protocol) server for monitoring and debugging NKP GitOps infrastructure managed by Flux CD.

## Features

- **GitOps Status**: Query Flux Kustomizations, GitRepositories, and HelmReleases
- **Cluster Health**: Check CAPI cluster status and machine health
- **App Deployments**: Monitor Kommander App/ClusterApp deployments
- **Debugging**: Trace reconciliation failures, view events and logs
- **Policy Compliance**: Check Gatekeeper and Kyverno policy violations

## Quick Start

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

| Tool | Description |
|------|-------------|
| `list_contexts` | List available Kubernetes contexts |
| `get_current_context` | Get current active context |
| `get_gitops_status` | Overall GitOps health summary |
| `list_kustomizations` | List Flux Kustomizations with status |
| `get_kustomization` | Get detailed Kustomization info |
| `list_gitrepositories` | List GitRepository sources |
| `get_cluster_status` | CAPI cluster status |
| `list_machines` | List CAPI machines |
| `get_app_deployments` | Kommander app deployment status |
| `get_helmreleases` | List HelmReleases |
| `debug_reconciliation` | Debug failing reconciliation |
| `get_events` | Get Kubernetes events |
| `get_pod_logs` | Get pod logs |
| `check_policy_violations` | Check policy violations |
| `list_constraints` | List Gatekeeper constraints |

## Example Queries

Once configured, ask your AI assistant:

```
"What's the GitOps status of my cluster?"

"Why is the clusterops-clusters Kustomization failing?"

"Show me all CAPI clusters and their health"

"Check for any policy violations"

"Debug the reconciliation failure for clusterops-workspace-applications"
```

## Documentation

- [MCP Primer](docs/MCP_PRIMER.md) - Understanding MCP protocol
- [Flux MCP Setup](docs/FLUX_MCP_SETUP.md) - Using the official Flux MCP Server

## Development

```bash
# Build
make build

# Run tests
make test

# Format code
make fmt

# Lint
make lint

# Test MCP protocol
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/dm-nkp-gitops-mcp-server serve
```

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

## Related Projects

- [dm-nkp-gitops-infra](https://github.com/deepak-muley/dm-nkp-gitops-infra) - GitOps infrastructure repo
- [Flux Operator MCP](https://fluxoperator.dev/docs/mcp/) - Official Flux MCP Server
- [Kagent](https://kagent.dev) - CNCF AI agent framework for Kubernetes

## License

MIT
