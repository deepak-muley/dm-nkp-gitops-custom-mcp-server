// Package main demonstrates a Consensus pattern for multi-agent systems.
//
// Consensus Pattern: Multiple expert agents independently analyze the same problem
// and their results are aggregated to reach a consensus decision.
//
// This example shows:
// 1. Parallel execution of multiple agents
// 2. Independent analysis by each agent
// 3. Consensus calculation (voting/weighted average)
// 4. Confidence scoring
//
// Run:
//
//	go run examples/multi-agent/consensus/main.go
//
// Prerequisites:
//
//	Start multiple A2A server instances:
//	  make run-a2a                    # Port 8080
//	  ./bin/dm-nkp-gitops-a2a-server serve --port 8081  # Port 8081
//	  ./bin/dm-nkp-gitops-a2a-server serve --port 8082  # Port 8082
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/deepak-muley/dm-nkp-gitops-custom-mcp-server/pkg/a2a"
)

// AgentResult represents the result from a single agent
type AgentResult struct {
	AgentID    string
	Status     string
	Confidence float64
	Message    string
	Error      error
}

// ConsensusResult represents the aggregated consensus
type ConsensusResult struct {
	Decision   string  // "approve", "reject", "needs-review"
	Confidence float64 // 0.0 to 1.0
	Votes      map[string]int
	Details    []AgentResult
}

