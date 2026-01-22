"""
FastMCP server for NKP GitOps monitoring.

This is the K8s-native equivalent of the Go-based dm-nkp-gitops-a2a-server.
Uses FastMCP for MCP protocol and integrates with Kagent for A2A support.
"""

from mcp.server.fastmcp import FastMCP
from typing import Optional
import logging

# Import tool modules
from gitops_mcp.tools import flux, cluster, apps, policy

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Initialize FastMCP server
mcp = FastMCP(
    name="dm-nkp-gitops-mcp",
    version="0.1.0",
    description="K8s-native MCP server for NKP GitOps infrastructure monitoring and debugging",
)


# =============================================================================
# Context Tools
# =============================================================================

@mcp.tool()
def list_contexts() -> str:
    """List all available Kubernetes contexts from kubeconfig.
    
    Returns:
        Markdown table of available contexts
    """
    from kubernetes import config
    
    try:
        contexts, active_context = config.list_kube_config_contexts()
    except Exception as e:
        return f"Error loading kubeconfig: {e}"
    
    rows = []
    for ctx in contexts:
        name = ctx["name"]
        cluster = ctx["context"].get("cluster", "")
        user = ctx["context"].get("user", "")
        is_current = "→" if name == active_context["name"] else ""
        rows.append(f"| {is_current} | {name} | {cluster} | {user} |")
    
    header = "| Current | Name | Cluster | User |\n|---------|------|---------|------|"
    return f"## Kubernetes Contexts\n\n{header}\n" + "\n".join(rows)


@mcp.tool()
def get_current_context() -> str:
    """Get the current active Kubernetes context.
    
    Returns:
        Current context information
    """
    from kubernetes import config
    
    try:
        _, active_context = config.list_kube_config_contexts()
        return f"""## Current Context

**Name:** {active_context["name"]}
**Cluster:** {active_context["context"].get("cluster", "N/A")}
**User:** {active_context["context"].get("user", "N/A")}
**Namespace:** {active_context["context"].get("namespace", "default")}
"""
    except Exception as e:
        return f"Error: {e}"


# =============================================================================
# Register tools from modules
# =============================================================================

# Flux/GitOps tools
mcp.tool()(flux.get_gitops_status)
mcp.tool()(flux.list_kustomizations)
mcp.tool()(flux.get_kustomization)
mcp.tool()(flux.list_gitrepositories)
mcp.tool()(flux.get_helmreleases)

# Cluster tools
mcp.tool()(cluster.get_cluster_status)
mcp.tool()(cluster.list_machines)

# App tools
mcp.tool()(apps.get_app_deployments)

# Debug tools
mcp.tool()(flux.debug_reconciliation)
mcp.tool()(cluster.get_events)
mcp.tool()(cluster.get_pod_logs)

# Policy tools
mcp.tool()(policy.check_policy_violations)
mcp.tool()(policy.list_constraints)


def main():
    """Entry point for the MCP server."""
    logger.info("Starting dm-nkp-gitops-mcp server (K8s-native)")
    mcp.run()


if __name__ == "__main__":
    main()
