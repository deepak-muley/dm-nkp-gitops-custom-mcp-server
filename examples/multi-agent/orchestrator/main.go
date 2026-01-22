// Package main demonstrates an A2A orchestrator that coordinates multiple agents.
//
// This example shows:
// 1. Agent discovery using Agent Cards
// 2. Task-based skill execution
// 3. Coordination across multiple agents
// 4. Async task management
//
// Run:
//
//	go run examples/multi-agent/orchestrator/main.go
//
// Prerequisites:
//
//	Start the A2A server first: make run-a2a
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/a2a"
)

func main() {
	fmt.Println("=== A2A Multi-Agent Orchestrator Demo ===")
	fmt.Println()

	// Configuration
	gitopsAgentURL := getEnv("GITOPS_AGENT_URL", "http://localhost:8080")
	policyAgentURL := getEnv("POLICY_AGENT_URL", "http://localhost:8081")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run demo scenarios
	runDemo(ctx, gitopsAgentURL, policyAgentURL)
}

func runDemo(ctx context.Context, gitopsURL, policyURL string) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 1: Agent Discovery                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Create clients for each agent
	gitopsClient := a2a.NewClient(gitopsURL)
	policyClient := a2a.NewClient(policyURL)

	// Discover GitOps Agent
	fmt.Printf("Discovering GitOps Agent at %s...\n", gitopsURL)
	gitopsCard, err := gitopsClient.GetAgentCard(ctx)
	if err != nil {
		fmt.Printf("  ✗ Failed to discover GitOps Agent: %v\n", err)
		fmt.Println("\n  Make sure the A2A server is running: make run-a2a")
		os.Exit(1)
	}
	printAgentCard(gitopsCard)

	// Try to discover Policy Agent (optional)
	fmt.Printf("\nDiscovering Policy Agent at %s...\n", policyURL)
	policyCard, err := policyClient.GetAgentCard(ctx)
	hasPolicyAgent := err == nil
	if hasPolicyAgent {
		printAgentCard(policyCard)
	} else {
		fmt.Printf("  ⚠ Policy Agent not available (optional for demo)\n")
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 2: Simple Skill Execution                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 1: List Kubernetes contexts
	fmt.Println("Task 1: List Kubernetes Contexts")
	fmt.Println("─────────────────────────────────")
	runSkill(ctx, gitopsClient, "list-contexts", nil)

	// Example 2: Get current context
	fmt.Println("\nTask 2: Get Current Context")
	fmt.Println("───────────────────────────")
	runSkill(ctx, gitopsClient, "get-current-context", nil)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 3: GitOps Status Check                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 3: Get GitOps status
	fmt.Println("Task 3: Get GitOps Status")
	fmt.Println("─────────────────────────")
	runSkill(ctx, gitopsClient, "get-gitops-status", map[string]interface{}{})

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 4: Parallel Agent Execution               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 4: Parallel execution
	fmt.Println("Running multiple skills in parallel...")
	fmt.Println("───────────────────────────────────────")
	runParallelSkills(ctx, gitopsClient)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 5: Task Lifecycle Demo                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Example 5: Show task lifecycle
	fmt.Println("Demonstrating Task Lifecycle")
	fmt.Println("────────────────────────────")
	demonstrateTaskLifecycle(ctx, gitopsClient)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Phase 6: Multi-Agent Coordination               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if hasPolicyAgent {
		fmt.Println("Running multi-agent workflow...")
		runMultiAgentWorkflow(ctx, gitopsClient, policyClient)
	} else {
		fmt.Println("⚠ Skipping multi-agent workflow (Policy Agent not available)")
		fmt.Println("  To run this: start a second A2A server on port 8081")
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Demo Complete!                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Agent Discovery: /.well-known/agent.json exposes capabilities")
	fmt.Println("  2. Task-based: Skills execute as tracked tasks (not direct calls)")
	fmt.Println("  3. Async: Tasks run asynchronously, poll for completion")
	fmt.Println("  4. Multi-agent: Orchestrator coordinates multiple specialized agents")
	fmt.Println()
	fmt.Println("Compare with MCP:")
	fmt.Println("  MCP:  tools/call → immediate result")
	fmt.Println("  A2A:  tasks/create → task ID → tasks/get → result")
}

// printAgentCard displays agent capabilities
func printAgentCard(card *a2a.AgentCard) {
	fmt.Printf("  ✓ Found: %s v%s\n", card.Name, card.Version)
	fmt.Printf("    Description: %s\n", truncate(card.Description, 60))
	fmt.Printf("    Skills: %d available\n", len(card.Skills))

	// Show first few skills
	for i, skill := range card.Skills {
		if i >= 5 {
			fmt.Printf("      ... and %d more\n", len(card.Skills)-5)
			break
		}
		fmt.Printf("      - %s\n", skill.ID)
	}
}

// runSkill executes a skill and displays the result
func runSkill(ctx context.Context, client *a2a.Client, skillID string, input map[string]interface{}) {
	fmt.Printf("  Executing skill: %s\n", skillID)

	start := time.Now()
	task, err := client.ExecuteSkill(ctx, skillID, input, 30*time.Second)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
		return
	}

	fmt.Printf("  ✓ Completed in %v (status: %s)\n", duration, task.Status)

	// Show result preview
	if len(task.Messages) > 0 {
		for _, msg := range task.Messages {
			for _, content := range msg.Content {
				if content.Type == "text" {
					preview := truncate(content.Text, 200)
					fmt.Printf("  Result preview:\n")
					for _, line := range strings.Split(preview, "\n") {
						fmt.Printf("    %s\n", line)
					}
				}
			}
		}
	}
}

