// Package agent implements the core AI agent loop with tool calling,
// storm breaker, parallel dispatch, and context compaction.
package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/config/model"
	"github.com/WuSuBuDuoMing/aidev/internal/permission"
	"github.com/WuSuBuDuoMing/aidev/internal/provider"
	"github.com/WuSuBuDuoMing/aidev/internal/tool"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

const (
	// MaxToolRounds is the maximum number of tool call rounds.
	MaxToolRounds = 10
	// StormBreakThreshold is the number of identical (tool, error) signatures before aborting.
	StormBreakThreshold = 3
)

// Options configures the agent loop.
type Options struct {
	Provider       provider.Provider
	Tools          *tool.Registry
	SystemPrompt   string
	ModelID        string
	ContextWindow  int
	PermissionMode types.PermissionModeName
	PermissionRules []types.PermissionEntry
	OnToken        func(chunk types.StreamChunk)
	OnToolStart    func(name string, args map[string]interface{})
	OnToolEnd      func(name string, result *types.ToolResult)
	OnRetry        func(attempt int, delay time.Duration, reason string)
	Stream         bool
}

// StormBreaker detects death-spiral loops by (tool, error) signature.
type StormBreaker struct {
	signatures map[string]int
}

// NewStormBreaker creates a new storm breaker.
func NewStormBreaker() *StormBreaker {
	return &StormBreaker{signatures: make(map[string]int)}
}

// Record records a failed tool call. Returns true if the threshold is breached.
func (s *StormBreaker) Record(toolName, errStr string) bool {
	if errStr == "" {
		return false
	}
	// Key on first 120 chars of error to group similar failures
	sig := toolName + ":" + errStr[:min(120, len(errStr))]
	s.signatures[sig]++
	return s.signatures[sig] >= StormBreakThreshold
}

// Run executes the agent loop with the given messages and returns the final result.
func Run(ctx context.Context, opts Options, messages []types.Message) (*types.GenerateResult, []types.Message, error) {
	if opts.OnToken == nil {
		opts.OnToken = func(types.StreamChunk) {}
	}

	tools := opts.Tools.GetDefinitions()
	storm := NewStormBreaker()
	policy := permission.NewPolicy(opts.PermissionMode, opts.PermissionRules, "")
	spec := model.GetSpec(opts.ModelID)
	compactThreshold := model.CompactThreshold(spec.ContextWindow)
	_ = compactThreshold // will be used for compaction

	var allMessages []types.Message
	allMessages = append(allMessages, messages...)

	for round := 0; round < MaxToolRounds; round++ {
		var result *types.GenerateResult
		var err error

		if opts.Stream {
			result, err = opts.Provider.Stream(ctx, allMessages, tools, opts.SystemPrompt, opts.OnToken)
		} else {
			result, err = opts.Provider.Generate(ctx, allMessages, tools, opts.SystemPrompt)
		}

		if err != nil {
			return nil, allMessages, fmt.Errorf("provider error: %w", err)
		}

		// Store assistant response
		allMessages = append(allMessages, types.Message{
			Role:      types.RoleAssistant,
			Content:   result.Content,
			Timestamp: time.Now(),
			ToolCalls: result.ToolCalls,
		})

		// No tool calls — we're done
		if len(result.ToolCalls) == 0 {
			return result, allMessages, nil
		}

		// Check if all tool calls are read-only for parallel dispatch
		allReadOnly := opts.Tools.AllReadOnly(result.ToolCalls)

		if allReadOnly && len(result.ToolCalls) > 1 {
			// Parallel dispatch for all-readOnly batches
			type toolResult struct {
				call   types.ToolCall
				result *types.ToolResult
			}

			ch := make(chan toolResult, len(result.ToolCalls))
			for _, tc := range result.ToolCalls {
				go func(call types.ToolCall) {
					if opts.OnToolStart != nil {
						opts.OnToolStart(call.Name, call.Arguments)
					}

					dec := policy.Evaluate(call.Name, call.Arguments)
					var tr *types.ToolResult

					if dec == permission.Deny {
						tr = &types.ToolResult{Success: false, Error: "Permission denied"}
					} else {
						tr, _ = opts.Tools.Execute(ctx, call.Name, call.Arguments)
					}

					if opts.OnToolEnd != nil {
						opts.OnToolEnd(call.Name, tr)
					}
					ch <- toolResult{call: call, result: tr}
				}(tc)
			}

			for i := 0; i < len(result.ToolCalls); i++ {
				tr := <-ch
				allMessages = append(allMessages, types.Message{
					Role:       types.RoleTool,
					Content:    toolResultContent(tr.result),
					Timestamp:  time.Now(),
					ToolCallID: tr.call.ID,
					ToolName:   tr.call.Name,
				})

				// Storm breaker check
				if !tr.result.Success && storm.Record(tr.call.Name, tr.result.Error) {
					allMessages = append(allMessages, types.Message{
						Role:      types.RoleSystem,
						Content:   fmt.Sprintf("[System: Tool loop detected — same error %q repeating. Try a different approach.]", tr.result.Error[:min(80, len(tr.result.Error))]),
						Timestamp: time.Now(),
					})
					return result, allMessages, nil
				}
			}
		} else {
			// Sequential dispatch
			for _, tc := range result.ToolCalls {
				if opts.OnToolStart != nil {
					opts.OnToolStart(tc.Name, tc.Arguments)
				}

				dec := policy.Evaluate(tc.Name, tc.Arguments)
				var tr *types.ToolResult

				if dec == permission.Deny {
					tr = &types.ToolResult{Success: false, Error: "Permission denied"}
				} else {
					tr, _ = opts.Tools.Execute(ctx, tc.Name, tc.Arguments)
				}

				if opts.OnToolEnd != nil {
					opts.OnToolEnd(tc.Name, tr)
				}

				allMessages = append(allMessages, types.Message{
					Role:       types.RoleTool,
					Content:    toolResultContent(tr),
					Timestamp:  time.Now(),
					ToolCallID: tc.ID,
					ToolName:   tc.Name,
				})

				// Storm breaker check
				if !tr.Success && storm.Record(tc.Name, tr.Error) {
					allMessages = append(allMessages, types.Message{
						Role:      types.RoleSystem,
						Content:   fmt.Sprintf("[System: Tool loop detected — same error %q repeating. Try a different approach.]", tr.Error[:min(80, len(tr.Error))]),
						Timestamp: time.Now(),
					})
					return result, allMessages, nil
				}
			}
		}
	}

	return nil, allMessages, fmt.Errorf("reached maximum tool call rounds (%d)", MaxToolRounds)
}

