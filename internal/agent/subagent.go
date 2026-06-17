// Package agent implements the sub-agent system for specialized tasks.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/provider"
	"github.com/WuSuBuDuoMing/aidev/internal/tool"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// SubAgent defines a specialized agent with limited tools and a focused prompt.
type SubAgent struct {
	Name     string
	Tools    []string // Tool names this agent can use
	Prompt   string   // System prompt
	ReadOnly bool     // If true, only read-only tools
}

// builtinAgents defines the built-in sub-agents.
var builtinAgents = map[string]SubAgent{
	"explore": {
		Name:     "Explore Agent",
		Tools:    []string{"readFile", "searchFiles", "listDir", "grep"},
		ReadOnly: true,
		Prompt: `You are an explore agent. Your job is to search and read files to answer questions about the codebase.
You have read-only access. Do NOT attempt to modify any files.
Provide clear, structured findings with file paths and line numbers.`,
	},
	"review": {
		Name:     "Review Agent",
		Tools:    []string{"readFile", "searchFiles", "grep", "gitDiff", "gitStatus"},
		ReadOnly: true,
		Prompt: `You are a code review agent. Review code for:
1. Correctness — logic bugs, edge cases, off-by-one errors
2. Security — injection, XSS, SSRF, insecure defaults
3. Performance — N+1 queries, unnecessary allocations, missing caching
4. Readability — naming, structure, comments, dead code
5. Best practices — language idioms, error handling, typing

For each issue found, provide:
- File and line number
- Severity: critical / warning / suggestion
- A clear explanation of the problem
- A concrete fix or recommendation

End with a summary rating: APPROVE, REQUEST_CHANGES, or COMMENT.`,
	},
	"security": {
		Name:     "Security Agent",
		Tools:    []string{"readFile", "searchFiles", "grep", "listDir"},
		ReadOnly: true,
		Prompt: `You are a security audit agent. Analyze the codebase for security vulnerabilities:
1. SQL injection
2. Cross-site scripting (XSS)
3. Command injection
4. Path traversal
5. Insecure deserialization
6. Hardcoded secrets or API keys
7. Insecure defaults
8. Missing input validation

For each finding, provide:
- File and line number
- Vulnerability type (OWASP category)
- Severity: critical / high / medium / low
- Evidence (code snippet)
- Remediation recommendation`,
	},
	"research": {
		Name:     "Research Agent",
		Tools:    []string{"readFile", "searchFiles", "grep", "listDir"},
		ReadOnly: true,
		Prompt: `You are a research agent. Your job is to thoroughly investigate a specific topic in the codebase.
Find all relevant files, understand the architecture, and provide a comprehensive summary.
Include file paths, key function/class names, and how components interact.`,
	},
}

// GetSubAgent returns a built-in sub-agent by name.
func GetSubAgent(name string) (SubAgent, bool) {
	agent, ok := builtinAgents[name]
	return agent, ok
}

// ListSubAgents returns all available sub-agents.
func ListSubAgents() []SubAgent {
	var agents []SubAgent
	for _, a := range builtinAgents {
		agents = append(agents, a)
	}
	return agents
}

// RunSubAgent executes a sub-agent with a focused task and returns the result.
func RunSubAgent(ctx context.Context, name string, task string, prov provider.Provider, registry *tool.Registry, modelID string) (*types.GenerateResult, error) {
	sub, ok := builtinAgents[name]
	if !ok {
		return nil, fmt.Errorf("unknown sub-agent: %s", name)
	}

	// Build a filtered tool registry with only the tools this sub-agent needs
	filteredDefs := filterToolDefinitions(registry, sub.Tools)

	messages := []types.Message{
		{Role: types.RoleUser, Content: task, Timestamp: time.Now()},
	}

	// Run with limited tools (max 3 rounds for sub-agents)
	return runLimited(ctx, prov, filteredDefs, sub.Prompt, messages, modelID, 3)
}

// filterToolDefinitions returns only the tool definitions that match the given names.
func filterToolDefinitions(registry *tool.Registry, allowedNames []string) []types.ToolDefinition {
	allowed := make(map[string]bool)
	for _, name := range allowedNames {
		allowed[name] = true
	}

	var defs []types.ToolDefinition
	for _, def := range registry.GetDefinitions() {
		if allowed[def.Name] {
			defs = append(defs, def)
		}
	}
	return defs
}

// runLimited executes an agent loop with a maximum number of rounds.
func runLimited(ctx context.Context, prov provider.Provider, tools []types.ToolDefinition, systemPrompt string, messages []types.Message, modelID string, maxRounds int) (*types.GenerateResult, error) {
	registry := tool.NewRegistry() // empty registry for sub-agents (they use filtered defs)

	for round := 0; round < maxRounds; round++ {
		result, err := prov.Generate(ctx, messages, tools, systemPrompt)
		if err != nil {
			return nil, err
		}

		messages = append(messages, types.Message{
			Role:      types.RoleAssistant,
			Content:   result.Content,
			Timestamp: time.Now(),
			ToolCalls: result.ToolCalls,
		})

		if len(result.ToolCalls) == 0 {
			return result, nil
		}

		// Execute tool calls
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

	return nil, fmt.Errorf("sub-agent %s reached max rounds", systemPrompt[:20])
}
