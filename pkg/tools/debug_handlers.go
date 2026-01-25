package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// handleDebugReconciliation handles the debug_reconciliation tool.
func (r *Registry) handleDebugReconciliation(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resourceType, ok := args["resource_type"].(string)
	if !ok {
		return nil, fmt.Errorf("resource_type is required")
	}

	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}

	namespace, ok := args["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("namespace is required")
	}

	// Validate input to prevent injection attacks
	if err := validateToolArgs(args); err != nil {
		return nil, err
	}

	var gvr schema.GroupVersionResource
	switch resourceType {
	case "kustomization":
		gvr = kustomizationGVR
	case "gitrepository":
		gvr = gitRepositoryGVR
	case "helmrelease":
		gvr = helmReleaseGVR
	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}

	// Get the resource
	resource, err := r.clients.Dynamic.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get %s %s/%s: %w", resourceType, namespace, name, err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Debug: %s %s/%s\n\n", strings.Title(resourceType), namespace, name))

	// Overall status
	ready := isResourceReady(resource)
	suspended := isResourceSuspended(resource)

	sb.WriteString("## Status Summary\n\n")
	if suspended {
		sb.WriteString("⏸️ **Status:** Suspended (reconciliation paused)\n\n")
	} else if ready {
		sb.WriteString("✅ **Status:** Ready\n\n")
	} else {
		sb.WriteString("❌ **Status:** NOT Ready - See conditions below\n\n")
	}

	// All conditions
	sb.WriteString("## Conditions\n\n")
	conditions, _, _ := unstructured.NestedSlice(resource.Object, "status", "conditions")
	if len(conditions) > 0 {
		sb.WriteString("| Type | Status | Reason | Last Transition | Message |\n")
		sb.WriteString("|------|:------:|--------|-----------------|--------|\n")

		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				condType, _ := cond["type"].(string)
				status, _ := cond["status"].(string)
				reason, _ := cond["reason"].(string)
				message, _ := cond["message"].(string)
				lastTransition, _ := cond["lastTransitionTime"].(string)

				statusIcon := "❓"
				switch status {
				case "True":
					statusIcon = "✅"
				case "False":
					statusIcon = "❌"
				}

				// Format time
				if t, err := time.Parse(time.RFC3339, lastTransition); err == nil {
					lastTransition = t.Format("2006-01-02 15:04:05")
				}

				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					condType, statusIcon, reason, lastTransition, truncateString(message, 50)))
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No conditions found.\n\n")
	}

	// Source reference (for Kustomizations)
	if resourceType == "kustomization" {
		if sourceRef, found, _ := unstructured.NestedMap(resource.Object, "spec", "sourceRef"); found {
			sb.WriteString("## Source Reference\n\n")
			kind, _ := sourceRef["kind"].(string)
			srcName, _ := sourceRef["name"].(string)
			srcNs, _ := sourceRef["namespace"].(string)
			if srcNs == "" {
				srcNs = namespace
			}

			sb.WriteString(fmt.Sprintf("- **Kind:** %s\n", kind))
			sb.WriteString(fmt.Sprintf("- **Name:** %s/%s\n\n", srcNs, srcName))

			// Get source status
			if kind == "GitRepository" {
				src, err := r.clients.Dynamic.Resource(gitRepositoryGVR).Namespace(srcNs).Get(ctx, srcName, metav1.GetOptions{})
				if err == nil {
					srcReady := isResourceReady(src)
					if srcReady {
						sb.WriteString("✅ Source is ready\n\n")
					} else {
						sb.WriteString("❌ **Source is NOT ready** - This may be causing the failure\n")
						sb.WriteString(fmt.Sprintf("   Message: %s\n\n", getConditionMessage(src, "Ready")))
					}
				}
			}
		}

		// Check dependencies
		if deps, found, _ := unstructured.NestedSlice(resource.Object, "spec", "dependsOn"); found && len(deps) > 0 {
			sb.WriteString("## Dependencies\n\n")
			sb.WriteString("| Dependency | Status |\n")
			sb.WriteString("|------------|:------:|\n")

			for _, dep := range deps {
				if depMap, ok := dep.(map[string]interface{}); ok {
					depName, _ := depMap["name"].(string)
					depNs, _ := depMap["namespace"].(string)
					if depNs == "" {
						depNs = namespace
					}

					// Check dependency status
					depKs, err := r.clients.Dynamic.Resource(kustomizationGVR).Namespace(depNs).Get(ctx, depName, metav1.GetOptions{})
					depStatus := "❓"
					if err == nil {
						if isResourceReady(depKs) {
							depStatus = "✅"
						} else {
							depStatus = "❌"
						}
					}

					sb.WriteString(fmt.Sprintf("| %s/%s | %s |\n", depNs, depName, depStatus))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Get related events
	sb.WriteString("## Recent Events\n\n")
	// Sanitize name to prevent injection in field selector
	sanitizedName := sanitizeForLogging(name)
	events, err := r.clients.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", sanitizedName),
	})

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Error fetching events: %s\n", err))
	} else if len(events.Items) == 0 {
		sb.WriteString("No recent events found.\n")
	} else {
		// Sort by last timestamp
		sort.Slice(events.Items, func(i, j int) bool {
			return events.Items[i].LastTimestamp.After(events.Items[j].LastTimestamp.Time)
		})

		sb.WriteString("| Type | Reason | Age | Message |\n")
		sb.WriteString("|------|--------|-----|--------|\n")

		// Show last 10 events
		count := 10
		if len(events.Items) < count {
			count = len(events.Items)
		}

		for _, event := range events.Items[:count] {
			eventType := "ℹ️"
			if event.Type == "Warning" {
				eventType = "⚠️"
			}
			age := formatAge(event.LastTimestamp.Time)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				eventType, event.Reason, age, truncateString(event.Message, 60)))
		}
	}

	// Suggestions
	if !ready && !suspended {
		sb.WriteString("\n## Troubleshooting Suggestions\n\n")
		message := getConditionMessage(resource, "Ready")

		if strings.Contains(message, "Source") || strings.Contains(message, "source") {
			sb.WriteString("1. Check the source (GitRepository/HelmRepository) status\n")
			sb.WriteString("2. Verify the repository URL and credentials\n")
			sb.WriteString("3. Check network connectivity to the git server\n")
		}

		if strings.Contains(message, "dependency") || strings.Contains(message, "Dependency") {
			sb.WriteString("1. Check the status of dependent Kustomizations\n")
			sb.WriteString("2. Ensure dependencies are in the correct order\n")
			sb.WriteString("3. Look for circular dependencies\n")
		}

		if strings.Contains(message, "validation") || strings.Contains(message, "invalid") {
			sb.WriteString("1. Run `kustomize build` locally to validate the manifests\n")
			sb.WriteString("2. Check for YAML syntax errors\n")
			sb.WriteString("3. Verify all referenced resources exist\n")
		}

		if strings.Contains(message, "secret") || strings.Contains(message, "Secret") {
			sb.WriteString("1. Check if the required secrets exist\n")
			sb.WriteString("2. Verify sealed secrets are properly decrypted\n")
			sb.WriteString("3. Check secret names and namespaces\n")
		}
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleGetEvents handles the get_events tool.
func (r *Registry) handleGetEvents(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate input to prevent injection attacks
	if err := validateToolArgs(args); err != nil {
		return nil, err
	}

	namespace, ok := args["namespace"].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	resourceName, _ := args["resource_name"].(string)
	eventType, _ := args["event_type"].(string)
	limitStr, _ := args["limit"].(string)

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	listOptions := metav1.ListOptions{}
	if resourceName != "" {
		// Sanitize resource name to prevent injection in field selector
		sanitizedResourceName := sanitizeForLogging(resourceName)
		listOptions.FieldSelector = fmt.Sprintf("involvedObject.name=%s", sanitizedResourceName)
	}

	events, err := r.clients.Clientset.CoreV1().Events(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Filter by event type
	var filtered []corev1.Event
	for _, event := range events.Items {
		if eventType == "all" || eventType == "" || event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	// Sort by last timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].LastTimestamp.After(filtered[j].LastTimestamp.Time)
	})

	// Apply limit
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	var sb strings.Builder
	// Sanitize user input before including in output
	sanitizedNamespace := sanitizeForLogging(namespace)
	sb.WriteString(fmt.Sprintf("# Events in %s\n\n", sanitizedNamespace))

	if resourceName != "" {
		sanitizedResourceName := sanitizeForLogging(resourceName)
		sb.WriteString(fmt.Sprintf("**Resource:** %s\n\n", sanitizedResourceName))
	}

	if len(filtered) == 0 {
		sb.WriteString("No events found matching the criteria.\n")
	} else {
		sb.WriteString("| Type | Object | Reason | Age | Count | Message |\n")
		sb.WriteString("|------|--------|--------|-----|:-----:|--------|\n")

		for _, event := range filtered {
			typeIcon := "ℹ️"
			if event.Type == "Warning" {
				typeIcon = "⚠️"
			}

			object := fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)
			age := formatAge(event.LastTimestamp.Time)

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %d | %s |\n",
				typeIcon, truncateString(object, 30), event.Reason, age, event.Count,
				truncateString(event.Message, 50)))
		}

		sb.WriteString(fmt.Sprintf("\n**Showing:** %d events\n", len(filtered)))
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleGetPodLogs handles the get_pod_logs tool.
func (r *Registry) handleGetPodLogs(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Validate input to prevent injection attacks
	if err := validateToolArgs(args); err != nil {
		return nil, err
	}

	podName, ok := args["pod_name"].(string)
	if !ok || podName == "" {
		return nil, fmt.Errorf("pod_name is required")
	}

	namespace, ok := args["namespace"].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	container, _ := args["container"].(string)
	tailLinesStr, _ := args["tail_lines"].(string)

	tailLines := int64(100)
	if tailLinesStr != "" {
		if l, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil {
			tailLines = l
		}
	}

	// Get pod to find containers if needed
	pod, err := r.clients.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s/%s: %w", namespace, podName, err)
	}

	// If no container specified, use first one
	if container == "" && len(pod.Spec.Containers) > 0 {
		container = pod.Spec.Containers[0].Name
	}

	// Get logs
	logOptions := &corev1.PodLogOptions{
		Container: container,
		TailLines: &tailLines,
	}

	req := r.clients.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	var sb strings.Builder
	// Sanitize user input before including in output
	sanitizedNamespace := sanitizeForLogging(namespace)
	sanitizedPodName := sanitizeForLogging(podName)
	sanitizedContainer := sanitizeForLogging(container)
	
	sb.WriteString(fmt.Sprintf("# Pod Logs: %s/%s\n\n", sanitizedNamespace, sanitizedPodName))
	sb.WriteString(fmt.Sprintf("**Container:** %s\n", sanitizedContainer))
	sb.WriteString(fmt.Sprintf("**Tail Lines:** %d\n\n", tailLines))
	sb.WriteString("```\n")
	
	// Redact sensitive data from pod logs before returning
	logContent := buf.String()
	redactedLogs := redactSensitiveData(logContent)
	sb.WriteString(redactedLogs)
	sb.WriteString("```\n")

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// formatAge formats a time as a human-readable age string.
func formatAge(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	d := time.Since(t)

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
