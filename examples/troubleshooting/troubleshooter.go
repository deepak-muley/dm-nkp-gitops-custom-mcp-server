package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// TroubleshootWorkflow represents a step-by-step troubleshooting procedure
type TroubleshootWorkflow struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []Step   `json:"steps"`
	DecisionTree DecisionTree `json:"decision_tree,omitempty"`
}

// Step represents a single troubleshooting step
type Step struct {
	Number      int                    `json:"number"`
	Description string                 `json:"description"`
	Tool        string                 `json:"tool"`
	Arguments   map[string]interface{} `json:"arguments"`
	Expected    string                 `json:"expected,omitempty"`
	NextStep    map[string]int         `json:"next_step,omitempty"` // Maps condition to next step number
}

// DecisionTree represents a decision-making structure
type DecisionTree struct {
	Root       DecisionNode `json:"root"`
}

// DecisionNode represents a node in the decision tree
type DecisionNode struct {
	Condition string         `json:"condition"`
	IfTrue    *DecisionNode  `json:"if_true,omitempty"`
	IfFalse   *DecisionNode  `json:"if_false,omitempty"`
	Action    string         `json:"action,omitempty"`
}

// Example 1: GitOps Reconciliation Failure Workflow
func getGitOpsFailureWorkflow() TroubleshootWorkflow {
	return TroubleshootWorkflow{
		Name:        "GitOps Reconciliation Failure",
		Description: "Systematic approach to diagnosing and fixing GitOps reconciliation failures",
		Steps: []Step{
			{
				Number:      1,
				Description: "Get overall GitOps status to identify failing resources",
				Tool:        "get_gitops_status",
				Arguments: map[string]interface{}{
					"namespace": "${namespace}",
				},
				Expected: "Shows count of ready/failed/suspended resources",
				NextStep: map[string]int{
					"has_failures": 2,
					"all_ready":    0, // End
				},
			},
			{
				Number:      2,
				Description: "List all failing Kustomizations",
				Tool:        "list_kustomizations",
				Arguments: map[string]interface{}{
					"namespace":    "${namespace}",
					"status_filter": "failed",
				},
				Expected: "List of Kustomizations with Ready=False",
				NextStep: map[string]int{
					"found": 3,
				},
			},
			{
				Number:      3,
				Description: "Debug the specific failing Kustomization",
				Tool:        "debug_reconciliation",
				Arguments: map[string]interface{}{
					"resource_type": "kustomization",
					"name":          "${kustomization_name}",
					"namespace":     "${namespace}",
				},
				Expected: "Detailed status including conditions and error messages",
				NextStep: map[string]int{
					"source_error":     4,
					"dependency_error": 5,
					"validation_error": 6,
					"secret_error":     7,
				},
			},
			{
				Number:      4,
				Description: "Source issue - Check GitRepository status",
				Tool:        "list_gitrepositories",
				Arguments: map[string]interface{}{
					"namespace": "${namespace}",
				},
				Expected: "List of GitRepositories with sync status",
			},
			{
				Number:      5,
				Description: "Dependency issue - Recursively debug dependencies",
				Tool:        "debug_reconciliation",
				Arguments: map[string]interface{}{
					"resource_type": "kustomization",
					"name":          "${dependency_name}",
					"namespace":     "${namespace}",
				},
				Expected: "Status of dependent resource",
			},
			{
				Number:      6,
				Description: "Validation error - Check events for validation details",
				Tool:        "get_events",
				Arguments: map[string]interface{}{
					"namespace":   "${namespace}",
					"event_type":  "Warning",
					"limit":       "50",
				},
				Expected: "Recent warning events with validation errors",
			},
			{
				Number:      7,
				Description: "Secret error - Check for missing secrets",
				Tool:        "get_events",
				Arguments: map[string]interface{}{
					"namespace":  "${namespace}",
					"event_type": "Warning",
				},
				Expected: "Events mentioning secret issues",
			},
		},
		DecisionTree: DecisionTree{
			Root: DecisionNode{
				Condition: "Kustomization Ready status",
				IfTrue: &DecisionNode{
					Condition: "Error message contains 'Source'",
					IfTrue: &DecisionNode{
						Action: "Check GitRepository - Step 4",
					},
					IfFalse: &DecisionNode{
						Condition: "Error message contains 'Dependency'",
						IfTrue: &DecisionNode{
							Action: "Debug dependency - Step 5",
						},
						IfFalse: &DecisionNode{
							Condition: "Error message contains 'Validation'",
							IfTrue: &DecisionNode{
								Action: "Check validation events - Step 6",
							},
							IfFalse: &DecisionNode{
								Action: "Check secret events - Step 7",
							},
						},
					},
				},
				IfFalse: &DecisionNode{
					Action: "No action needed - resource is healthy",
				},
			},
		},
	}
}

