// Package model defines the model registry with pricing and capabilities.
package model

// Spec describes a model's capabilities and pricing.
type Spec struct {
	ID              string  `json:"id"`
	Provider        string  `json:"provider"`
	ContextWindow   int     `json:"contextWindow"`
	MaxOutput       int     `json:"maxOutput"`
	SupportsTools   bool    `json:"supportsTools"`
	SupportsVision  bool    `json:"supportsVision"`
	SupportsReasoning bool  `json:"supportsReasoning"`
	PriceInput      float64 `json:"priceInput"`  // USD per 1M tokens
	PriceOutput     float64 `json:"priceOutput"` // USD per 1M tokens
}

// Registry is a map of model ID to Spec.
var Registry = map[string]Spec{
	// === Claude (Anthropic) ===
	"claude-fable-5":    {ID: "claude-fable-5", Provider: "claude", ContextWindow: 200000, MaxOutput: 32000, SupportsTools: true, SupportsReasoning: true, PriceInput: 15, PriceOutput: 75},
	"claude-opus-4-8":   {ID: "claude-opus-4-8", Provider: "claude", ContextWindow: 200000, MaxOutput: 32000, SupportsTools: true, SupportsReasoning: true, PriceInput: 15, PriceOutput: 75},
	"claude-sonnet-4-6": {ID: "claude-sonnet-4-6", Provider: "claude", ContextWindow: 200000, MaxOutput: 64000, SupportsTools: true, SupportsVision: true, SupportsReasoning: true, PriceInput: 3, PriceOutput: 15},
	"claude-haiku-4-5":  {ID: "claude-haiku-4-5", Provider: "claude", ContextWindow: 200000, MaxOutput: 8192, SupportsTools: true, SupportsVision: true, PriceInput: 0.8, PriceOutput: 4},

	// === DeepSeek ===
	"deepseek-v4":       {ID: "deepseek-v4", Provider: "deepseek", ContextWindow: 128000, MaxOutput: 65536, SupportsTools: true, SupportsReasoning: true, PriceInput: 0.27, PriceOutput: 1.1},
	"deepseek-v4-pro":   {ID: "deepseek-v4-pro", Provider: "deepseek", ContextWindow: 128000, MaxOutput: 65536, SupportsTools: true, SupportsReasoning: true, PriceInput: 0.55, PriceOutput: 2.19},
	"deepseek-v4-flash": {ID: "deepseek-v4-flash", Provider: "deepseek", ContextWindow: 128000, MaxOutput: 65536, SupportsTools: true, PriceInput: 0.10, PriceOutput: 0.40},

	// === OpenAI ===
	"gpt-5.5":     {ID: "gpt-5.5", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsReasoning: true, PriceInput: 15, PriceOutput: 60},
	"gpt-5.4-mini": {ID: "gpt-5.4-mini", Provider: "openai", ContextWindow: 200000, MaxOutput: 32000, SupportsTools: true, PriceInput: 1, PriceOutput: 4},
	"gpt-o3":       {ID: "gpt-o3", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsReasoning: true, PriceInput: 10, PriceOutput: 40},
	"gpt-o3-mini":  {ID: "gpt-o3-mini", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsReasoning: true, PriceInput: 1.10, PriceOutput: 4.40},
	"gpt-4o":       {ID: "gpt-4o", Provider: "openai", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true, SupportsVision: true, PriceInput: 2.50, PriceOutput: 10},

	// === Gemini (Google) ===
	"gemini-2.5-pro":   {ID: "gemini-2.5-pro", Provider: "gemini", ContextWindow: 1000000, MaxOutput: 65536, SupportsTools: true, SupportsVision: true, SupportsReasoning: true, PriceInput: 1.25, PriceOutput: 10},
	"gemini-2.5-flash": {ID: "gemini-2.5-flash", Provider: "gemini", ContextWindow: 1000000, MaxOutput: 65536, SupportsTools: true, SupportsVision: true, PriceInput: 0.15, PriceOutput: 0.60},
	"gemini-3.5-flash": {ID: "gemini-3.5-flash", Provider: "gemini", ContextWindow: 1000000, MaxOutput: 65536, SupportsTools: true, SupportsVision: true, PriceInput: 0.075, PriceOutput: 0.30},

	// === MiMo (Xiaomi) ===
	"mimo-v2.5":     {ID: "mimo-v2.5", Provider: "mimo", ContextWindow: 128000, MaxOutput: 128000, SupportsTools: true, SupportsVision: true},
	"mimo-v2.5-pro": {ID: "mimo-v2.5-pro", Provider: "mimo", ContextWindow: 128000, MaxOutput: 128000, SupportsTools: true, SupportsVision: true, SupportsReasoning: true},

	// === GLM (Zhipu) ===
	"glm-5.1": {ID: "glm-5.1", Provider: "glm", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true},

	// === Kimi (Moonshot) ===
	"kimi-k2.7-code": {ID: "kimi-k2.7-code", Provider: "moonshot", ContextWindow: 128000, MaxOutput: 65536, SupportsTools: true, SupportsReasoning: true},

	// === StepFun ===
	"step-3.5-flash-2603": {ID: "step-3.5-flash-2603", Provider: "stepfun", ContextWindow: 128000, MaxOutput: 8192, SupportsTools: true},

	// === MiniMax ===
	"minimax-m2.7": {ID: "minimax-m2.7", Provider: "minimax", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true},

	// === Baidu (Qianfan) ===
	"qianfan-code-latest": {ID: "qianfan-code-latest", Provider: "baidu", ContextWindow: 128000, MaxOutput: 8192, SupportsTools: true},

	// === Volcengine (Doubao) ===
	"ark-code-latest":       {ID: "ark-code-latest", Provider: "volcengine", ContextWindow: 128000, MaxOutput: 8192, SupportsTools: true},
	"doubao-seed-2-0-code": {ID: "doubao-seed-2-0-code", Provider: "volcengine", ContextWindow: 128000, MaxOutput: 8192, SupportsTools: true},

	// === BaiLing ===
	"ling-2.5-1t": {ID: "ling-2.5-1t", Provider: "bailing", ContextWindow: 128000, MaxOutput: 8192, SupportsTools: true},

	// === Ollama (Local) ===
	"llama3.1":  {ID: "llama3.1", Provider: "ollama", ContextWindow: 128000, MaxOutput: 4096, SupportsTools: true},
	"qwen2.5":   {ID: "qwen2.5", Provider: "ollama", ContextWindow: 128000, MaxOutput: 4096, SupportsTools: true},
	"deepseek-v3": {ID: "deepseek-v3", Provider: "ollama", ContextWindow: 128000, MaxOutput: 4096, SupportsTools: true},
}

