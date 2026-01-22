"""Gatekeeper and Kyverno policy tools for MCP server."""

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


def check_policy_violations(
    namespace: Optional[str] = None,
    policy_engine: str = "both"
) -> str:
    """Check for Gatekeeper or Kyverno policy violations across the cluster.
    
    Args:
        namespace: Namespace to filter (default: all namespaces)
        policy_engine: Policy engine to check: gatekeeper, kyverno, or both
    
    Returns:
        Summary of policy violations
    """
    violations = []
    
    # Check Gatekeeper
    if policy_engine in ["gatekeeper", "both"]:
        violations.extend(_check_gatekeeper_violations(namespace))
    
    # Check Kyverno
    if policy_engine in ["kyverno", "both"]:
        violations.extend(_check_kyverno_violations(namespace))
    
    if not violations:
        return f"""## Policy Violations

**Engine(s) checked:** {policy_engine}
**Result:** ✅ No violations found
"""
    
    rows = []
    for v in violations:
        rows.append(f"| {v['engine']} | {v['policy']} | {v['resource']} | {v['message'][:50]} |")
    
    header = "| Engine | Policy | Resource | Message |\n|--------|--------|----------|---------|"
    return f"""## Policy Violations

**Engine(s) checked:** {policy_engine}
**Violations found:** {len(violations)}

{header}
""" + "\n".join(rows)


def _check_gatekeeper_violations(namespace: Optional[str]) -> list:
    """Check Gatekeeper constraint violations."""
    violations = []
    
    try:
        # Get all constraint templates to find constraint kinds
        ct_api = dyn_client.resources.get(
            api_version="templates.gatekeeper.sh/v1",
            kind="ConstraintTemplate"
        )
        templates = ct_api.get().items
        
        for template in templates:
            kind = template.metadata.name
            
            try:
                # Get constraints of this kind
                constraint_api = dyn_client.resources.get(
                    api_version="constraints.gatekeeper.sh/v1beta1",
                    kind=kind.title().replace("-", "")
                )
                constraints = constraint_api.get().items
                
                for constraint in constraints:
                    # Check for violations in status
                    total_violations = constraint.get("status", {}).get("totalViolations", 0)
                    
                    if total_violations > 0:
                        violation_list = constraint.get("status", {}).get("violations", [])
                        for v in violation_list:
                            if namespace and v.get("namespace") != namespace:
                                continue
                            
                            violations.append({
                                "engine": "Gatekeeper",
                                "policy": constraint.metadata.name,
                                "resource": f"{v.get('kind', '')}/{v.get('name', '')}",
                                "message": v.get("message", "No message")
                            })
            except Exception:
                continue
                
    except Exception:
        pass
    
    return violations


def _check_kyverno_violations(namespace: Optional[str]) -> list:
    """Check Kyverno policy report violations."""
    violations = []
    
    try:
        # Check cluster policy reports
        cpr_api = dyn_client.resources.get(
            api_version="wgpolicyk8s.io/v1alpha2",
            kind="ClusterPolicyReport"
        )
        reports = cpr_api.get().items
        
        for report in reports:
            results = report.get("results", [])
            for r in results:
                if r.get("result") == "fail":
                    violations.append({
                        "engine": "Kyverno",
                        "policy": r.get("policy", "Unknown"),
                        "resource": f"{r.get('resources', [{}])[0].get('kind', '')}/{r.get('resources', [{}])[0].get('name', '')}",
                        "message": r.get("message", "No message")
                    })
    except Exception:
        pass
    
    # Check namespaced policy reports
    try:
        pr_api = dyn_client.resources.get(
            api_version="wgpolicyk8s.io/v1alpha2",
            kind="PolicyReport"
        )
        
        if namespace:
            reports = pr_api.get(namespace=namespace).items
        else:
            reports = pr_api.get().items
        
        for report in reports:
            results = report.get("results", [])
            for r in results:
                if r.get("result") == "fail":
                    violations.append({
                        "engine": "Kyverno",
                        "policy": r.get("policy", "Unknown"),
                        "resource": f"{r.get('resources', [{}])[0].get('kind', '')}/{r.get('resources', [{}])[0].get('name', '')}",
                        "message": r.get("message", "No message")
                    })
    except Exception:
        pass
    
    return violations


def list_constraints(constraint_kind: Optional[str] = None) -> str:
    """List Gatekeeper constraints and their enforcement status.
    
    Args:
        constraint_kind: Filter by constraint kind (e.g., K8sRequiredLabels)
    
    Returns:
        Markdown table of constraints
    """
    constraints = []
    
    try:
        # Get all constraint templates
        ct_api = dyn_client.resources.get(
            api_version="templates.gatekeeper.sh/v1",
            kind="ConstraintTemplate"
        )
        templates = ct_api.get().items
        
        for template in templates:
            kind = template.metadata.name
            
            if constraint_kind and kind.lower() != constraint_kind.lower():
                continue
            
            try:
                constraint_api = dyn_client.resources.get(
                    api_version="constraints.gatekeeper.sh/v1beta1",
                    kind=kind.title().replace("-", "")
                )
                items = constraint_api.get().items
                
                for c in items:
                    name = c.metadata.name
                    enforcement = c.get("spec", {}).get("enforcementAction", "deny")
                    total_violations = c.get("status", {}).get("totalViolations", 0)
                    
                    constraints.append({
                        "kind": kind,
                        "name": name,
                        "enforcement": enforcement,
                        "violations": total_violations
                    })
            except Exception:
                continue
                
    except Exception as e:
        return f"Error: Gatekeeper not installed or no access: {e}"
    
    if not constraints:
        return "No Gatekeeper constraints found"
    
    rows = []
    for c in constraints:
        violation_str = f"❌ {c['violations']}" if c['violations'] > 0 else "✅ 0"
        rows.append(f"| {c['kind']} | {c['name']} | {c['enforcement']} | {violation_str} |")
    
    header = "| Kind | Name | Enforcement | Violations |\n|------|------|-------------|------------|"
    return f"## Gatekeeper Constraints\n\n{header}\n" + "\n".join(rows)
