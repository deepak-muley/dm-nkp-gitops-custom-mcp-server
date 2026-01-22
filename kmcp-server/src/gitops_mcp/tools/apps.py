"""Kommander/NKP application tools for MCP server."""

from typing import Optional
from kubernetes import client, config
from kubernetes.dynamic import DynamicClient

# Initialize K8s clients
try:
    config.load_incluster_config()
except:
    config.load_kube_config()

k8s_client = client.ApiClient()
dyn_client = DynamicClient(k8s_client)


def get_app_deployments(
    workspace: Optional[str] = None,
    app_name: Optional[str] = None
) -> str:
    """Get application deployment status across workspaces.
    
    Args:
        workspace: Workspace name (e.g., dm-dev-workspace). Leave empty for all workspaces.
        app_name: Application name to filter. Leave empty for all apps.
    
    Returns:
        Shows App and ClusterApp resources from Kommander
    """
    results = []
    
    # Try to get Apps
    try:
        app_api = dyn_client.resources.get(
            api_version="apps.kommander.d2iq.io/v1alpha2",
            kind="App"
        )
        
        if workspace:
            items = app_api.get(namespace=workspace).items
        else:
            items = app_api.get().items
        
        for app in items:
            if app_name and app.metadata.name != app_name:
                continue
            
            name = app.metadata.name
            ns = app.metadata.namespace
            
            conditions = app.get("status", {}).get("conditions", [])
            is_ready = any(
                c["type"] == "Ready" and c["status"] == "True"
                for c in conditions
            )
            status = "Ready" if is_ready else "Not Ready"
            
            results.append({
                "type": "App",
                "name": name,
                "namespace": ns,
                "status": status
            })
            
    except Exception as e:
        # Kommander Apps API may not be available
        pass
    
    # Try to get ClusterApps
    try:
        cluster_app_api = dyn_client.resources.get(
            api_version="apps.kommander.d2iq.io/v1alpha2",
            kind="ClusterApp"
        )
        
        items = cluster_app_api.get().items
        
        for app in items:
            if app_name and app.metadata.name != app_name:
                continue
            
            name = app.metadata.name
            ns = app.metadata.namespace
            
            conditions = app.get("status", {}).get("conditions", [])
            is_ready = any(
                c["type"] == "Ready" and c["status"] == "True"
                for c in conditions
            )
            status = "Ready" if is_ready else "Not Ready"
            
            results.append({
                "type": "ClusterApp",
                "name": name,
                "namespace": ns,
                "status": status
            })
            
    except Exception:
        pass
    
    if not results:
        return "No Kommander Apps/ClusterApps found (Kommander may not be installed)"
    
    rows = []
    for r in results:
        rows.append(f"| {r['type']} | {r['name']} | {r['namespace']} | {r['status']} |")
    
    header = "| Type | Name | Namespace | Status |\n|------|------|-----------|--------|"
    return f"## Kommander Applications\n\n{header}\n" + "\n".join(rows)
