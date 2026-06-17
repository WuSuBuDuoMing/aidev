// Package provider defines the AI provider abstraction layer.
package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/config"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// Provider is the interface all AI providers must implement.
type Provider interface {
	// Generate produces a full (non-streaming) response.
	Generate(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string) (*types.GenerateResult, error)

	// Stream produces a streaming response, calling onChunk for each incremental piece.
	Stream(ctx context.Context, messages []types.Message, tools []types.ToolDefinition, systemPrompt string, onChunk func(types.StreamChunk)) (*types.GenerateResult, error)

	// Name returns the provider name.
	Name() string

	// DefaultModel returns the default model for this provider.
	DefaultModel() string
}

// Registry holds all registered providers.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

// Register adds a provider to the registry.
func (r *Registry) Register(name string, p Provider) {
	r.providers[name] = p
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// GetAll returns all registered providers.
func (r *Registry) GetAll() map[string]Provider {
	return r.providers
}

// ProviderKind returns the provider kind from config.
func ProviderKind(name string) string {
	switch name {
	case "claude":
		return "anthropic"
	case "openai":
		return "openai"
	case "deepseek":
		return "openai"
	case "gemini":
		return "gemini"
	case "mimo":
		return "openai"
	case "qwen":
		return "openai"
	case "glm":
		return "openai"
	case "moonshot":
		return "openai"
	case "stepfun":
		return "openai"
	case "minimax":
		return "openai"
	case "baidu":
		return "openai"
	case "volcengine":
		return "openai"
	case "bailing":
		return "openai"
	case "ollama":
		return "openai"
	default:
		return "openai"
	}
}

// DefaultBaseURL returns the default base URL for a provider.
func DefaultBaseURL(provider string) string {
	urls := map[string]string{
		"claude":     "https://api.anthropic.com",
		"openai":     "https://api.openai.com",
		"deepseek":   "https://api.deepseek.com",
		"gemini":     "https://generativelanguage.googleapis.com",
		"mimo":       "https://api.xiaomimimo.com/v1",
		"qwen":       "https://dashscope.aliyuncs.com/compatible-mode/v1",
		"glm":        "https://open.bigmodel.cn/api/paas/v4",
		"moonshot":   "https://api.moonshot.cn/v1",
		"stepfun":    "https://api.stepfun.com/v1",
		"minimax":    "https://api.minimax.chat/v1",
		"baidu":      "https://qianfan.baidubce.com/v2",
		"volcengine": "https://ark.cn-beijing.volces.com/api/v3",
		"bailing":    "https://api.tbox.cn/v1",
		"ollama":     "http://localhost:11434",
	}
	if u, ok := urls[provider]; ok {
		return u
	}
	return ""
}

// CreateProvider creates a provider instance from configuration.
func CreateProvider(cfg *config.Config) (Provider, string, error) {
	providerName := cfg.DefaultProvider
	modelName := cfg.DefaultModel

	// Check if there's a provider config override
	for _, p := range cfg.Providers {
		if p.Name == providerName {
			if p.BaseURL != "" {
				cfg.BaseURL = p.BaseURL
			}
			if p.Model != "" {
				modelName = p.Model
			}
			break
		}
	}

	// Determine base URL
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL(providerName)
	}

	// Determine provider kind
	kind := ProviderKind(providerName)
	for _, p := range cfg.Providers {
		if p.Name == providerName && p.Kind != "" {
			kind = p.Kind
		}
	}

	httpClient := &http.Client{Timeout: 5 * time.Minute}

	switch kind {
	case "anthropic":
		p := NewAnthropicProvider(baseURL, cfg.APIKey, modelName, httpClient)
		return p, modelName, nil
	case "gemini":
		p := NewGeminiProvider(baseURL, cfg.APIKey, modelName, httpClient)
		return p, modelName, nil
	case "openai", "custom":
		p := NewOpenAIProvider(baseURL, cfg.APIKey, modelName, httpClient)
		return p, modelName, nil
	default:
		return nil, "", fmt.Errorf("unsupported provider kind: %s", kind)
	}
}
