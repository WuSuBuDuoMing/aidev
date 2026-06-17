// Package types defines all core type definitions for NeoCode.
package types

import "time"

// ProviderID identifies an AI provider.
type ProviderID string

const (
	ProviderClaude      ProviderID = "claude"
	ProviderOpenAI      ProviderID = "openai"
	ProviderDeepSeek    ProviderID = "deepseek"
	ProviderGemini      ProviderID = "gemini"
	ProviderMiMo        ProviderID = "mimo"
	ProviderQwen        ProviderID = "qwen"
	ProviderGLM         ProviderID = "glm"
	ProviderMoonshot    ProviderID = "moonshot"
	ProviderStepFun     ProviderID = "stepfun"
	ProviderMiniMax     ProviderID = "minimax"
	ProviderBaidu       ProviderID = "baidu"
	ProviderVolcengine  ProviderID = "volcengine"
	ProviderBaiLing     ProviderID = "bailing"
	ProviderOllama      ProviderID = "ollama"
	ProviderCustom      ProviderID = "custom"
)

// MessageRole represents the role of a conversation message.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// ToolCall represents a single tool call embedded in an assistant message.
type ToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role       MessageRole `json:"role"`
	Content    string      `json:"content"`
	Timestamp  time.Time   `json:"timestamp"`
	ToolCalls  []ToolCall  `json:"toolCalls,omitempty"`
	ToolCallID string      `json:"toolCallId,omitempty"`
	ToolName   string      `json:"toolName,omitempty"`
}

// Conversation represents a full conversation with metadata.
type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
}

// ToolDefinition describes a tool for AI providers.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolResult is returned by a tool execution.
type ToolResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// TokenUsage holds token count statistics.
type TokenUsage struct {
	PromptTokens     int     `json:"promptTokens"`
	CompletionTokens int     `json:"completionTokens"`
	TotalTokens      int     `json:"totalTokens"`
	EstimatedCost    float64 `json:"estimatedCost"`
}

// GenerateResult is the unified response from AI providers.
type GenerateResult struct {
	Content        string      `json:"content"`
	Thinking       string      `json:"thinking,omitempty"`
	ToolCalls      []ToolCall  `json:"toolCalls,omitempty"`
	Usage          TokenUsage  `json:"usage"`
	StopReason     string      `json:"stopReason,omitempty"`
}

// StreamChunk is an incremental streaming chunk.
type StreamChunk struct {
	Content  string      `json:"content"`
	Thinking string      `json:"thinking,omitempty"`
	Done     bool        `json:"done"`
	Usage    *TokenUsage `json:"usage,omitempty"`
}

// PermissionMode defines tool access modes.
type PermissionMode string

const (
	PermAllow PermissionMode = "allow"
	PermDeny  PermissionMode = "deny"
	PermAsk   PermissionMode = "ask"
)

// PermissionEntry maps a tool to a permission mode.
type PermissionEntry struct {
	ToolName string         `json:"toolName"`
	Pattern  string         `json:"pattern"`
	Mode     PermissionMode `json:"mode"`
}

// EffortLevel controls reasoning depth.
type EffortLevel int

const (
	EffortLow       EffortLevel = 1
	EffortMedium    EffortLevel = 2
	EffortHigh      EffortLevel = 3
	EffortXHigh     EffortLevel = 4
	EffortUltracode EffortLevel = 5
)

// PermissionModeName is the name of a permission mode for configuration.
type PermissionModeName string

const (
	PermModeAsk  PermissionModeName = "ask"
	PermModeAuto PermissionModeName = "auto"
	PermModePlan PermissionModeName = "plan"
	PermModeEdit PermissionModeName = "edit"
)