// GetSpec returns the model spec for a given model ID, or a default spec if not found.
func GetSpec(modelID string) Spec {
	if spec, ok := Registry[modelID]; ok {
		return spec
	}
	// Default: assume 128K context, tools support
	return Spec{
		ID:            modelID,
		ContextWindow: 128000,
		MaxOutput:     4096,
		SupportsTools: true,
	}
}

// DefaultModels returns default models per provider.
func DefaultModels() map[string]string {
	return map[string]string{
		"claude":     "claude-sonnet-4-6",
		"openai":     "gpt-4o",
		"deepseek":   "deepseek-v4",
		"gemini":     "gemini-2.5-flash",
		"mimo":       "mimo-v2.5-pro",
		"qwen":       "qwen-max",
		"glm":        "glm-5.1",
		"moonshot":   "kimi-k2.7-code",
		"stepfun":    "step-3.5-flash-2603",
		"minimax":    "minimax-m2.7",
		"baidu":      "qianfan-code-latest",
		"volcengine": "ark-code-latest",
		"bailing":    "ling-2.5-1t",
		"ollama":     "llama3.1",
		"custom":     "",
	}
}

// CompactThreshold returns the token count at which to trigger compaction.
func CompactThreshold(ctxWindow int) int {
	switch {
	case ctxWindow >= 500_000:
		return 400_000 // 1M models
	case ctxWindow >= 100_000:
		return 80_000 // 128K/200K models
	default:
		return 40_000 // Small context models
	}
}
