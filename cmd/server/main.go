// Package main provides the entry point for the dm-nkp-gitops MCP server.
//
// This MCP (Model Context Protocol) server provides AI assistants with tools
// to monitor and debug GitOps infrastructure managed by Flux CD on NKP clusters.
package main

import (
	"fmt"
	"os"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/config"
	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/mcp"
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
	cfg := config.ParseFlags(os.Args[2:])

	// Setup logging
	logger := config.NewLogger(cfg.LogLevel)

	logger.Info("Starting dm-nkp-gitops MCP server",
		"version", Version,
		"commit", GitCommit,
		"readOnly", cfg.ReadOnly,
	)

	// Load Kubernetes configuration
	k8sConfig, err := config.LoadKubeConfig(cfg.Kubeconfig, cfg.Context)
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

	// Register tools
	registry := tools.NewRegistry(clients, cfg.ReadOnly, logger)
	registry.RegisterAllTools()

	// Create and run MCP server
	server := mcp.NewServer(mcp.ServerConfig{
		Name:        "dm-nkp-gitops-mcp-server",
		Version:     Version,
		Description: "MCP server for NKP GitOps infrastructure monitoring and debugging",
		Tools:       registry.GetTools(),
		Handlers:    registry.GetHandlers(),
		Logger:      logger,
	})

	// Run server (blocks until stdin is closed)
	if err := server.Run(); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("dm-nkp-gitops-mcp-server version %s\n", Version)
	fmt.Printf("  Git commit: %s\n", GitCommit)
	fmt.Printf("  Build time: %s\n", BuildTime)
}

func printUsage() {
	fmt.Println(`dm-nkp-gitops-mcp-server - MCP server for NKP GitOps infrastructure

USAGE:
    dm-nkp-gitops-mcp-server <command> [options]

COMMANDS:
    serve       Start the MCP server (communicates via stdin/stdout)
    version     Show version information
    help        Show this help message

OPTIONS for 'serve':
    --kubeconfig string   Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)
    --context string      Kubernetes context to use (default: current context)
    --read-only           Enable read-only mode (no mutations allowed)
    --log-level string    Log level: debug, info, warn, error (default: info)

ENVIRONMENT VARIABLES:
    KUBECONFIG            Path to kubeconfig file
    MCP_READ_ONLY         Set to "true" for read-only mode
    MCP_LOG_LEVEL         Log level

EXAMPLES:
    # Start server with default kubeconfig
    dm-nkp-gitops-mcp-server serve

    # Start in read-only mode with specific kubeconfig
    dm-nkp-gitops-mcp-server serve --read-only --kubeconfig=/path/to/config

    # Start with debug logging
    dm-nkp-gitops-mcp-server serve --log-level=debug

CURSOR CONFIGURATION:
    Add to ~/.cursor/mcp.json:
    {
      "mcpServers": {
        "dm-nkp-gitops": {
          "command": "/path/to/dm-nkp-gitops-mcp-server",
          "args": ["serve", "--read-only"],
          "env": {
            "KUBECONFIG": "/path/to/kubeconfig"
          }
        }
      }
    }

For more information, see: https://github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server`)
}
