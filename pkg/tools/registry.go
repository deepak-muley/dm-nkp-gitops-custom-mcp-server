// Package tools provides the tool implementations for the MCP server.
package tools

import (
	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/config"
	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
)

// Logger interface for logging.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Registry manages tool registration and execution.
type Registry struct {
	clients  *config.K8sClients
	readOnly bool
	logger   Logger
	tools    []mcp.Tool
	handlers map[string]mcp.ToolHandler
}

// NewRegistry creates a new tool registry.
func NewRegistry(clients *config.K8sClients, readOnly bool, logger Logger) *Registry {
	return &Registry{
		clients:  clients,
		readOnly: readOnly,
		logger:   logger,
		tools:    []mcp.Tool{},
		handlers: make(map[string]mcp.ToolHandler),
	}
}

// RegisterAllTools registers all available tools.
func (r *Registry) RegisterAllTools() {
	// Context tools
	r.registerContextTools()

	// Flux/GitOps tools
	r.registerFluxTools()

	// Cluster tools
	r.registerClusterTools()

	// App deployment tools
	r.registerAppTools()

	// Debugging tools
	r.registerDebugTools()

	// Policy tools
	r.registerPolicyTools()

	r.logger.Info("Registered tools", "count", len(r.tools))
}

// GetTools returns all registered tools.
func (r *Registry) GetTools() []mcp.Tool {
	return r.tools
}

// GetHandlers returns all registered handlers.
func (r *Registry) GetHandlers() map[string]mcp.ToolHandler {
	return r.handlers
}

// register adds a tool and its handler to the registry.
func (r *Registry) register(tool mcp.Tool, handler mcp.ToolHandler) {
	r.tools = append(r.tools, tool)
	r.handlers[tool.Name] = handler
}

// registerContextTools registers context-related tools.
func (r *Registry) registerContextTools() {
	// list_contexts
	r.register(
		mcp.Tool{
			Name:        "list_contexts",
			Description: "List all available Kubernetes contexts from the kubeconfig",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		r.handleListContexts,
	)

	// get_current_context
	r.register(
		mcp.Tool{
			Name:        "get_current_context",
			Description: "Get the current active Kubernetes context",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
			},
		},
		r.handleGetCurrentContext,
	)
}

// registerFluxTools registers Flux/GitOps tools.
func (r *Registry) registerFluxTools() {
	// get_gitops_status
	r.register(
		mcp.Tool{
			Name:        "get_gitops_status",
			Description: "Get overall GitOps status including all Flux Kustomizations and GitRepositories. Returns summary of healthy/unhealthy/suspended resources.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces). Common values: dm-nkp-gitops-infra, kommander",
					},
				},
			},
		},
		r.handleGetGitOpsStatus,
	)

	// list_kustomizations
	r.register(
		mcp.Tool{
			Name:        "list_kustomizations",
			Description: "List all Flux Kustomizations with their reconciliation status",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
					"status_filter": {
						Type:        "string",
						Description: "Filter by status: all, ready, failed, suspended",
						Enum:        []string{"all", "ready", "failed", "suspended"},
						Default:     "all",
					},
				},
			},
		},
		r.handleListKustomizations,
	)

	// get_kustomization
	r.register(
		mcp.Tool{
			Name:        "get_kustomization",
			Description: "Get detailed information about a specific Flux Kustomization including conditions, source, and dependencies",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"name": {
						Type:        "string",
						Description: "Name of the Kustomization",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace of the Kustomization",
					},
				},
				Required: []string{"name", "namespace"},
			},
		},
		r.handleGetKustomization,
	)

	// list_gitrepositories
	r.register(
		mcp.Tool{
			Name:        "list_gitrepositories",
			Description: "List all Flux GitRepository sources with their sync status",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
				},
			},
		},
		r.handleListGitRepositories,
	)
}

// registerClusterTools registers CAPI cluster tools.
func (r *Registry) registerClusterTools() {
	// get_cluster_status
	r.register(
		mcp.Tool{
			Name:        "get_cluster_status",
			Description: "Get status of CAPI (Cluster API) clusters. Shows phase, conditions, and infrastructure status.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"cluster_name": {
						Type:        "string",
						Description: "Name of the cluster (e.g., dm-nkp-workload-1). Leave empty for all clusters.",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
				},
			},
		},
		r.handleGetClusterStatus,
	)

	// list_machines
	r.register(
		mcp.Tool{
			Name:        "list_machines",
			Description: "List CAPI Machines for a cluster showing node status and provider info",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"cluster_name": {
						Type:        "string",
						Description: "Name of the cluster to filter machines",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
				},
			},
		},
		r.handleListMachines,
	)
}