// Example 2: Cluster Node Health Workflow
func getClusterNodeWorkflow() TroubleshootWorkflow {
	return TroubleshootWorkflow{
		Name:        "Cluster Node Health Issues",
		Description: "Diagnose and resolve cluster node provisioning and health problems",
		Steps: []Step{
			{
				Number:      1,
				Description: "Get cluster status",
				Tool:        "get_cluster_status",
				Arguments: map[string]interface{}{
					"cluster_name": "${cluster_name}",
					"namespace":    "${namespace}",
				},
				Expected: "Cluster phase and conditions",
				NextStep: map[string]int{
					"provisioning": 2,
					"running":      3,
					"failed":       4,
				},
			},
			{
				Number:      2,
				Description: "Cluster in Provisioning - Check infrastructure readiness",
				Tool:        "get_cluster_status",
				Arguments: map[string]interface{}{
					"cluster_name": "${cluster_name}",
					"namespace":    "${namespace}",
				},
				Expected: "InfrastructureReady condition details",
			},
			{
				Number:      3,
				Description: "Cluster Running - Check machine/node status",
				Tool:        "list_machines",
				Arguments: map[string]interface{}{
					"cluster_name": "${cluster_name}",
					"namespace":    "${namespace}",
				},
				Expected: "List of machines with their provisioning status",
			},
			{
				Number:      4,
				Description: "Cluster Failed - Check events for failure reason",
				Tool:        "get_events",
				Arguments: map[string]interface{}{
					"namespace":  "${namespace}",
					"event_type": "Warning",
					"limit":      "50",
				},
				Expected: "Warning events explaining failure",
			},
		},
	}
}

// Example 3: Application Deployment Failure Workflow
func getAppDeploymentWorkflow() TroubleshootWorkflow {
	return TroubleshootWorkflow{
		Name:        "Application Deployment Failure",
		Description: "Troubleshoot application deployment issues via Kommander/Helm",
		Steps: []Step{
			{
				Number:      1,
				Description: "Check application deployment status",
				Tool:        "get_app_deployments",
				Arguments: map[string]interface{}{
					"workspace": "${workspace}",
					"app_name":  "${app_name}",
				},
				Expected: "App status and cluster information",
				NextStep: map[string]int{
					"not_ready": 2,
				},
			},
			{
				Number:      2,
				Description: "Check HelmRelease status",
				Tool:        "get_helmreleases",
				Arguments: map[string]interface{}{
					"namespace":    "${namespace}",
					"status_filter": "failed",
				},
				Expected: "Failing HelmReleases",
				NextStep: map[string]int{
					"found": 3,
				},
			},
			{
				Number:      3,
				Description: "Debug the failing HelmRelease",
				Tool:        "debug_reconciliation",
				Arguments: map[string]interface{}{
					"resource_type": "helmrelease",
					"name":          "${helmrelease_name}",
					"namespace":     "${namespace}",
				},
				Expected: "Detailed HelmRelease status and error messages",
			},
			{
				Number:      4,
				Description: "Get pod logs if pods are crashing",
				Tool:        "get_pod_logs",
				Arguments: map[string]interface{}{
					"pod_name":    "${pod_name}",
					"namespace":   "${namespace}",
					"tail_lines":  "100",
				},
				Expected: "Recent pod logs showing errors",
			},
		},
	}
}

