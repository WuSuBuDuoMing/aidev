// Package agent implements the Plan mode for read-only analysis.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/provider"
	"github.com/WuSuBuDuoMing/aidev/internal/tool"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// PlanResult represents the result of a plan analysis.
type PlanResult struct {
	Summary  string
	Steps    []PlanStep
	Findings []string
}

// PlanStep is a single step in a plan.
type PlanStep struct {
	Action  string // "read", "search", "create", "edit", "run"
	Target  string // file path or command
	Details string
}

// RunPlan executes a plan-mode analysis: read-only exploration followed by a structured plan.
func RunPlan(ctx context.Context, task string, prov provider.Provider, registry *tool.Registry, modelID string) (*PlanResult, error) {
	// Build read-only tool definitions
	var readOnlyDefs []types.ToolDefinition
	for _, def := range registry.GetDefinitions() {
		t, ok := registry.Get(def.Name)
		if ok && t.ReadOnly() {
			readOnlyDefs = append(readOnlyDefs, def)
		}
	}

	systemPrompt := `You are in Plan Mode. You can ONLY read and analyze files — you MUST NOT modify anything.

Your job is to:
1. Explore the codebase to understand the current state
2. Analyze what changes are needed
3. Present a structured plan with numbered steps

Output format:
## Analysis
Brief summary of what you found.

## Plan
1. [action] Description of step
2. [action] Description of step
...

Where [action] is one of: read, create, edit, run

## Findings
- Key observations that inform the plan

Be thorough in your analysis. Read relevant files before proposing changes.`

	messages := []types.Message{
		{Role: types.RoleUser, Content: task, Timestamp: time.Now()},
	}

	// Run up to 5 rounds (read-only, so more rounds for thorough analysis)
	var result *types.GenerateResult
	var err error

	for round := 0; round < 5; round++ {
		result, err = prov.Generate(ctx, messages, readOnlyDefs, systemPrompt)
		if err != nil {
			return nil, fmt.Errorf("plan generation: %w", err)
		}

		messages = append(messages, types.Message{
			Role:      types.RoleAssistant,
			Content:   result.Content,
			Timestamp: time.Now(),
			ToolCalls: result.ToolCalls,
		})

		if len(result.ToolCalls) == 0 {
			break
		}

		// Execute read-only tool calls
		for _, tc := range result.ToolCalls {
			tr, _ := registry.Execute(ctx, tc.Name, tc.Arguments)
			messages = append(messages, types.Message{
				Role:       types.RoleTool,
				Content:    toolResultContent(tr),
				Timestamp:  time.Now(),
				ToolCallID: tc.ID,
				ToolName:   tc.Name,
			})
		}
	}

	// Parse the plan from the final response
	plan := &PlanResult{
		Summary: result.Content,
	}

	return plan, nil
}

// PrintPlan displays a plan result to the user.
func PrintPlan(plan *PlanResult) {
	fmt.Println("\n  📋 Plan Mode Result")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Println(plan.Summary)
}