// toolResultContent formats a tool result for the message.
func toolResultContent(tr *types.ToolResult) string {
	if tr.Success {
		return tr.Output
	}
	return "Error: " + tr.Error
}

// EstimateTokens provides a rough token estimate from character count.
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4 // ~4 chars per token for English
}

// EstimateConversationTokens estimates total tokens in a conversation.
func EstimateConversationTokens(messages []types.Message) int {
	total := 0
	for _, m := range messages {
		total += EstimateTokens(m.Content)
		for _, tc := range m.ToolCalls {
			for _, v := range tc.Arguments {
				total += EstimateTokens(fmt.Sprintf("%v", v))
			}
		}
	}
	return total
}

// CompactIfNeeded performs context compaction if the conversation is too long.
// Returns true if compaction was performed.
func CompactIfNeeded(messages []types.Message, threshold int) ([]types.Message, bool) {
	totalTokens := EstimateConversationTokens(messages)
	if totalTokens < threshold {
		return messages, false
	}

	// Keep recent messages (last 6 turns) and replace older ones with a summary
	recentCount := min(6, len(messages))
	keepRecent := messages[len(messages)-recentCount:]
	foldable := messages[:len(messages)-recentCount]

	if len(foldable) == 0 {
		return messages, false
	}

	// Count what we're replacing
	userTurns := 0
	assistantTurns := 0
	toolCalls := 0
	for _, m := range foldable {
		switch m.Role {
		case types.RoleUser:
			userTurns++
		case types.RoleAssistant:
			assistantTurns++
			toolCalls += len(m.ToolCalls)
		}
	}

	// Keep first system message
	var systemMsg *types.Message
	for _, m := range foldable {
		if m.Role == types.RoleSystem {
			msg := m
			systemMsg = &msg
			break
		}
	}

	summary := types.Message{
		Role:      types.RoleSystem,
		Content:   fmt.Sprintf("[Context Summary: %d user turns, %d assistant responses, %d tool calls. Conversation auto-compacted.]", userTurns, assistantTurns, toolCalls),
		Timestamp: time.Now(),
	}

	var result []types.Message
	if systemMsg != nil {
		result = append(result, *systemMsg)
	}
	result = append(result, summary)
	result = append(result, keepRecent...)

	return result, true
}

// PruneToolResults replaces old, re-derivable tool outputs with placeholders.
func PruneToolResults(messages []types.Message) []types.Message {
	// Only prune messages older than the last 8
	cutoff := max(0, len(messages)-8)
	prunable := map[string]bool{
		"readFile":    true,
		"searchFiles": true,
		"listDir":     true,
		"gitStatus":   true,
		"gitDiff":     true,
		"grep":        true,
	}

	for i := 0; i < cutoff; i++ {
		m := &messages[i]
		if m.Role != types.RoleTool || m.ToolName == "" {
			continue
		}
		if !prunable[m.ToolName] {
			continue
		}
		if strings.HasPrefix(m.Content, "[Pruned:") {
			continue
		}
		if len(m.Content) > 500 {
			m.Content = fmt.Sprintf("[Pruned: %s output (%d chars). Re-run the tool to get fresh data.]", m.ToolName, len(m.Content))
		}
	}

	return messages
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