// ExecuteWorkflow simulates executing a troubleshooting workflow
func ExecuteWorkflow(workflow TroubleshootWorkflow, context map[string]string) {
	fmt.Printf("=== Starting Workflow: %s ===\n\n", workflow.Name)
	fmt.Printf("Description: %s\n\n", workflow.Description)

	for _, step := range workflow.Steps {
		fmt.Printf("Step %d: %s\n", step.Number, step.Description)
		fmt.Printf("  Tool: %s\n", step.Tool)
		
		// Resolve template variables
		args := resolveArguments(step.Arguments, context)
		argsJSON, _ := json.MarshalIndent(args, "  ", "  ")
		fmt.Printf("  Arguments:\n%s\n", string(argsJSON))
		
		if step.Expected != "" {
			fmt.Printf("  Expected: %s\n", step.Expected)
		}
		
		fmt.Println()
	}
}

// resolveArguments replaces template variables with actual values
func resolveArguments(args map[string]interface{}, context map[string]string) map[string]interface{} {
	resolved := make(map[string]interface{})
	for k, v := range args {
		switch val := v.(type) {
		case string:
			if val != "" && val[0] == '$' {
				// Template variable - replace with context value
				key := val[2 : len(val)-1] // Remove "${" and "}"
				if ctxVal, ok := context[key]; ok {
					resolved[k] = ctxVal
				} else {
					resolved[k] = val // Keep original if not found
				}
			} else {
				resolved[k] = val
			}
		default:
			resolved[k] = val
		}
	}
	return resolved
}

// GenerateMCPRequest generates an MCP JSON-RPC request for a step
func GenerateMCPRequest(step Step, context map[string]string) (string, error) {
	args := resolveArguments(step.Arguments, context)
	
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      step.Number,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      step.Tool,
			"arguments": args,
		},
	}
	
	jsonData, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(jsonData), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run troubleshooter.go <workflow_name>")
		fmt.Println("\nAvailable workflows:")
		fmt.Println("  - gitops-failure")
		fmt.Println("  - cluster-node")
		fmt.Println("  - app-deployment")
		fmt.Println("  - all (prints all workflows as JSON)")
		os.Exit(1)
	}

	workflowName := os.Args[1]
	
	var workflows map[string]TroubleshootWorkflow
	
	switch workflowName {
	case "gitops-failure":
		workflows = map[string]TroubleshootWorkflow{
			"gitops-failure": getGitOpsFailureWorkflow(),
		}
	case "cluster-node":
		workflows = map[string]TroubleshootWorkflow{
			"cluster-node": getClusterNodeWorkflow(),
		}
	case "app-deployment":
		workflows = map[string]TroubleshootWorkflow{
			"app-deployment": getAppDeploymentWorkflow(),
		}
	case "all":
		workflows = map[string]TroubleshootWorkflow{
			"gitops-failure":  getGitOpsFailureWorkflow(),
			"cluster-node":    getClusterNodeWorkflow(),
			"app-deployment":  getAppDeploymentWorkflow(),
		}
	default:
		log.Fatalf("Unknown workflow: %s", workflowName)
	}

	if workflowName == "all" {
		// Print as JSON
		jsonData, err := json.MarshalIndent(workflows, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Execute workflow with example context
		workflow := workflows[workflowName]
		
		// Example context - replace with actual values
		context := map[string]string{
			"namespace":         "flux-system",
			"kustomization_name": "infrastructure",
			"cluster_name":      "dm-nkp-workload-1",
			"workspace":         "dm-dev-workspace",
			"app_name":          "traefik",
			"helmrelease_name":  "traefik-helmrelease",
			"pod_name":          "traefik-xxx-xxx",
			"dependency_name":   "base-cluster-resources",
		}
		
		ExecuteWorkflow(workflow, context)
		
		// Generate example MCP request for first step
		if len(workflow.Steps) > 0 {
			fmt.Println("=== Example MCP Request (First Step) ===")
			request, err := GenerateMCPRequest(workflow.Steps[0], context)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(request)
		}
	}
}