// registerAppTools registers application deployment tools.
func (r *Registry) registerAppTools() {
	// get_app_deployments
	r.register(
		mcp.Tool{
			Name:        "get_app_deployments",
			Description: "Get application deployment status across workspaces. Shows App and ClusterApp resources from Kommander.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"workspace": {
						Type:        "string",
						Description: "Workspace name (e.g., dm-dev-workspace). Leave empty for all workspaces.",
					},
					"app_name": {
						Type:        "string",
						Description: "Application name to filter. Leave empty for all apps.",
					},
				},
			},
		},
		r.handleGetAppDeployments,
	)

	// get_helmreleases
	r.register(
		mcp.Tool{
			Name:        "get_helmreleases",
			Description: "List Flux HelmReleases with their status",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
					"status_filter": {
						Type:        "string",
						Description: "Filter by status: all, ready, failed, suspended",
						Enum:        []string{"all", "ready", "failed", "suspended"},
						Default:     "all",
					},
				},
			},
		},
		r.handleGetHelmReleases,
	)
}

// registerDebugTools registers debugging tools.
func (r *Registry) registerDebugTools() {
	// debug_reconciliation
	r.register(
		mcp.Tool{
			Name:        "debug_reconciliation",
			Description: "Debug a failing Flux reconciliation. Shows conditions, events, and related resource status.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"resource_type": {
						Type:        "string",
						Description: "Type of resource: kustomization, gitrepository, helmrelease",
						Enum:        []string{"kustomization", "gitrepository", "helmrelease"},
					},
					"name": {
						Type:        "string",
						Description: "Name of the resource",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace of the resource",
					},
				},
				Required: []string{"resource_type", "name", "namespace"},
			},
		},
		r.handleDebugReconciliation,
	)

	// get_events
	r.register(
		mcp.Tool{
			Name:        "get_events",
			Description: "Get Kubernetes events for debugging. Can filter by namespace, resource, or event type.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"namespace": {
						Type:        "string",
						Description: "Namespace to get events from (required)",
					},
					"resource_name": {
						Type:        "string",
						Description: "Filter events for a specific resource name",
					},
					"event_type": {
						Type:        "string",
						Description: "Filter by event type: all, Normal, Warning",
						Enum:        []string{"all", "Normal", "Warning"},
						Default:     "all",
					},
					"limit": {
						Type:        "string",
						Description: "Maximum number of events to return (default: 20)",
						Default:     "20",
					},
				},
				Required: []string{"namespace"},
			},
		},
		r.handleGetEvents,
	)

	// get_pod_logs
	r.register(
		mcp.Tool{
			Name:        "get_pod_logs",
			Description: "Get logs from a pod for debugging",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"pod_name": {
						Type:        "string",
						Description: "Name of the pod",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace of the pod",
					},
					"container": {
						Type:        "string",
						Description: "Container name (optional, uses first container if not specified)",
					},
					"tail_lines": {
						Type:        "string",
						Description: "Number of lines to return from end (default: 100)",
						Default:     "100",
					},
				},
				Required: []string{"pod_name", "namespace"},
			},
		},
		r.handleGetPodLogs,
	)
}

// registerPolicyTools registers policy-related tools.
func (r *Registry) registerPolicyTools() {
	// check_policy_violations
	r.register(
		mcp.Tool{
			Name:        "check_policy_violations",
			Description: "Check for Gatekeeper or Kyverno policy violations across the cluster",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"policy_engine": {
						Type:        "string",
						Description: "Policy engine to check: gatekeeper, kyverno, or both",
						Enum:        []string{"gatekeeper", "kyverno", "both"},
						Default:     "both",
					},
					"namespace": {
						Type:        "string",
						Description: "Namespace to filter (default: all namespaces)",
					},
				},
			},
		},
		r.handleCheckPolicyViolations,
	)

	// list_constraints
	r.register(
		mcp.Tool{
			Name:        "list_constraints",
			Description: "List Gatekeeper constraints and their enforcement status",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"constraint_kind": {
						Type:        "string",
						Description: "Filter by constraint kind (e.g., K8sRequiredLabels)",
					},
				},
			},
		},
		r.handleListConstraints,
	)
}