// runParallelSkills demonstrates parallel execution
func runParallelSkills(ctx context.Context, client *a2a.Client) {
	skills := []string{
		"list-contexts",
		"get-current-context",
	}

	var wg sync.WaitGroup
	results := make(chan string, len(skills))

	start := time.Now()

	for _, skillID := range skills {
		wg.Add(1)
		go func(sid string) {
			defer wg.Done()

			task, err := client.ExecuteSkill(ctx, sid, nil, 30*time.Second)
			if err != nil {
				results <- fmt.Sprintf("  ✗ %s: %v", sid, err)
			} else {
				results <- fmt.Sprintf("  ✓ %s: %s", sid, task.Status)
			}
		}(skillID)
	}

	wg.Wait()
	close(results)

	duration := time.Since(start)

	fmt.Printf("  Completed %d skills in parallel (%v total):\n", len(skills), duration)
	for result := range results {
		fmt.Println(result)
	}
}

// demonstrateTaskLifecycle shows the task state machine
func demonstrateTaskLifecycle(ctx context.Context, client *a2a.Client) {
	fmt.Println("  1. Creating task...")

	// Create task (don't wait)
	task, err := client.CreateTask(ctx, "get-gitops-status", map[string]interface{}{})
	if err != nil {
		fmt.Printf("  ✗ Failed to create task: %v\n", err)
		return
	}
	fmt.Printf("     Task ID: %s\n", task.ID)
	fmt.Printf("     Initial Status: %s\n", task.Status)

	fmt.Println("  2. Polling for completion...")

	// Poll a few times to show status changes
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)

		current, err := client.GetTask(ctx, task.ID)
		if err != nil {
			fmt.Printf("  ✗ Failed to get task: %v\n", err)
			return
		}

		fmt.Printf("     Poll %d: %s\n", i+1, current.Status)

		if current.Status == a2a.TaskStatusCompleted ||
			current.Status == a2a.TaskStatusFailed {
			break
		}
	}

	// Final state
	final, _ := client.GetTask(ctx, task.ID)
	fmt.Printf("  3. Final Status: %s\n", final.Status)
	fmt.Printf("     Messages: %d, Artifacts: %d\n", len(final.Messages), len(final.Artifacts))
}

// runMultiAgentWorkflow coordinates multiple agents
func runMultiAgentWorkflow(ctx context.Context, gitopsClient, policyClient *a2a.Client) {
	fmt.Println("  Running coordinated workflow:")
	fmt.Println("    Step 1: Check GitOps status")
	fmt.Println("    Step 2: Check policy violations")
	fmt.Println("    Step 3: Synthesize results")
	fmt.Println()

	var wg sync.WaitGroup
	var gitopsResult, policyResult *a2a.Task
	var gitopsErr, policyErr error

	// Step 1 & 2: Parallel execution
	wg.Add(2)

	go func() {
		defer wg.Done()
		gitopsResult, gitopsErr = gitopsClient.ExecuteSkill(ctx, "get-gitops-status", nil, 30*time.Second)
	}()

	go func() {
		defer wg.Done()
		policyResult, policyErr = policyClient.ExecuteSkill(ctx, "check-policy-violations", nil, 30*time.Second)
	}()

	wg.Wait()

	// Step 3: Synthesize
	fmt.Println("  Results:")

	if gitopsErr != nil {
		fmt.Printf("    ✗ GitOps: %v\n", gitopsErr)
	} else {
		fmt.Printf("    ✓ GitOps: %s\n", gitopsResult.Status)
	}

	if policyErr != nil {
		fmt.Printf("    ✗ Policy: %v\n", policyErr)
	} else {
		fmt.Printf("    ✓ Policy: %s\n", policyResult.Status)
	}

	fmt.Println()
	fmt.Println("  Synthesis: Both agents queried successfully!")
	fmt.Println("  (In production, you'd analyze and correlate the results)")
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
