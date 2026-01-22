"""Flux/GitOps tools for MCP server."""

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


def get_gitops_status(namespace: Optional[str] = None) -> str:
    """Get overall GitOps status including all Flux Kustomizations and GitRepositories.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Summary of healthy/unhealthy/suspended resources
    """
    results = {"kustomizations": {"ready": 0, "failed": 0, "suspended": 0}}
    
    try:
        kust_api = dyn_client.resources.get(
            api_version="kustomize.toolkit.fluxcd.io/v1",
            kind="Kustomization"
        )
        
        if namespace:
            items = kust_api.get(namespace=namespace).items
        else:
            items = kust_api.get().items
        
        for k in items:
            conditions = k.get("status", {}).get("conditions", [])
            is_suspended = k.get("spec", {}).get("suspend", False)
            is_ready = any(
                c["type"] == "Ready" and c["status"] == "True" 
                for c in conditions
            )
            
            if is_suspended:
                results["kustomizations"]["suspended"] += 1
            elif is_ready:
                results["kustomizations"]["ready"] += 1
            else:
                results["kustomizations"]["failed"] += 1
                
    except Exception as e:
        return f"Error querying Flux resources: {e}"
    
    total = sum(results["kustomizations"].values())
    r = results["kustomizations"]
    
    return f"""## GitOps Status Summary

| Resource | Ready | Failed | Suspended | Total |
|----------|-------|--------|-----------|-------|
| Kustomizations | {r["ready"]} | {r["failed"]} | {r["suspended"]} | {total} |

**Health:** {"✅ Healthy" if r["failed"] == 0 else "❌ Issues Detected"}
"""


def list_kustomizations(
    namespace: Optional[str] = None,
    status_filter: str = "all"
) -> str:
    """List all Flux Kustomizations with their reconciliation status.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
        status_filter: Filter by status: all, ready, failed, suspended
    
    Returns:
        Markdown table of Kustomizations
    """
    try:
        kust_api = dyn_client.resources.get(
            api_version="kustomize.toolkit.fluxcd.io/v1",
            kind="Kustomization"
        )
        
        if namespace:
            items = kust_api.get(namespace=namespace).items
        else:
            items = kust_api.get().items
    except Exception as e:
        return f"Error: {e}"
    
    rows = []
    for k in items:
        name = k.metadata.name
        ns = k.metadata.namespace
        conditions = k.get("status", {}).get("conditions", [])
        
        is_suspended = k.get("spec", {}).get("suspend", False)
        is_ready = any(
            c["type"] == "Ready" and c["status"] == "True" 
            for c in conditions
        )
        
        if is_suspended:
            status = "Suspended"
        elif is_ready:
            status = "Ready"
        else:
            status = "Failed"
        
        if status_filter != "all" and status.lower() != status_filter.lower():
            continue
        
        source = k.get("spec", {}).get("sourceRef", {})
        source_str = f"{source.get('kind', '')}/{source.get('name', '')}"
        
        rows.append(f"| {name} | {ns} | {status} | {source_str} |")
    
    if not rows:
        return "No Kustomizations found"
    
    header = "| Name | Namespace | Status | Source |\n|------|-----------|--------|--------|"
    return f"## Flux Kustomizations\n\n{header}\n" + "\n".join(rows)


def get_kustomization(name: str, namespace: str) -> str:
    """Get detailed information about a specific Flux Kustomization.
    
    Args:
        name: Name of the Kustomization
        namespace: Namespace of the Kustomization
    
    Returns:
        Detailed Kustomization information including conditions
    """
    try:
        kust_api = dyn_client.resources.get(
            api_version="kustomize.toolkit.fluxcd.io/v1",
            kind="Kustomization"
        )
        k = kust_api.get(name=name, namespace=namespace)
    except Exception as e:
        return f"Error: {e}"
    
    spec = k.get("spec", {})
    status = k.get("status", {})
    conditions = status.get("conditions", [])
    
    output = f"""## Kustomization: {name}

**Namespace:** {namespace}
**Path:** {spec.get("path", "N/A")}
**Interval:** {spec.get("interval", "N/A")}
**Suspended:** {spec.get("suspend", False)}

### Source
- **Kind:** {spec.get("sourceRef", {}).get("kind", "N/A")}
- **Name:** {spec.get("sourceRef", {}).get("name", "N/A")}

### Conditions

| Type | Status | Reason | Message |
|------|--------|--------|---------|
"""
    
    for c in conditions:
        msg = c.get("message", "")[:60]
        output += f"| {c['type']} | {c['status']} | {c.get('reason', '')} | {msg} |\n"
    
    return output


