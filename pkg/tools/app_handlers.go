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

// Kommander App GVRs
var (
	appGVR = schema.GroupVersionResource{
		Group:    "apps.kommander.d2iq.io",
		Version:  "v1alpha2",
		Resource: "apps",
	}

	clusterAppGVR = schema.GroupVersionResource{
		Group:    "apps.kommander.d2iq.io",
		Version:  "v1alpha2",
		Resource: "clusterapps",
	}
)

// handleGetAppDeployments handles the get_app_deployments tool.
func (r *Registry) handleGetAppDeployments(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workspace, _ := args["workspace"].(string)
	appName, _ := args["app_name"].(string)

	var sb strings.Builder
	sb.WriteString("# Application Deployments\n\n")

	// Get ClusterApps (workspace-level apps)
	sb.WriteString("## ClusterApps (Workspace Level)\n\n")

	var caList *unstructured.UnstructuredList
	var err error

	if workspace != "" {
		caList, err = r.clients.Dynamic.Resource(clusterAppGVR).Namespace(workspace).List(ctx, metav1.ListOptions{})
	} else {
		caList, err = r.clients.Dynamic.Resource(clusterAppGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Error fetching ClusterApps: %s\n\n", err))
	} else if len(caList.Items) == 0 {
		sb.WriteString("No ClusterApps found.\n\n")
	} else {
		sb.WriteString("| Workspace | Name | Status | Clusters | Message |\n")
		sb.WriteString("|-----------|------|:------:|:--------:|--------|\n")

		for _, ca := range caList.Items {
			name := ca.GetName()
			if appName != "" && !strings.Contains(name, appName) {
				continue
			}

			status := getAppStatus(&ca)
			statusIcon := getStatusIcon(status)
			clusterCount := getDeployedClusterCount(&ca)
			message := truncateString(getConditionMessage(&ca, "Ready"), 40)

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
				ca.GetNamespace(), name, statusIcon, clusterCount, message))
		}
		sb.WriteString("\n")
	}

	// Get Apps (project-level apps)
	sb.WriteString("## Apps (Project Level)\n\n")

	var appList *unstructured.UnstructuredList

	if workspace != "" {
		appList, err = r.clients.Dynamic.Resource(appGVR).Namespace(workspace).List(ctx, metav1.ListOptions{})
	} else {
		appList, err = r.clients.Dynamic.Resource(appGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Error fetching Apps: %s\n\n", err))
	} else if len(appList.Items) == 0 {
		sb.WriteString("No Apps found.\n\n")
	} else {
		sb.WriteString("| Namespace | Name | Status | Version | Message |\n")
		sb.WriteString("|-----------|------|:------:|---------|--------|\n")

		for _, app := range appList.Items {
			name := app.GetName()
			if appName != "" && !strings.Contains(name, appName) {
				continue
			}

			status := getAppStatus(&app)
			statusIcon := getStatusIcon(status)
			version := getAppVersion(&app)
			message := truncateString(getConditionMessage(&app, "Ready"), 40)

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				app.GetNamespace(), name, statusIcon, version, message))
		}
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// getAppStatus returns the status of an App/ClusterApp.
func getAppStatus(obj *unstructured.Unstructured) string {
	// Check Ready condition
	conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found {
		return "Unknown"
	}

	for _, c := range conditions {
		if cond, ok := c.(map[string]interface{}); ok {
			if condType, _ := cond["type"].(string); condType == "Ready" {
				status, _ := cond["status"].(string)
				if status == "True" {
					return "Ready"
				}
				reason, _ := cond["reason"].(string)
				return reason
			}
		}
	}
	return "Unknown"
}

// getStatusIcon returns an icon for the status.
func getStatusIcon(status string) string {
	switch status {
	case "Ready":
		return "✅"
	case "Unknown":
		return "❓"
	case "Progressing", "Pending":
		return "⏳"
	default:
		return "❌"
	}
}

// getDeployedClusterCount returns the number of clusters an app is deployed to.
func getDeployedClusterCount(obj *unstructured.Unstructured) int {
	// Try to get cluster statuses
	clusterStatuses, found, _ := unstructured.NestedMap(obj.Object, "status", "clusterStatuses")
	if found {
		return len(clusterStatuses)
	}
	return 0
}

// getAppVersion returns the app version.
func getAppVersion(obj *unstructured.Unstructured) string {
	version, found, _ := unstructured.NestedString(obj.Object, "spec", "version")
	if found {
		return version
	}
	return "-"
}
