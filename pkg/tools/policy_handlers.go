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

// Gatekeeper GVRs
var (
	constraintTemplateGVR = schema.GroupVersionResource{
		Group:    "templates.gatekeeper.sh",
		Version:  "v1",
		Resource: "constrainttemplates",
	}

	// Kyverno GVRs
	kyvernoPolicyGVR = schema.GroupVersionResource{
		Group:    "kyverno.io",
		Version:  "v1",
		Resource: "clusterpolicies",
	}

	kyvernoPolicyReportGVR = schema.GroupVersionResource{
		Group:    "wgpolicyk8s.io",
		Version:  "v1alpha2",
		Resource: "clusterpolicyreports",
	}
)

// handleCheckPolicyViolations handles the check_policy_violations tool.
func (r *Registry) handleCheckPolicyViolations(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	policyEngine, _ := args["policy_engine"].(string)
	if policyEngine == "" {
		policyEngine = "both"
	}

	var sb strings.Builder
	sb.WriteString("# Policy Violations Report\n\n")

	// Check Gatekeeper
	if policyEngine == "gatekeeper" || policyEngine == "both" {
		sb.WriteString("## Gatekeeper Violations\n\n")

		// Get all constraint templates to find constraint kinds
		ctList, err := r.clients.Dynamic.Resource(constraintTemplateGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			sb.WriteString(fmt.Sprintf("⚠️ Gatekeeper not installed or error: %s\n\n", err))
		} else {
			totalViolations := 0
			constraintKinds := []string{}

			for _, ct := range ctList.Items {
				kind, _, _ := unstructured.NestedString(ct.Object, "spec", "crd", "spec", "names", "kind")
				if kind != "" {
					constraintKinds = append(constraintKinds, kind)
				}
			}

			if len(constraintKinds) == 0 {
				sb.WriteString("No constraint templates found.\n\n")
			} else {
				sb.WriteString("| Constraint | Kind | Violations | Enforcement |\n")
				sb.WriteString("|------------|------|:----------:|:-----------:|\n")

				for _, kind := range constraintKinds {
					// Get constraints of this kind
					constraintGVR := schema.GroupVersionResource{
						Group:    "constraints.gatekeeper.sh",
						Version:  "v1beta1",
						Resource: strings.ToLower(kind),
					}

					constraints, err := r.clients.Dynamic.Resource(constraintGVR).List(ctx, metav1.ListOptions{})
					if err != nil {
						continue
					}

					for _, constraint := range constraints.Items {
						name := constraint.GetName()
						violations, _, _ := unstructured.NestedInt64(constraint.Object, "status", "totalViolations")
						enforcement, _, _ := unstructured.NestedString(constraint.Object, "spec", "enforcementAction")
						if enforcement == "" {
							enforcement = "deny"
						}

						violationIcon := "✅"
						if violations > 0 {
							violationIcon = "❌"
							totalViolations += int(violations)
						}

						sb.WriteString(fmt.Sprintf("| %s | %s | %s %d | %s |\n",
							name, kind, violationIcon, violations, enforcement))
					}
				}

				sb.WriteString(fmt.Sprintf("\n**Total Gatekeeper Violations:** %d\n\n", totalViolations))
			}
		}
	}

	// Check Kyverno
	if policyEngine == "kyverno" || policyEngine == "both" {
		sb.WriteString("## Kyverno Policy Status\n\n")

		// Get ClusterPolicies
		policies, err := r.clients.Dynamic.Resource(kyvernoPolicyGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			sb.WriteString(fmt.Sprintf("⚠️ Kyverno not installed or error: %s\n\n", err))
		} else if len(policies.Items) == 0 {
			sb.WriteString("No Kyverno policies found.\n\n")
		} else {
			sb.WriteString("| Policy | Ready | Background | Validation Mode |\n")
			sb.WriteString("|--------|:-----:|:----------:|:---------------:|\n")

			for _, policy := range policies.Items {
				name := policy.GetName()
				ready := isResourceReady(&policy)
				readyIcon := "❌"
				if ready {
					readyIcon = "✅"
				}

				background, _, _ := unstructured.NestedBool(policy.Object, "spec", "background")
				backgroundStr := "No"
				if background {
					backgroundStr = "Yes"
				}

				validationMode, _, _ := unstructured.NestedString(policy.Object, "spec", "validationFailureAction")
				if validationMode == "" {
					validationMode = "Audit"
				}

				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					name, readyIcon, backgroundStr, validationMode))
			}
			sb.WriteString("\n")
		}

		// Get Policy Reports
		sb.WriteString("### Policy Reports\n\n")
		reports, err := r.clients.Dynamic.Resource(kyvernoPolicyReportGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			sb.WriteString(fmt.Sprintf("⚠️ Could not fetch policy reports: %s\n\n", err))
		} else if len(reports.Items) == 0 {
			sb.WriteString("No policy reports found.\n\n")
		} else {
			totalPass := 0
			totalFail := 0
			totalWarn := 0

			for _, report := range reports.Items {
				summary, found, _ := unstructured.NestedMap(report.Object, "summary")
				if found {
					if pass, ok := summary["pass"].(int64); ok {
						totalPass += int(pass)
					}
					if fail, ok := summary["fail"].(int64); ok {
						totalFail += int(fail)
					}
					if warn, ok := summary["warn"].(int64); ok {
						totalWarn += int(warn)
					}
				}
			}

			sb.WriteString(fmt.Sprintf("- ✅ Pass: %d\n", totalPass))
			sb.WriteString(fmt.Sprintf("- ❌ Fail: %d\n", totalFail))
			sb.WriteString(fmt.Sprintf("- ⚠️ Warn: %d\n\n", totalWarn))
		}
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// handleListConstraints handles the list_constraints tool.
func (r *Registry) handleListConstraints(args map[string]interface{}) (*mcp.ToolCallResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	constraintKind, _ := args["constraint_kind"].(string)

	var sb strings.Builder
	sb.WriteString("# Gatekeeper Constraints\n\n")

	// Get constraint templates
	ctList, err := r.clients.Dynamic.Resource(constraintTemplateGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list constraint templates: %w", err)
	}

	if len(ctList.Items) == 0 {
		sb.WriteString("No constraint templates found. Gatekeeper may not be installed.\n")
		return &mcp.ToolCallResult{
			Content: []mcp.Content{
				{Type: "text", Text: sb.String()},
			},
		}, nil
	}

	// First list constraint templates
	sb.WriteString("## Constraint Templates\n\n")
	sb.WriteString("| Name | Kind | Description |\n")
	sb.WriteString("|------|------|-------------|\n")

	constraintKinds := make(map[string]string) // kind -> template name

	for _, ct := range ctList.Items {
		name := ct.GetName()
		kind, _, _ := unstructured.NestedString(ct.Object, "spec", "crd", "spec", "names", "kind")

		// Get description from annotation or targets
		description := ""
		if annotations := ct.GetAnnotations(); annotations != nil {
			description = annotations["description"]
		}
		if description == "" {
			description = truncateString(name, 50)
		}

		if kind != "" {
			constraintKinds[kind] = name
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", name, kind, truncateString(description, 40)))
	}

	sb.WriteString(fmt.Sprintf("\n**Total Templates:** %d\n\n", len(ctList.Items)))

	// List constraints
	sb.WriteString("## Active Constraints\n\n")

	if constraintKind != "" {
		// Filter to specific kind
		kindsToCheck := []string{constraintKind}
		sb.WriteString(fmt.Sprintf("**Filtering by kind:** %s\n\n", constraintKind))
		r.listConstraintsForKinds(ctx, &sb, kindsToCheck)
	} else {
		// List all
		var kinds []string
		for kind := range constraintKinds {
			kinds = append(kinds, kind)
		}
		r.listConstraintsForKinds(ctx, &sb, kinds)
	}

	return &mcp.ToolCallResult{
		Content: []mcp.Content{
			{Type: "text", Text: sb.String()},
		},
	}, nil
}

// listConstraintsForKinds lists constraints for the given kinds.
func (r *Registry) listConstraintsForKinds(ctx context.Context, sb *strings.Builder, kinds []string) {
	sb.WriteString("| Kind | Name | Enforcement | Violations | Matches |\n")
	sb.WriteString("|------|------|:-----------:|:----------:|--------|\n")

	totalConstraints := 0

	for _, kind := range kinds {
		constraintGVR := schema.GroupVersionResource{
			Group:    "constraints.gatekeeper.sh",
			Version:  "v1beta1",
			Resource: strings.ToLower(kind),
		}

		constraints, err := r.clients.Dynamic.Resource(constraintGVR).List(ctx, metav1.ListOptions{})
		if err != nil {
			continue
		}

		for _, constraint := range constraints.Items {
			name := constraint.GetName()
			violations, _, _ := unstructured.NestedInt64(constraint.Object, "status", "totalViolations")
			enforcement, _, _ := unstructured.NestedString(constraint.Object, "spec", "enforcementAction")
			if enforcement == "" {
				enforcement = "deny"
			}

			// Get match info
			matchKinds := []string{}
			if match, found, _ := unstructured.NestedSlice(constraint.Object, "spec", "match", "kinds"); found {
				for _, m := range match {
					if mMap, ok := m.(map[string]interface{}); ok {
						if kinds, ok := mMap["kinds"].([]interface{}); ok {
							for _, k := range kinds {
								if kStr, ok := k.(string); ok {
									matchKinds = append(matchKinds, kStr)
								}
							}
						}
					}
				}
			}
			matchStr := strings.Join(matchKinds, ", ")
			if matchStr == "" {
				matchStr = "*"
			}

			violationIcon := ""
			if violations > 0 {
				violationIcon = "❌ "
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s%d | %s |\n",
				kind, name, enforcement, violationIcon, violations, truncateString(matchStr, 30)))
			totalConstraints++
		}
	}

	sb.WriteString(fmt.Sprintf("\n**Total Constraints:** %d\n", totalConstraints))
}
