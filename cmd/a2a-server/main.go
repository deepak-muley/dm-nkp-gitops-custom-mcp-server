// Package main provides the entry point for the A2A (Agent-to-Agent) server.
//
// This server exposes your existing MCP tools as A2A skills, enabling
// agent-to-agent communication over HTTP.
//
// Key Differences from MCP Server:
//   - Transport: HTTP instead of stdio
//   - Discovery: Agent Card at /.well-known/agent.json
//   - Execution: Task-based (stateful) instead of direct tool calls
//   - Communication: Between agents, not just human/AI → tool
//
// Usage:
//
//	dm-nkp-gitops-a2a-server serve --port 8080
//
// Then access:
//   - Agent Card: http://localhost:8080/.well-known/agent.json
//   - Health: http://localhost:8080/health
//   - JSON-RPC: POST http://localhost:8080/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/a2a"
	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/config"
	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/tools"
)

// Version information (set via ldflags)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		runServer()
	case "version":
		printVersion()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runServer() {
	// Parse flags
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	kubeconfig := fs.String("kubeconfig", "", "Path to kubeconfig file")
	kubeContext := fs.String("context", "", "Kubernetes context to use")
	port := fs.Int("port", 8080, "Port to listen on")
	baseURL := fs.String("base-url", "", "Base URL for agent card (auto-generated if empty)")
	readOnly := fs.Bool("read-only", true, "Enable read-only mode")
	logLevel := fs.String("log-level", "info", "Log level: debug, info, warn, error")
	fs.Parse(os.Args[2:])

	// Setup logging
	logger := config.NewLogger(*logLevel)

	logger.Info("Starting dm-nkp-gitops A2A server",
		"version", Version,
		"commit", GitCommit,
		"port", *port,
		"readOnly", *readOnly,
	)

	// Load Kubernetes configuration
	k8sConfig, err := config.LoadKubeConfig(*kubeconfig, *kubeContext)
	if err != nil {
		logger.Error("Failed to load kubeconfig", "error", err)
		os.Exit(1)
	}

	// Create Kubernetes clients
	clients, err := config.NewK8sClients(k8sConfig)
	if err != nil {
		logger.Error("Failed to create Kubernetes clients", "error", err)
		os.Exit(1)
	}

	// Register tools (same as MCP server)
	registry := tools.NewRegistry(clients, *readOnly, logger)
	registry.RegisterAllTools()

	// Create A2A server with MCP tools
	server := a2a.NewServer(a2a.ServerConfig{
		Name:        "dm-nkp-gitops-agent",
		Version:     Version,
		Description: "A2A agent for NKP GitOps infrastructure monitoring and debugging. Exposes GitOps, Cluster, App, and Policy tools as A2A skills.",
		Port:        *port,
		BaseURL:     *baseURL,
		Tools:       registry.GetTools(),
		Handlers:    registry.GetHandlers(),
		Logger:      logger,
		ReadOnly:    *readOnly,
	})

	// Print agent info
	card := server.GetAgentCard()
	logger.Info("Agent card ready",
		"name", card.Name,
		"skills", len(card.Skills),
		"url", card.URL,
	)

	fmt.Fprintf(os.Stderr, "\n=== A2A Server Ready ===\n")
	fmt.Fprintf(os.Stderr, "Agent Card:    %s/.well-known/agent.json\n", card.URL)
	fmt.Fprintf(os.Stderr, "Health Check:  %s/health\n", card.URL)
	fmt.Fprintf(os.Stderr, "JSON-RPC:      POST %s/\n", card.URL)
	fmt.Fprintf(os.Stderr, "\nAvailable Skills:\n")
	for _, skill := range card.Skills {
		fmt.Fprintf(os.Stderr, "  - %s: %s\n", skill.ID, skill.Name)
	}
	fmt.Fprintf(os.Stderr, "\n")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Run()
	}()

	// Wait for signal or error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Shutdown error", "error", err)
		}
	case err := <-errChan:
		if err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}

func printVersion() {
	fmt.Printf("dm-nkp-gitops-a2a-server version %s\n", Version)
	fmt.Printf("  Git commit: %s\n", GitCommit)
	fmt.Printf("  Build time: %s\n", BuildTime)
}

func printUsage() {
	fmt.Println(`dm-nkp-gitops-a2a-server - A2A server for NKP GitOps infrastructure

This is an Agent-to-Agent (A2A) server that exposes the same tools as the
MCP server, but over HTTP for agent-to-agent communication.

USAGE:
    dm-nkp-gitops-a2a-server <command> [options]

COMMANDS:
    serve       Start the A2A HTTP server
    version     Show version information
    help        Show this help message

OPTIONS for 'serve':
    --port int            Port to listen on (default: 8080)
    --base-url string     Base URL for agent card (default: auto-generated)
    --kubeconfig string   Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)
    --context string      Kubernetes context to use (default: current context)
    --read-only           Enable read-only mode (default: true)
    --log-level string    Log level: debug, info, warn, error (default: info)

EXAMPLES:
    # Start A2A server on default port
    dm-nkp-gitops-a2a-server serve

    # Start on custom port with specific kubeconfig
    dm-nkp-gitops-a2a-server serve --port 9090 --kubeconfig=/path/to/config

    # Start with debug logging
    dm-nkp-gitops-a2a-server serve --log-level=debug

TESTING WITH CURL:
    # Get agent card (discovery)
    curl http://localhost:8080/.well-known/agent.json | jq

    # Check health
    curl http://localhost:8080/health | jq

    # Create a task (execute a skill)
    curl -X POST http://localhost:8080/ \
      -H "Content-Type: application/json" \
      -d '{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tasks/create",
        "params": {
          "skill": "get-gitops-status",
          "input": {}
        }
      }' | jq

    # Get task status
    curl -X POST http://localhost:8080/ \
      -H "Content-Type: application/json" \
      -d '{
        "jsonrpc": "2.0",
        "id": 2,
        "method": "tasks/get",
        "params": {"taskId": "TASK_ID_HERE"}
      }' | jq

KEY DIFFERENCES FROM MCP:
    MCP (stdio):
      - Direct tool invocation
      - Stateless
      - For AI assistant → tool communication

    A2A (HTTP):
      - Task-based execution
      - Stateful (tasks have lifecycle)
      - For agent → agent communication

For more information, see: docs/A2A_PROTOCOL.md`)
}
