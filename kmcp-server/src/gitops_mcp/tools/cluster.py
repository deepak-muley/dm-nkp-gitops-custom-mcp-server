"""Cluster API (CAPI) tools for MCP server."""

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
core_v1 = client.CoreV1Api()


def get_cluster_status(
    cluster_name: Optional[str] = None,
    namespace: Optional[str] = None
) -> str:
    """Get status of CAPI (Cluster API) clusters.
    
    Args:
        cluster_name: Name of the cluster (leave empty for all clusters)
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Cluster status including phase, conditions, and infrastructure status
    """
    try:
        cluster_api = dyn_client.resources.get(
            api_version="cluster.x-k8s.io/v1beta1",
            kind="Cluster"
        )
        
        if namespace:
            items = cluster_api.get(namespace=namespace).items
        else:
            items = cluster_api.get().items
    except Exception as e:
        return f"Error: CAPI not installed or no access: {e}"
    
    rows = []
    for c in items:
        if cluster_name and c.metadata.name != cluster_name:
            continue
        
        name = c.metadata.name
        ns = c.metadata.namespace
        phase = c.get("status", {}).get("phase", "Unknown")
        
        # Get infrastructure ready status
        infra_ready = c.get("status", {}).get("infrastructureReady", False)
        cp_ready = c.get("status", {}).get("controlPlaneReady", False)
        
        status_icon = "✅" if phase == "Provisioned" else "⏳" if phase == "Provisioning" else "❌"
        
        rows.append(f"| {status_icon} | {name} | {ns} | {phase} | {infra_ready} | {cp_ready} |")
    
    if not rows:
        return "No CAPI clusters found"
    
    header = "| | Name | Namespace | Phase | Infra Ready | CP Ready |\n|---|------|-----------|-------|-------------|----------|"
    return f"## CAPI Clusters\n\n{header}\n" + "\n".join(rows)


def list_machines(
    cluster_name: Optional[str] = None,
    namespace: Optional[str] = None
) -> str:
    """List CAPI Machines for a cluster showing node status and provider info.
    
    Args:
        cluster_name: Name of the cluster to filter machines
        namespace: Namespace to filter (default: all namespaces)
    
    Returns:
        Markdown table of machines
    """
    try:
        machine_api = dyn_client.resources.get(
            api_version="cluster.x-k8s.io/v1beta1",
            kind="Machine"
        )
        
        if namespace:
            items = machine_api.get(namespace=namespace).items
        else:
            items = machine_api.get().items
    except Exception as e:
        return f"Error: {e}"
    
    rows = []
    for m in items:
        # Filter by cluster if specified
        m_cluster = m.metadata.labels.get("cluster.x-k8s.io/cluster-name", "")
        if cluster_name and m_cluster != cluster_name:
            continue
        
        name = m.metadata.name
        ns = m.metadata.namespace
        phase = m.get("status", {}).get("phase", "Unknown")
        node_ref = m.get("status", {}).get("nodeRef", {}).get("name", "N/A")
        
        rows.append(f"| {name} | {ns} | {m_cluster} | {phase} | {node_ref} |")
    
    if not rows:
        return "No CAPI machines found"
    
    header = "| Name | Namespace | Cluster | Phase | Node |\n|------|-----------|---------|-------|------|"
    return f"## CAPI Machines\n\n{header}\n" + "\n".join(rows)


def get_events(
    namespace: str,
    resource_name: Optional[str] = None,
    event_type: str = "all",
    limit: str = "20"
) -> str:
    """Get Kubernetes events for debugging.
    
    Args:
        namespace: Namespace to get events from (required)
        resource_name: Filter events for a specific resource name
        event_type: Filter by event type: all, Normal, Warning
        limit: Maximum number of events to return (default: 20)
    
    Returns:
        Markdown table of events
    """
    try:
        events = core_v1.list_namespaced_event(namespace=namespace)
    except Exception as e:
        return f"Error: {e}"
    
    # Sort by last timestamp (newest first)
    sorted_events = sorted(
        events.items,
        key=lambda e: e.last_timestamp or e.event_time or "",
        reverse=True
    )
    
    rows = []
    count = 0
    max_count = int(limit)
    
    for e in sorted_events:
        if count >= max_count:
            break
        
        # Filter by resource name
        if resource_name and e.involved_object.name != resource_name:
            continue
        
        # Filter by event type
        if event_type != "all" and e.type != event_type:
            continue
        
        kind = e.involved_object.kind
        name = e.involved_object.name
        reason = e.reason or ""
        message = (e.message or "")[:60]
        etype = e.type
        
        type_icon = "⚠️" if etype == "Warning" else "ℹ️"
        
        rows.append(f"| {type_icon} | {kind}/{name} | {reason} | {message} |")
        count += 1
    
    if not rows:
        return f"No events found in namespace {namespace}"
    
    header = "| Type | Resource | Reason | Message |\n|------|----------|--------|---------|"
    return f"## Events in {namespace}\n\n{header}\n" + "\n".join(rows)


def get_pod_logs(
    pod_name: str,
    namespace: str,
    container: Optional[str] = None,
    tail_lines: str = "100"
) -> str:
    """Get logs from a pod for debugging.
    
    Args:
        pod_name: Name of the pod
        namespace: Namespace of the pod
        container: Container name (optional, uses first container if not specified)
        tail_lines: Number of lines to return from end (default: 100)
    
    Returns:
        Pod logs
    """
    try:
        logs = core_v1.read_namespaced_pod_log(
            name=pod_name,
            namespace=namespace,
            container=container,
            tail_lines=int(tail_lines)
        )
    except Exception as e:
        return f"Error getting logs: {e}"
    
    container_str = f" (container: {container})" if container else ""
    
    return f"""## Pod Logs: {pod_name}{container_str}

**Namespace:** {namespace}
**Lines:** {tail_lines}

```
{logs}
```
"""