func main() {
	fmt.Println("=== A2A Consensus Pattern Demo ===")
	fmt.Println()
	fmt.Println("Scenario: Should we approve this deployment?")
	fmt.Println("Multiple expert agents will independently evaluate and vote.")
	fmt.Println()

	// Configuration - multiple agents (simulating different experts)
	agents := []string{
		getEnv("AGENT_1_URL", "http://localhost:8080"),
		getEnv("AGENT_2_URL", "http://localhost:8081"),
		getEnv("AGENT_3_URL", "http://localhost:8082"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run consensus workflow
	runConsensus(ctx, agents)
}

func runConsensus(ctx context.Context, agentURLs []string) {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Step 1: Agent Discovery                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Discover all agents
	clients := make(map[string]*a2a.Client)
	availableAgents := []string{}

	for i, url := range agentURLs {
		fmt.Printf("Discovering Agent %d at %s...\n", i+1, url)
		client := a2a.NewClient(url)
		card, err := client.GetAgentCard(ctx)
		if err != nil {
			fmt.Printf("  ⚠ Agent %d not available (skipping)\n", i+1)
			continue
		}
		fmt.Printf("  ✓ Found: %s\n", card.Name)
		clients[url] = client
		availableAgents = append(availableAgents, url)
	}

	if len(availableAgents) == 0 {
		fmt.Println("\n✗ No agents available!")
		fmt.Println("\nTo run this demo:")
		fmt.Println("  Terminal 1: make run-a2a                    # Port 8080")
		fmt.Println("  Terminal 2: ./bin/dm-nkp-gitops-a2a-server serve --port 8081")
		fmt.Println("  Terminal 3: ./bin/dm-nkp-gitops-a2a-server serve --port 8082")
		os.Exit(1)
	}

	fmt.Printf("\n✓ %d agent(s) available for consensus\n", len(availableAgents))
	fmt.Println()

	// Step 2: Parallel evaluation
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Step 2: Parallel Agent Evaluation                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("Question: Should we approve this deployment?")
	fmt.Println("Each agent will independently evaluate GitOps status...")
	fmt.Println()

	results := evaluateAgents(ctx, clients, availableAgents)

	// Step 3: Consensus calculation
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Step 3: Consensus Calculation                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	consensus := calculateConsensus(results)

	// Display results
	fmt.Println("Individual Agent Results:")
	fmt.Println("────────────────────────")
	for i, result := range results {
		fmt.Printf("Agent %d (%s):\n", i+1, result.AgentID)
		if result.Error != nil {
			fmt.Printf("  ✗ Error: %v\n", result.Error)
		} else {
			fmt.Printf("  Status: %s\n", result.Status)
			fmt.Printf("  Confidence: %.2f\n", result.Confidence)
			fmt.Printf("  Message: %s\n", truncate(result.Message, 60))
		}
		fmt.Println()
	}

	fmt.Println("Consensus Result:")
	fmt.Println("─────────────────")
	fmt.Printf("Decision: %s\n", consensus.Decision)
	fmt.Printf("Confidence: %.2f\n", consensus.Confidence)
	fmt.Printf("Votes: %v\n", consensus.Votes)
	fmt.Println()

	// Final decision
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          Final Decision                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	switch consensus.Decision {
	case "approve":
		fmt.Println("✅ CONSENSUS: APPROVE DEPLOYMENT")
		fmt.Printf("   High confidence (%.0f%%) - All agents agree\n", consensus.Confidence*100)
	case "reject":
		fmt.Println("❌ CONSENSUS: REJECT DEPLOYMENT")
		fmt.Printf("   Issues detected (%.0f%% confidence)\n", consensus.Confidence*100)
	case "needs-review":
		fmt.Println("⚠️  CONSENSUS: NEEDS MANUAL REVIEW")
		fmt.Printf("   Agents disagree (%.0f%% confidence)\n", consensus.Confidence*100)
	}

	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Multiple agents independently evaluate the same question")
	fmt.Println("  2. Each agent provides its own answer with confidence score")
	fmt.Println("  3. Consensus mechanism aggregates results")
	fmt.Println("  4. Final decision based on collective intelligence")
	fmt.Println()
	fmt.Println("Use cases:")
	fmt.Println("  - Critical deployment approvals")
	fmt.Println("  - Security policy decisions")
	fmt.Println("  - Quality assurance checks")
	fmt.Println("  - Expert consultation scenarios")
}

func evaluateAgents(ctx context.Context, clients map[string]*a2a.Client, agentURLs []string) []AgentResult {
	var wg sync.WaitGroup
	results := make([]AgentResult, len(agentURLs))
	resultChan := make(chan AgentResult, len(agentURLs))

	// Execute all agents in parallel
	for i, url := range agentURLs {
		wg.Add(1)
		go func(idx int, agentURL string) {
			defer wg.Done()

			client := clients[agentURL]
			result := AgentResult{
				AgentID: fmt.Sprintf("agent-%d", idx+1),
			}

			fmt.Printf("  → Agent %d evaluating...\n", idx+1)

			// Execute skill (simulating expert evaluation)
			task, err := client.ExecuteSkill(ctx, "get-gitops-status", map[string]interface{}{}, 30*time.Second)
			if err != nil {
				result.Error = err
				resultChan <- result
				return
			}

			// Simulate confidence based on task status
			// In real scenario, agents would return confidence scores
			if task.Status == a2a.TaskStatusCompleted {
				result.Status = "healthy"
				result.Confidence = 0.9 // High confidence if GitOps is healthy
				result.Message = "GitOps status check passed"
			} else {
				result.Status = "unhealthy"
				result.Confidence = 0.3 // Low confidence if issues found
				result.Message = "GitOps status check found issues"
			}

			resultChan <- result
		}(i, url)
	}

	wg.Wait()
	close(resultChan)

	// Collect results
	idx := 0
	for result := range resultChan {
		results[idx] = result
		idx++
	}

	return results
}

func calculateConsensus(results []AgentResult) ConsensusResult {
	consensus := ConsensusResult{
		Votes:   make(map[string]int),
		Details: results,
	}

	// Count votes (approve/reject based on confidence threshold)
	approveCount := 0
	totalConfidence := 0.0
	validResults := 0

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		validResults++
		totalConfidence += result.Confidence

		// Vote based on confidence threshold
		if result.Confidence >= 0.7 {
			consensus.Votes["approve"]++
			approveCount++
		} else {
			consensus.Votes["reject"]++
		}
	}

	if validResults == 0 {
		consensus.Decision = "needs-review"
		consensus.Confidence = 0.0
		return consensus
	}

	// Calculate average confidence
	consensus.Confidence = totalConfidence / float64(validResults)

	// Determine decision
	approveRatio := float64(approveCount) / float64(validResults)

	if approveRatio >= 0.67 {
		consensus.Decision = "approve"
	} else if approveRatio <= 0.33 {
		consensus.Decision = "reject"
	} else {
		consensus.Decision = "needs-review"
	}

	return consensus
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
