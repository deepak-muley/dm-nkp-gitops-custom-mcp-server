// Package config provides configuration management for the MCP server.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds the server configuration.
type Config struct {
	Kubeconfig string
	Context    string
	ReadOnly   bool
	LogLevel   string
}

// K8sClients holds Kubernetes client instances.
type K8sClients struct {
	// Clientset for typed resources
	Clientset *kubernetes.Clientset

	// Dynamic client for unstructured resources (Flux, CAPI, etc.)
	Dynamic dynamic.Interface

	// REST config for creating additional clients
	RestConfig *rest.Config

	// Current context name
	CurrentContext string

	// Available contexts
	AvailableContexts []string
}

// ParseFlags parses command-line flags and environment variables.
func ParseFlags(args []string) *Config {
	cfg := &Config{
		Kubeconfig: os.Getenv("KUBECONFIG"),
		ReadOnly:   os.Getenv("MCP_READ_ONLY") == "true",
		LogLevel:   getEnvOrDefault("MCP_LOG_LEVEL", "info"),
	}

	// Parse args
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--read-only":
			cfg.ReadOnly = true
		case arg == "--kubeconfig" && i+1 < len(args):
			i++
			cfg.Kubeconfig = args[i]
		case strings.HasPrefix(arg, "--kubeconfig="):
			cfg.Kubeconfig = strings.TrimPrefix(arg, "--kubeconfig=")
		case arg == "--context" && i+1 < len(args):
			i++
			cfg.Context = args[i]
		case strings.HasPrefix(arg, "--context="):
			cfg.Context = strings.TrimPrefix(arg, "--context=")
		case arg == "--log-level" && i+1 < len(args):
			i++
			cfg.LogLevel = args[i]
		case strings.HasPrefix(arg, "--log-level="):
			cfg.LogLevel = strings.TrimPrefix(arg, "--log-level=")
		}
	}

	// Default kubeconfig location
	if cfg.Kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.Kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	return cfg
}

// LoadKubeConfig loads the Kubernetes configuration.
func LoadKubeConfig(kubeconfig, context string) (*rest.Config, error) {
	// Build config from kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfig != "" {
		loadingRules.ExplicitPath = kubeconfig
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		configOverrides.CurrentContext = context
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	return config, nil
}

// NewK8sClients creates Kubernetes clients from the config.
func NewK8sClients(config *rest.Config) (*K8sClients, error) {
	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Get available contexts
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, _ := os.UserHomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	var contexts []string
	var currentContext string

	rawConfig, err := clientcmd.LoadFromFile(kubeconfig)
	if err == nil {
		currentContext = rawConfig.CurrentContext
		for name := range rawConfig.Contexts {
			contexts = append(contexts, name)
		}
	}

	return &K8sClients{
		Clientset:         clientset,
		Dynamic:           dynamicClient,
		RestConfig:        config,
		CurrentContext:    currentContext,
		AvailableContexts: contexts,
	}, nil
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