def list_gitrepositories(namespace: Optional[str] = None) -> str:
    """List all Flux GitRepository sources with their sync status.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Markdown table of GitRepositories
    """
    try:
        gr_api = dyn_client.resources.get(
            api_version="source.toolkit.fluxcd.io/v1",
            kind="GitRepository"
        )
        
        if namespace:
            items = gr_api.get(namespace=namespace).items
        else:
            items = gr_api.get().items
    except Exception as e:
        return f"Error: {e}"
    
    rows = []
    for gr in items:
        name = gr.metadata.name
        ns = gr.metadata.namespace
        url = gr.get("spec", {}).get("url", "N/A")
        branch = gr.get("spec", {}).get("ref", {}).get("branch", "N/A")
        
        conditions = gr.get("status", {}).get("conditions", [])
        is_ready = any(
            c["type"] == "Ready" and c["status"] == "True" 
            for c in conditions
        )
        status = "Ready" if is_ready else "Failed"
        
        rows.append(f"| {name} | {ns} | {status} | {branch} |")
    
    if not rows:
        return "No GitRepositories found"
    
    header = "| Name | Namespace | Status | Branch |\n|------|-----------|--------|--------|"
    return f"## GitRepositories\n\n{header}\n" + "\n".join(rows)


def get_helmreleases(
    namespace: Optional[str] = None,
    status_filter: str = "all"
) -> str:
    """List Flux HelmReleases with their status.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
        status_filter: Filter by status: all, ready, failed, suspended
    
    Returns:
        Markdown table of HelmReleases
    """
    try:
        hr_api = dyn_client.resources.get(
            api_version="helm.toolkit.fluxcd.io/v2",
            kind="HelmRelease"
        )
        
        if namespace:
            items = hr_api.get(namespace=namespace).items
        else:
            items = hr_api.get().items
    except Exception as e:
        return f"Error: {e}"
    
    rows = []
    for hr in items:
        name = hr.metadata.name
        ns = hr.metadata.namespace
        
        conditions = hr.get("status", {}).get("conditions", [])
        is_suspended = hr.get("spec", {}).get("suspend", False)
        is_ready = any(
            c["type"] == "Ready" and c["status"] == "True" 
            for c in conditions
        )
        
        if is_suspended:
            status = "Suspended"
        elif is_ready:
            status = "Ready"
        else:
            status = "Failed"
        
        if status_filter != "all" and status.lower() != status_filter.lower():
            continue
        
        chart = hr.get("spec", {}).get("chart", {}).get("spec", {})
        chart_name = chart.get("chart", "N/A")
        
        rows.append(f"| {name} | {ns} | {status} | {chart_name} |")
    
    if not rows:
        return "No HelmReleases found"
    
    header = "| Name | Namespace | Status | Chart |\n|------|-----------|--------|-------|"
    return f"## HelmReleases\n\n{header}\n" + "\n".join(rows)


def debug_reconciliation(
    resource_type: str,
    name: str,
    namespace: str
) -> str:
    """Debug a failing Flux reconciliation.
    
    Args:
        resource_type: Type of resource: kustomization, gitrepository, helmrelease
        name: Name of the resource
        namespace: Namespace of the resource
    
    Returns:
        Detailed debug information including conditions and events
    """
    api_map = {
        "kustomization": ("kustomize.toolkit.fluxcd.io/v1", "Kustomization"),
        "gitrepository": ("source.toolkit.fluxcd.io/v1", "GitRepository"),
        "helmrelease": ("helm.toolkit.fluxcd.io/v2", "HelmRelease"),
    }
    
    if resource_type.lower() not in api_map:
        return f"Unknown resource type: {resource_type}. Use: kustomization, gitrepository, helmrelease"
    
    api_version, kind = api_map[resource_type.lower()]
    
    try:
        resource_api = dyn_client.resources.get(api_version=api_version, kind=kind)
        resource = resource_api.get(name=name, namespace=namespace)
    except Exception as e:
        return f"Error: {e}"
    
    conditions = resource.get("status", {}).get("conditions", [])
    
    output = f"""## Debug: {kind}/{name}

**Namespace:** {namespace}

### Conditions

| Type | Status | Reason | Last Transition | Message |
|------|--------|--------|-----------------|---------|
"""
    
    for c in conditions:
        msg = c.get("message", "")[:50] + "..." if len(c.get("message", "")) > 50 else c.get("message", "")
        output += f"| {c['type']} | {c['status']} | {c.get('reason', '')} | {c.get('lastTransitionTime', '')[:19]} | {msg} |\n"
    
    # Add recommendations
    output += "\n### Recommendations\n\n"
    
    for c in conditions:
        if c["status"] == "False" and c["type"] == "Ready":
            reason = c.get("reason", "")
            if "Source" in reason:
                output += "- Check if the source (GitRepository/HelmRepository) exists and is ready\n"
            if "Validation" in reason:
                output += "- Check the manifest syntax and Kubernetes API compatibility\n"
            if "Health" in reason:
                output += "- Check if deployed resources are healthy (pods running, etc.)\n"
    
    return output
