// Package main demonstrates a Pipeline pattern for multi-agent systems.
//
// Pipeline Pattern: Sequential data transformation through multiple agents
//   Agent A (gather) → Agent B (process) → Agent C (output)
//
// This example shows:
// 1. Sequential agent execution
// 2. Data passing between agents
// 3. Progressive data enrichment
//
// Run:
//
//	go run examples/multi-agent/pipeline/main.go
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
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/a2a"
)

func main() {
	fmt.Println("=== A2A Pipeline Pattern Demo ===")
	fmt.Println()
	fmt.Println("Pipeline: Data Collection → Analysis → Report Generation")
	fmt.Println()

	// Configuration
	agentURL := getEnv("AGENT_URL", "http://localhost:8080")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create client
	client := a2a.NewClient(agentURL)

	// Discover agent
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Step 1: Agent Discovery                           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	card, err := client.GetAgentCard(ctx)
	if err != nil {
		fmt.Printf("✗ Failed to discover agent: %v\n", err)
		fmt.Println("\nMake sure the A2A server is running: make run-a2a")
		os.Exit(1)
	}
	fmt.Printf("✓ Connected to: %s\n", card.Name)
	fmt.Printf("  Available skills: %d\n", len(card.Skills))
	fmt.Println()

	// Run pipeline
	runPipeline(ctx, client)
}

func runPipeline(ctx context.Context, client *a2a.Client) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Pipeline Execution                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Stage 1: Data Collection
	fmt.Println("Stage 1: Data Collection")
	fmt.Println("─────────────────────────")
	fmt.Println("  Collecting GitOps status and context information...")

	stage1Input := map[string]interface{}{}
	stage1Result, err := executeStage(ctx, client, "get-gitops-status", stage1Input, "Collecting")
	if err != nil {
		fmt.Printf("  ✗ Stage 1 failed: %v\n", err)
		return
	}

	// Extract data from stage 1
	stage1Data := extractTextFromMessages(stage1Result.Messages)
	fmt.Printf("  ✓ Collected %d bytes of data\n", len(stage1Data))
	fmt.Println()

	// Stage 2: Analysis (simulated - in real scenario, this would be a different agent)
	fmt.Println("Stage 2: Data Analysis")
	fmt.Println("──────────────────────")
	fmt.Println("  Analyzing collected data...")

	// In a real pipeline, stage2 would call a different agent
	// For demo, we'll use the same agent but simulate analysis
	// stage2Input would contain: data from stage1, analysis_type, etc.
	// For this demo, we simulate analysis by getting current context
	stage2Result, err := executeStage(ctx, client, "get-current-context", map[string]interface{}{}, "Analyzing")
	if err != nil {
		fmt.Printf("  ✗ Stage 2 failed: %v\n", err)
		return
	}

	stage2Data := extractTextFromMessages(stage2Result.Messages)
	fmt.Printf("  ✓ Analysis complete\n")
	fmt.Printf("  Analysis result: %s\n", truncate(stage2Data, 100))
	fmt.Println()

	// Stage 3: Report Generation
	fmt.Println("Stage 3: Report Generation")
	fmt.Println("──────────────────────────")
	fmt.Println("  Generating final report...")

	// Combine data from previous stages
	reportData := fmt.Sprintf(`
=== GitOps Status Report ===
Generated: %s

Data Collection:
%s

Analysis:
%s

Summary:
- GitOps status collected successfully
- Context information retrieved
- Report generated
`, time.Now().Format(time.RFC3339), truncate(stage1Data, 200), truncate(stage2Data, 100))

	fmt.Println("  ✓ Report generated")
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Final Report                                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println(reportData)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Pipeline Complete!                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Pipeline executes stages sequentially")
	fmt.Println("  2. Each stage receives output from previous stage")
	fmt.Println("  3. Data is transformed/enriched at each stage")
	fmt.Println("  4. Final stage produces the end result")
	fmt.Println()
	fmt.Println("Real-world example:")
	fmt.Println("  Stage 1: Data Collector → Gathers logs, metrics, events")
	fmt.Println("  Stage 2: Analyzer → Identifies patterns and issues")
	fmt.Println("  Stage 3: Report Generator → Creates formatted report")
}

func executeStage(ctx context.Context, client *a2a.Client, skillID string, input map[string]interface{}, stageName string) (*a2a.Task, error) {
	fmt.Printf("  → Executing: %s\n", skillID)

	start := time.Now()
	task, err := client.ExecuteSkill(ctx, skillID, input, 30*time.Second)
	duration := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	if task.Status == a2a.TaskStatusCompleted {
		fmt.Printf("  ✓ Completed in %v\n", duration)
	} else {
		return nil, fmt.Errorf("task ended with status: %s", task.Status)
	}

	return task, nil
}

func extractTextFromMessages(messages []a2a.Message) string {
	var textParts []string
	for _, msg := range messages {
		for _, content := range msg.Content {
			if content.Type == "text" {
				textParts = append(textParts, content.Text)
			}
		}
	}
	return strings.Join(textParts, "\n")
}

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
