package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CAPI GVRs
var (
	clusterGVR = schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "clusters",
	}

	machineGVR = schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "machines",
	}

	machineDeploymentGVR = schema.GroupVersionResource{
		Group:    "cluster.x-k8s.io",
		Version:  "v1beta1",
		Resource: "machinedeployments",
	}
)

// handleGetClusterStatus handles the get_cluster_status tool.
func (r *Registry) handleGetClusterStatus(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate input to prevent injection attacks
	if err := validateToolArgs(args); err != nil {
		return nil, err
	}

	clusterName, _ := args["cluster_name"].(string)
	namespace, _ := args["namespace"].(string)

	var sb strings.Builder

	if clusterName != "" {
		// Get specific cluster
		var cluster *unstructured.Unstructured
		var err error

		if namespace != "" {
			cluster, err = r.clients.Dynamic.Resource(clusterGVR).Namespace(namespace).Get(ctx, clusterName, metav1.GetOptions{})
		} else {
			// Search all namespaces
			clusters, err := r.clients.Dynamic.Resource(clusterGVR).List(ctx, metav1.ListOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to list clusters: %w", err)
			}
			for _, c := range clusters.Items {
				if c.GetName() == clusterName {
					cluster = &c
					break
				}
			}
			if cluster == nil {
				return nil, fmt.Errorf("cluster %s not found", clusterName)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get cluster %s: %w", clusterName, err)
		}

		sb.WriteString(formatClusterDetails(cluster))
	} else {
		// List all clusters
		var clusterList *unstructured.UnstructuredList
		var err error

		if namespace != "" {
			clusterList, err = r.clients.Dynamic.Resource(clusterGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
		} else {
			clusterList, err = r.clients.Dynamic.Resource(clusterGVR).List(ctx, metav1.ListOptions{})
		}

		if err != nil {
			return nil, fmt.Errorf("failed to list clusters: %w", err)
		}

		sb.WriteString("# CAPI Clusters\n\n")
		sb.WriteString("| Namespace | Name | Phase | Control Plane | Workers | Infrastructure |\n")
		sb.WriteString("|-----------|------|-------|:-------------:|:-------:|:--------------:|\n")

		for _, cluster := range clusterList.Items {
			phase, _, _ := unstructured.NestedString(cluster.Object, "status", "phase")
			cpReady := getClusterConditionStatus(&cluster, "ControlPlaneReady")
			infraReady := getClusterConditionStatus(&cluster, "InfrastructureReady")

			// Get worker count from MachineDeployments
			workerCount := "-"
			mdList, err := r.clients.Dynamic.Resource(machineDeploymentGVR).Namespace(cluster.GetNamespace()).List(ctx, metav1.ListOptions{
				LabelSelector: fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", cluster.GetName()),
			})
			if err == nil && mdList != nil {
				total := 0
				for _, md := range mdList.Items {
					replicas, _, _ := unstructured.NestedInt64(md.Object, "status", "readyReplicas")
					total += int(replicas)
				}
				workerCount = fmt.Sprintf("%d", total)
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				cluster.GetNamespace(), cluster.GetName(), phase, cpReady, workerCount, infraReady))
		}

		sb.WriteString(fmt.Sprintf("\n**Total:** %d clusters\n", len(clusterList.Items)))
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleListMachines handles the list_machines tool.
func (r *Registry) handleListMachines(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate input to prevent injection attacks
	if err := validateToolArgs(args); err != nil {
		return nil, err
	}

	clusterName, _ := args["cluster_name"].(string)
	namespace, _ := args["namespace"].(string)

	listOptions := metav1.ListOptions{}
	if clusterName != "" {
		// Sanitize cluster name to prevent injection in label selector
		sanitizedClusterName := sanitizeForLogging(clusterName)
		listOptions.LabelSelector = fmt.Sprintf("cluster.x-k8s.io/cluster-name=%s", sanitizedClusterName)
	}

	var machineList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		machineList, err = r.clients.Dynamic.Resource(machineGVR).Namespace(namespace).List(ctx, listOptions)
	} else {
		machineList, err = r.clients.Dynamic.Resource(machineGVR).List(ctx, listOptions)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list machines: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# CAPI Machines\n\n")

	if clusterName != "" {
		// Sanitize user input before including in output
		sanitizedClusterName := sanitizeForLogging(clusterName)
		sb.WriteString(fmt.Sprintf("**Cluster:** %s\n\n", sanitizedClusterName))
	}

	sb.WriteString("| Namespace | Name | Cluster | Phase | Node | Provider ID |\n")
	sb.WriteString("|-----------|------|---------|-------|------|------------|\n")

	for _, machine := range machineList.Items {
		cluster, _, _ := unstructured.NestedString(machine.Object, "metadata", "labels", "cluster.x-k8s.io/cluster-name")
		phase, _, _ := unstructured.NestedString(machine.Object, "status", "phase")
		nodeRef, _, _ := unstructured.NestedString(machine.Object, "status", "nodeRef", "name")
		providerID, _, _ := unstructured.NestedString(machine.Object, "spec", "providerID")

		if nodeRef == "" {
			nodeRef = "-"
		}
		if providerID == "" {
			providerID = "-"
		} else {
			providerID = truncateString(providerID, 30)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			machine.GetNamespace(), machine.GetName(), cluster, phase, nodeRef, providerID))
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d machines\n", len(machineList.Items)))

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// formatClusterDetails formats detailed cluster information.
func formatClusterDetails(cluster *unstructured.Unstructured) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Cluster: %s/%s\n\n", cluster.GetNamespace(), cluster.GetName()))

	// Phase and basic info
	phase, _, _ := unstructured.NestedString(cluster.Object, "status", "phase")
	sb.WriteString("## Status\n\n")
	sb.WriteString(fmt.Sprintf("**Phase:** %s\n\n", phase))

	// Topology info
	if topology, found, _ := unstructured.NestedMap(cluster.Object, "spec", "topology"); found {
		sb.WriteString("## Topology\n\n")
		if class, ok := topology["class"].(string); ok {
			sb.WriteString(fmt.Sprintf("- **ClusterClass:** %s\n", class))
		}
		if version, ok := topology["version"].(string); ok {
			sb.WriteString(fmt.Sprintf("- **Kubernetes Version:** %s\n", version))
		}
		sb.WriteString("\n")
	}

	// Control Plane endpoint
	if cpEndpoint, found, _ := unstructured.NestedMap(cluster.Object, "spec", "controlPlaneEndpoint"); found {
		host, _ := cpEndpoint["host"].(string)
		port, _ := cpEndpoint["port"].(int64)
		if host != "" {
			sb.WriteString("## Control Plane Endpoint\n\n")
			sb.WriteString(fmt.Sprintf("**Endpoint:** %s:%d\n\n", host, port))
		}
	}

	// Conditions
	sb.WriteString("## Conditions\n\n")
	conditions, _, _ := unstructured.NestedSlice(cluster.Object, "status", "conditions")
	if len(conditions) > 0 {
		sb.WriteString("| Type | Status | Reason | Message |\n")
		sb.WriteString("|------|:------:|--------|--------|\n")
		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				condType, _ := cond["type"].(string)
				status, _ := cond["status"].(string)
				reason, _ := cond["reason"].(string)
				message, _ := cond["message"].(string)

				statusIcon := "❓"
				switch status {
				case "True":
					statusIcon = "✅"
				case "False":
					statusIcon = "❌"
				}

				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					condType, statusIcon, reason, truncateString(message, 50)))
			}
		}
	} else {
		sb.WriteString("No conditions available.\n")
	}

	return sb.String()
}

// getClusterConditionStatus returns a status icon for a cluster condition.
func getClusterConditionStatus(cluster *unstructured.Unstructured, conditionType string) string {
	conditions, found, _ := unstructured.NestedSlice(cluster.Object, "status", "conditions")
	if !found {
		return "❓"
	}

	for _, c := range conditions {
		if cond, ok := c.(map[string]interface{}); ok {
			if t, _ := cond["type"].(string); t == conditionType {
				status, _ := cond["status"].(string)
				switch status {
				case "True":
					return "✅"
				case "False":
					return "❌"
				default:
					return "⏳"
				}
			}
		}
	}
	return "❓"
}
