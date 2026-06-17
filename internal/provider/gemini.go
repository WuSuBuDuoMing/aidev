// Package provider implements the Google Gemini provider.
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

// GeminiProvider implements the Google AI (Gemini) API.
type GeminiProvider struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(baseURL, apiKey, model string, client *http.Client) *GeminiProvider {
	return &GeminiProvider{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
	}
}

func (p *GeminiProvider) Name() string        { return "gemini" }
func (p *GeminiProvider) DefaultModel() string { return p.model }

type geminiRequest struct {
	Contents         []geminiContent    `json:"contents"`
	SystemInstruction *geminiSystemInst `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenConfig  `json:"generationConfig,omitempty"`
	Tools             []geminiTool      `json:"tools,omitempty"`
}

type geminiContent struct {
	Role  string         `json:"role"`
	Parts []geminiPart   `json:"parts"`
}

type geminiPart struct {
	Text         string                  `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall     `json:"functionCall,omitempty"`
	FunctionResp *geminiFunctionResponse `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

type geminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

type geminiSystemInst struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenConfig struct {
	Temperature    float64 `json:"temperature,omitempty"`
	MaxOutputTokens int    `json:"maxOutputTokens,omitempty"`
}

type geminiTool struct {
	FunctionDeclarations []geminiFuncDecl `json:"function_declarations"`
}

type geminiFuncDecl struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (p *GeminiProvider) buildContents(messages []types.Message) []geminiContent {
	var contents []geminiContent
	for _, m := range messages {
		if m.Role == types.RoleSystem {
			continue // handled separately
		}
		role := "user"
		if m.Role == types.RoleAssistant {
			role = "model"
		}
		if m.Role == types.RoleTool {
			contents = append(contents, geminiContent{
				Role: "user",
				Parts: []geminiPart{{
					FunctionResp: &geminiFunctionResponse{
						Name:     m.ToolName,
						Response: map[string]interface{}{"content": m.Content},
					},
				}},
			})
			continue
		}
		if m.Role == types.RoleAssistant && len(m.ToolCalls) > 0 {
			var parts []geminiPart
			if m.Content != "" {
				parts = append(parts, geminiPart{Text: m.Content})
			}
			for _, tc := range m.ToolCalls {
				parts = append(parts, geminiPart{
					FunctionCall: &geminiFunctionCall{
						Name: tc.Name,
						Args: tc.Arguments,
					},
				})
			}
			contents = append(contents, geminiContent{Role: "model", Parts: parts})
			continue
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	return contents
}

func (p *GeminiProvider) Generate(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string) (*types.GenerateResult, error) {
	reqBody := geminiRequest{
		Contents: p.buildContents(messages),
		GenerationConfig: &geminiGenConfig{
			Temperature:    0.2,
			MaxOutputTokens: 4096,
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &geminiSystemInst{
			Parts: []geminiPart{{Text: systemPrompt}},
		}
	}
	if len(tools) > 0 {
		var decls []geminiFuncDecl
		for _, t := range tools {
			decls = append(decls, geminiFuncDecl{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
		reqBody.Tools = []geminiTool{{FunctionDeclarations: decls}}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var gResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&gResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := &types.GenerateResult{
		Usage: types.TokenUsage{
			PromptTokens:     gResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: gResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gResp.UsageMetadata.TotalTokenCount,
		},
	}

	if len(gResp.Candidates) > 0 {
		for _, part := range gResp.Candidates[0].Content.Parts {
			if part.FunctionCall != nil {
				result.ToolCalls = append(result.ToolCalls, types.ToolCall{
					ID:        fmt.Sprintf("gemini-%d", len(result.ToolCalls)),
					Name:      part.FunctionCall.Name,
					Arguments: part.FunctionCall.Args,
				})
			} else {
				result.Content += part.Text
			}
		}
	}

	return result, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string, onChunk func(types.StreamChunk)) (*types.GenerateResult, error) {
	reqBody := geminiRequest{
		Contents: p.buildContents(messages),
		GenerationConfig: &geminiGenConfig{
			Temperature:    0.2,
			MaxOutputTokens: 4096,
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &geminiSystemInst{
			Parts: []geminiPart{{Text: systemPrompt}},
		}
	}
	if len(tools) > 0 {
		var decls []geminiFuncDecl
		for _, t := range tools {
			decls = append(decls, geminiFuncDecl{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
		reqBody.Tools = []geminiTool{{FunctionDeclarations: decls}}
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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

		var evt struct {
			Candidates []struct {
				Content struct {
					Parts []geminiPart `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
			UsageMetadata *struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
				TotalTokenCount      int `json:"totalTokenCount"`
			} `json:"usageMetadata"`
		}

		if err := json.Unmarshal([]byte(payload), &evt); err != nil {
			continue
		}

		if len(evt.Candidates) > 0 {
			for _, part := range evt.Candidates[0].Content.Parts {
				if part.FunctionCall != nil {
					result.ToolCalls = append(result.ToolCalls, types.ToolCall{
						ID:        fmt.Sprintf("gemini-%d", len(result.ToolCalls)),
						Name:      part.FunctionCall.Name,
						Arguments: part.FunctionCall.Args,
					})
				} else if part.Text != "" {
					result.Content += part.Text
					onChunk(types.StreamChunk{Content: part.Text})
				}
			}
		}
		if evt.UsageMetadata != nil {
			result.Usage = types.TokenUsage{
				PromptTokens:     evt.UsageMetadata.PromptTokenCount,
				CompletionTokens: evt.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      evt.UsageMetadata.TotalTokenCount,
			}
		}
	}

	onChunk(types.StreamChunk{Done: true, Usage: &result.Usage})
	return result, nil
}
