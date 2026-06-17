// Package provider implements the Anthropic (Claude) provider.
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// AnthropicProvider implements the Anthropic Messages API.
type AnthropicProvider struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider.
func NewAnthropicProvider(baseURL, apiKey, model string, client *http.Client) *AnthropicProvider {
	return &AnthropicProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
	}
}

func (p *AnthropicProvider) Name() string        { return "claude" }
func (p *AnthropicProvider) DefaultModel() string { return p.model }

// anthropicRequest is the request body for the Messages API.
type anthropicRequest struct {
	Model       string           `json:"model"`
	Messages    []anthropicMsg   `json:"messages"`
	System      string           `json:"system,omitempty"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream"`
	Tools       []anthropicTool  `json:"tools,omitempty"`
}

type anthropicMsg struct {
	Role    string        `json:"role"`
	Content interface{}   `json:"content"` // string or []contentBlock
}

type contentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content []contentBlock `json:"content"`
	Usage   struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

func (p *AnthropicProvider) buildMessages(messages []types.Message) []anthropicMsg {
	var msgs []anthropicMsg
	for _, m := range messages {
		if m.Role == types.RoleSystem {
			continue // system messages go in the top-level "system" field
		}
		if m.Role == types.RoleTool {
			// Tool results go in a user message with tool_result content blocks
			msgs = append(msgs, anthropicMsg{
				Role: "user",
				Content: []contentBlock{{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}},
			})
			continue
		}
		if m.Role == types.RoleAssistant && len(m.ToolCalls) > 0 {
			var blocks []contentBlock
			if m.Content != "" {
				blocks = append(blocks, contentBlock{Type: "text", Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, contentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Arguments,
				})
			}
			msgs = append(msgs, anthropicMsg{Role: "assistant", Content: blocks})
			continue
		}
		msgs = append(msgs, anthropicMsg{Role: string(m.Role), Content: m.Content})
	}
	return msgs
}

func (p *AnthropicProvider) buildTools(tools []types.ToolDefinition) []anthropicTool {
	var aTools []anthropicTool
	for _, t := range tools {
		aTools = append(aTools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.Parameters,
		})
	}
	return aTools
}

func (p *AnthropicProvider) Generate(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string) (*types.GenerateResult, error) {
	reqBody := anthropicRequest{
		Model:       p.model,
		Messages:    p.buildMessages(messages),
		System:      systemPrompt,
		MaxTokens:   4096,
		Temperature: 0.2,
		Stream:      false,
	}
	if len(tools) > 0 {
		reqBody.Tools = p.buildTools(tools)
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var aResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&aResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &types.GenerateResult{
		StopReason: aResp.StopReason,
		Usage: types.TokenUsage{
			PromptTokens:     aResp.Usage.InputTokens,
			CompletionTokens: aResp.Usage.OutputTokens,
			TotalTokens:      aResp.Usage.InputTokens + aResp.Usage.OutputTokens,
		},
	}

	for _, block := range aResp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, types.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: block.Input,
			})
		}
	}

	return result, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string, onChunk func(types.StreamChunk)) (*types.GenerateResult, error) {
	reqBody := anthropicRequest{
		Model:       p.model,
		Messages:    p.buildMessages(messages),
		System:      systemPrompt,
		MaxTokens:   4096,
		Temperature: 0.2,
		Stream:      true,
	}
	if len(tools) > 0 {
		reqBody.Tools = p.buildTools(tools)
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	result := &types.GenerateResult{}
	currentToolID := ""

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")

		var evt struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				Thinking    string `json:"thinking"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
			ContentBlock *struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
			Usage *struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			continue
		}

		switch evt.Type {
		case "content_block_start":
			if evt.ContentBlock != nil && evt.ContentBlock.Type == "tool_use" {
				currentToolID = evt.ContentBlock.ID
				result.ToolCalls = append(result.ToolCalls, types.ToolCall{
					ID:   evt.ContentBlock.ID,
					Name: evt.ContentBlock.Name,
				})
			}

		case "content_block_delta":
			if evt.Delta.Type == "text_delta" && evt.Delta.Text != "" {
				result.Content += evt.Delta.Text
				onChunk(types.StreamChunk{Content: evt.Delta.Text})
			}
			if evt.Delta.Type == "thinking_delta" && evt.Delta.Thinking != "" {
				result.Thinking += evt.Delta.Thinking
				onChunk(types.StreamChunk{Thinking: evt.Delta.Thinking})
			}

		case "message_delta":
			if evt.Usage != nil {
				result.Usage = types.TokenUsage{
					PromptTokens:     evt.Usage.InputTokens,
					CompletionTokens: evt.Usage.OutputTokens,
					TotalTokens:      evt.Usage.InputTokens + evt.Usage.OutputTokens,
				}
			}
			_ = currentToolID
		}
	}

	onChunk(types.StreamChunk{Done: true, Usage: &result.Usage})
	return result, nil
}
