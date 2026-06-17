// Package provider implements the OpenAI-compatible provider.
// Covers: OpenAI, DeepSeek, MiMo, Qwen, GLM, Moonshot, StepFun, MiniMax, etc.
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

// OpenAIProvider implements the OpenAI Chat Completions API.
type OpenAIProvider struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI-compatible provider.
func NewOpenAIProvider(baseURL, apiKey, model string, client *http.Client) *OpenAIProvider {
	return &OpenAIProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
	}
}

func (p *OpenAIProvider) Name() string        { return "openai" }
func (p *OpenAIProvider) DefaultModel() string { return p.model }

// openAIRequest is the request body for the Chat Completions API.
type openAIRequest struct {
	Model       string           `json:"model"`
	Messages    []openAIMessage  `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream"`
	Tools       []openAITool     `json:"tools,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type openAIToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function openAIFuncCall `json:"function"`
}

type openAIFuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// openAIResponse is the response from the Chat Completions API.
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content   string           `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (p *OpenAIProvider) buildMessages(messages []types.Message, systemPrompt string) []openAIMessage {
	var msgs []openAIMessage
	if systemPrompt != "" {
		msgs = append(msgs, openAIMessage{Role: "system", Content: systemPrompt})
	}
	for _, m := range messages {
		msg := openAIMessage{Role: string(m.Role), Content: m.Content}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			for _, tc := range m.ToolCalls {
				args, _ := json.Marshal(tc.Arguments)
				msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: openAIFuncCall{
						Name:      tc.Name,
						Arguments: string(args),
					},
				})
			}
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

func (p *OpenAIProvider) buildTools(tools []types.ToolDefinition) []openAITool {
	var oTools []openAITool
	for _, t := range tools {
		oTools = append(oTools, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return oTools
}

func (p *OpenAIProvider) Generate(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string) (*types.GenerateResult, error) {
	reqBody := openAIRequest{
		Model:       p.model,
		Messages:    p.buildMessages(messages, systemPrompt),
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

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var oResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &types.GenerateResult{
		Usage: types.TokenUsage{
			PromptTokens:     oResp.Usage.PromptTokens,
			CompletionTokens: oResp.Usage.CompletionTokens,
			TotalTokens:      oResp.Usage.PromptTokens + oResp.Usage.CompletionTokens,
		},
	}

	if len(oResp.Choices) > 0 {
		choice := oResp.Choices[0]
		result.Content = choice.Message.Content
		result.StopReason = choice.FinishReason

		for _, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &args)
			result.ToolCalls = append(result.ToolCalls, types.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			})
		}
	}

	return result, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string, onChunk func(types.StreamChunk)) (*types.GenerateResult, error) {
	reqBody := openAIRequest{
		Model:       p.model,
		Messages:    p.buildMessages(messages, systemPrompt),
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

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content   string           `json:"content"`
					ToolCalls []openAIToolCall  `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				result.Content += delta.Content
				onChunk(types.StreamChunk{Content: delta.Content})
			}
			if len(delta.ToolCalls) > 0 {
				for _, tc := range delta.ToolCalls {
					found := false
					for i, existing := range result.ToolCalls {
						if existing.ID == tc.ID {
							if result.ToolCalls[i].Arguments == nil {
								result.ToolCalls[i].Arguments = make(map[string]interface{})
							}
							result.ToolCalls[i].Name = tc.Function.Name
							found = true
							break
						}
					}
					if !found {
						result.ToolCalls = append(result.ToolCalls, types.ToolCall{
							ID:   tc.ID,
							Name: tc.Function.Name,
						})
					}
				}
			}
			if chunk.Choices[0].FinishReason != "" {
				result.StopReason = chunk.Choices[0].FinishReason
			}
		}
		if chunk.Usage != nil {
			result.Usage = types.TokenUsage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.PromptTokens + chunk.Usage.CompletionTokens,
			}
		}
	}

	onChunk(types.StreamChunk{Done: true, Usage: &result.Usage})
	return result, nil
}
