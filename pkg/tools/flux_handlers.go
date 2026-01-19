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

// Flux GVRs (GroupVersionResource)
var (
	kustomizationGVR = schema.GroupVersionResource{
		Group:    "kustomize.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "kustomizations",
	}

	gitRepositoryGVR = schema.GroupVersionResource{
		Group:    "source.toolkit.fluxcd.io",
		Version:  "v1",
		Resource: "gitrepositories",
	}

	helmReleaseGVR = schema.GroupVersionResource{
		Group:    "helm.toolkit.fluxcd.io",
		Version:  "v2",
		Resource: "helmreleases",
	}
)

// handleGetGitOpsStatus handles the get_gitops_status tool.
func (r *Registry) handleGetGitOpsStatus(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespace, _ := args["namespace"].(string)

	var sb strings.Builder
	sb.WriteString("# GitOps Status Summary\n\n")

	// Get Kustomizations
	var ksList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		ksList, err = r.clients.Dynamic.Resource(kustomizationGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		ksList, err = r.clients.Dynamic.Resource(kustomizationGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Error fetching Kustomizations: %s\n\n", err))
	} else {
		ready, failed, suspended := countResourceStatus(ksList.Items)
		sb.WriteString("## Flux Kustomizations\n\n")
		sb.WriteString(fmt.Sprintf("- ✅ Ready: %d\n", ready))
		sb.WriteString(fmt.Sprintf("- ❌ Failed: %d\n", failed))
		sb.WriteString(fmt.Sprintf("- ⏸️ Suspended: %d\n", suspended))
		sb.WriteString(fmt.Sprintf("- 📊 Total: %d\n\n", len(ksList.Items)))

		// List failed ones
		if failed > 0 {
			sb.WriteString("### Failed Kustomizations\n\n")
			for _, ks := range ksList.Items {
				if !isResourceReady(&ks) && !isResourceSuspended(&ks) {
					ns := ks.GetNamespace()
					name := ks.GetName()
					message := getConditionMessage(&ks, "Ready")
					sb.WriteString(fmt.Sprintf("- **%s/%s**: %s\n", ns, name, truncateString(message, 100)))
				}
			}
			sb.WriteString("\n")
		}
	}

	// Get GitRepositories
	var grList *unstructured.UnstructuredList

	if namespace != "" {
		grList, err = r.clients.Dynamic.Resource(gitRepositoryGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		grList, err = r.clients.Dynamic.Resource(gitRepositoryGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Error fetching GitRepositories: %s\n\n", err))
	} else {
		ready, failed, suspended := countResourceStatus(grList.Items)
		sb.WriteString("## Flux GitRepositories\n\n")
		sb.WriteString(fmt.Sprintf("- ✅ Ready: %d\n", ready))
		sb.WriteString(fmt.Sprintf("- ❌ Failed: %d\n", failed))
		sb.WriteString(fmt.Sprintf("- ⏸️ Suspended: %d\n", suspended))
		sb.WriteString(fmt.Sprintf("- 📊 Total: %d\n\n", len(grList.Items)))

		// List failed ones
		if failed > 0 {
			sb.WriteString("### Failed GitRepositories\n\n")
			for _, gr := range grList.Items {
				if !isResourceReady(&gr) && !isResourceSuspended(&gr) {
					ns := gr.GetNamespace()
					name := gr.GetName()
					message := getConditionMessage(&gr, "Ready")
					sb.WriteString(fmt.Sprintf("- **%s/%s**: %s\n", ns, name, truncateString(message, 100)))
				}
			}
			sb.WriteString("\n")
		}
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleListKustomizations handles the list_kustomizations tool.
func (r *Registry) handleListKustomizations(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespace, _ := args["namespace"].(string)
	statusFilter, _ := args["status_filter"].(string)
	if statusFilter == "" {
		statusFilter = "all"
	}

	var ksList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		ksList, err = r.clients.Dynamic.Resource(kustomizationGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		ksList, err = r.clients.Dynamic.Resource(kustomizationGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list Kustomizations: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Flux Kustomizations\n\n")
	sb.WriteString("| Namespace | Name | Ready | Suspended | Last Applied | Message |\n")
	sb.WriteString("|-----------|------|:-----:|:---------:|--------------|--------|\n")

	count := 0
	for _, ks := range ksList.Items {
		ready := isResourceReady(&ks)
		suspended := isResourceSuspended(&ks)

		// Apply filter
		switch statusFilter {
		case "ready":
			if !ready || suspended {
				continue
			}
		case "failed":
			if ready || suspended {
				continue
			}
		case "suspended":
			if !suspended {
				continue
			}
		}

		readyStr := "❌"
		if ready {
			readyStr = "✅"
		}
		suspendedStr := ""
		if suspended {
			suspendedStr = "⏸️"
		}

		lastApplied := getLastAppliedTime(&ks)
		message := truncateString(getConditionMessage(&ks, "Ready"), 50)

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			ks.GetNamespace(), ks.GetName(), readyStr, suspendedStr, lastApplied, message))
		count++
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d Kustomizations\n", count))

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleGetKustomization handles the get_kustomization tool.
func (r *Registry) handleGetKustomization(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}

	namespace, ok := args["namespace"].(string)
	if !ok || namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}

	ks, err := r.clients.Dynamic.Resource(kustomizationGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Kustomization %s/%s: %w", namespace, name, err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Kustomization: %s/%s\n\n", namespace, name))

	// Status
	sb.WriteString("## Status\n\n")
	ready := isResourceReady(ks)
	suspended := isResourceSuspended(ks)

	if suspended {
		sb.WriteString("**Status:** ⏸️ Suspended\n\n")
	} else if ready {
		sb.WriteString("**Status:** ✅ Ready\n\n")
	} else {
		sb.WriteString("**Status:** ❌ Not Ready\n\n")
	}

	// Spec details
	spec, _, _ := unstructured.NestedMap(ks.Object, "spec")
	if spec != nil {
		sb.WriteString("## Spec\n\n")

		if path, ok := spec["path"].(string); ok {
			sb.WriteString(fmt.Sprintf("- **Path:** `%s`\n", path))
		}

		if interval, ok := spec["interval"].(string); ok {
			sb.WriteString(fmt.Sprintf("- **Interval:** %s\n", interval))
		}

		if sourceRef, ok := spec["sourceRef"].(map[string]interface{}); ok {
			kind, _ := sourceRef["kind"].(string)
			name, _ := sourceRef["name"].(string)
			sb.WriteString(fmt.Sprintf("- **Source:** %s/%s\n", kind, name))
		}

		if prune, ok := spec["prune"].(bool); ok {
			sb.WriteString(fmt.Sprintf("- **Prune:** %v\n", prune))
		}

		sb.WriteString("\n")
	}

	// Dependencies
	if deps, found, _ := unstructured.NestedSlice(ks.Object, "spec", "dependsOn"); found && len(deps) > 0 {
		sb.WriteString("## Dependencies\n\n")
		for _, dep := range deps {
			if depMap, ok := dep.(map[string]interface{}); ok {
				depName, _ := depMap["name"].(string)
				depNs, _ := depMap["namespace"].(string)
				if depNs == "" {
					depNs = namespace
				}
				sb.WriteString(fmt.Sprintf("- %s/%s\n", depNs, depName))
			}
		}
		sb.WriteString("\n")
	}

	// Conditions
	sb.WriteString("## Conditions\n\n")
	conditions, _, _ := unstructured.NestedSlice(ks.Object, "status", "conditions")
	if len(conditions) > 0 {
		sb.WriteString("| Type | Status | Reason | Message |\n")
		sb.WriteString("|------|:------:|--------|--------|\n")
		for _, c := range conditions {
			if cond, ok := c.(map[string]interface{}); ok {
				condType, _ := cond["type"].(string)
				status, _ := cond["status"].(string)
				reason, _ := cond["reason"].(string)
				message, _ := cond["message"].(string)
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					condType, status, reason, truncateString(message, 60)))
			}
		}
	} else {
		sb.WriteString("No conditions available.\n")
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleListGitRepositories handles the list_gitrepositories tool.
func (r *Registry) handleListGitRepositories(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespace, _ := args["namespace"].(string)

	var grList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		grList, err = r.clients.Dynamic.Resource(gitRepositoryGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		grList, err = r.clients.Dynamic.Resource(gitRepositoryGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list GitRepositories: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Flux GitRepositories\n\n")
	sb.WriteString("| Namespace | Name | URL | Branch | Ready | Last Fetched |\n")
	sb.WriteString("|-----------|------|-----|--------|:-----:|--------------|\n")

	for _, gr := range grList.Items {
		ready := isResourceReady(&gr)
		readyStr := "❌"
		if ready {
			readyStr = "✅"
		}

		url, _, _ := unstructured.NestedString(gr.Object, "spec", "url")
		branch, _, _ := unstructured.NestedString(gr.Object, "spec", "ref", "branch")
		if branch == "" {
			branch, _, _ = unstructured.NestedString(gr.Object, "spec", "ref", "tag")
		}
		lastFetched := getLastAppliedTime(&gr)

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			gr.GetNamespace(), gr.GetName(), truncateString(url, 40), branch, readyStr, lastFetched))
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d GitRepositories\n", len(grList.Items)))

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleGetHelmReleases handles the get_helmreleases tool.
func (r *Registry) handleGetHelmReleases(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	namespace, _ := args["namespace"].(string)
	statusFilter, _ := args["status_filter"].(string)
	if statusFilter == "" {
		statusFilter = "all"
	}

	var hrList *unstructured.UnstructuredList
	var err error

	if namespace != "" {
		hrList, err = r.clients.Dynamic.Resource(helmReleaseGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	} else {
		hrList, err = r.clients.Dynamic.Resource(helmReleaseGVR).List(ctx, metav1.ListOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list HelmReleases: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Flux HelmReleases\n\n")
	sb.WriteString("| Namespace | Name | Chart | Version | Ready | Message |\n")
	sb.WriteString("|-----------|------|-------|---------|:-----:|--------|\n")

	count := 0
	for _, hr := range hrList.Items {
		ready := isResourceReady(&hr)
		suspended := isResourceSuspended(&hr)

		// Apply filter
		switch statusFilter {
		case "ready":
			if !ready || suspended {
				continue
			}
		case "failed":
			if ready || suspended {
				continue
			}
		case "suspended":
			if !suspended {
				continue
			}
		}

		readyStr := "❌"
		if ready {
			readyStr = "✅"
		}
		if suspended {
			readyStr = "⏸️"
		}

		chart, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "chart")
		version, _, _ := unstructured.NestedString(hr.Object, "spec", "chart", "spec", "version")
		message := truncateString(getConditionMessage(&hr, "Ready"), 40)

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			hr.GetNamespace(), hr.GetName(), chart, version, readyStr, message))
		count++
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d HelmReleases\n", count))

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// Helper functions

func countResourceStatus(items []unstructured.Unstructured) (ready, failed, suspended int) {
	for _, item := range items {
		if isResourceSuspended(&item) {
			suspended++
		} else if isResourceReady(&item) {
			ready++
		} else {
			failed++
		}
	}
	return
}

func isResourceReady(obj *unstructured.Unstructured) bool {
	conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found {
		return false
	}

	for _, c := range conditions {
		if cond, ok := c.(map[string]interface{}); ok {
			if condType, _ := cond["type"].(string); condType == "Ready" {
				status, _ := cond["status"].(string)
				return status == "True"
			}
		}
	}
	return false
}

func isResourceSuspended(obj *unstructured.Unstructured) bool {
	suspended, found, _ := unstructured.NestedBool(obj.Object, "spec", "suspend")
	return found && suspended
}

func getConditionMessage(obj *unstructured.Unstructured, conditionType string) string {
	conditions, found, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found {
		return ""
	}

	for _, c := range conditions {
		if cond, ok := c.(map[string]interface{}); ok {
			if t, _ := cond["type"].(string); t == conditionType {
				message, _ := cond["message"].(string)
				return message
			}
		}
	}
	return ""
}

func getLastAppliedTime(obj *unstructured.Unstructured) string {
	// Try lastAppliedRevision first
	lastApplied, found, _ := unstructured.NestedString(obj.Object, "status", "lastAppliedRevision")
	if found && lastApplied != "" {
		// Extract commit SHA if present
		parts := strings.Split(lastApplied, "/")
		if len(parts) > 1 {
			return parts[len(parts)-1][:7] // Short SHA
		}
		return lastApplied
	}

	// Try lastHandledReconcileAt
	lastReconcile, found, _ := unstructured.NestedString(obj.Object, "status", "lastHandledReconcileAt")
	if found && lastReconcile != "" {
		return lastReconcile
	}

	return "-"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
