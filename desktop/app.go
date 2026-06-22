package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/agent"
	"github.com/WuSuBuDuoMing/aidev/internal/config"
	"github.com/WuSuBuDuoMing/aidev/internal/config/model"
	"github.com/WuSuBuDuoMing/aidev/internal/provider"
	"github.com/WuSuBuDuoMing/aidev/internal/session"
	"github.com/WuSuBuDuoMing/aidev/internal/storage"
	"github.com/WuSuBuDuoMing/aidev/internal/tool"
	"github.com/WuSuBuDuoMing/aidev/internal/tool/builtin"
	"github.com/WuSuBuDuoMing/aidev/internal/types"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main Wails application struct.
type App struct {
	ctx        context.Context
	cfg        *config.Config
	prov       provider.Provider
	modelName  string
	tools      *tool.Registry
	db         *storage.DB
	sessMgr    *session.Manager
	conv       *types.Conversation
	cancelFunc context.CancelFunc
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load config
	cwd, _ := os.Getwd()
	cfg, err := config.Load(cwd)
	if err != nil {
		cfg = config.DefaultConfig()
	}
	a.cfg = cfg

	// Create provider
	prov, modelName, err := provider.CreateProvider(cfg)
	if err != nil {
		fmt.Printf("Provider error: %v\n", err)
	}
	a.prov = prov
	a.modelName = modelName

	// Create tools
	a.tools = builtin.NewDefaultRegistry()

	// Create session manager
	db, err := storage.Open(config.DBPath())
	if err == nil {
		a.db = db
		a.sessMgr = session.NewManager(db)
	}

	// Create conversation
	a.conv = &types.Conversation{
		ID:        fmt.Sprintf("conv-%d", time.Now().UnixMilli()),
		Title:     "New Conversation",
		Messages:  []types.Message{},
		CreatedAt: time.Now(),
		Model:     modelName,
		Provider:  cfg.DefaultProvider,
	}
}

// shutdown is called when the app shuts down.
func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
}

// GetConfig returns the current configuration.
func (a *App) GetConfig() map[string]interface{} {
	spec := model.GetSpec(a.modelName)
	return map[string]interface{}{
		"provider":    a.cfg.DefaultProvider,
		"model":       a.modelName,
		"apiKey":      maskKey(a.cfg.APIKey),
		"baseUrl":     a.cfg.BaseURL,
		"temperature": a.cfg.Temperature,
		"effort":      a.cfg.Effort,
		"mode":        a.cfg.PermissionMode,
		"theme":       a.cfg.Theme,
		"language":    a.cfg.Language,
		"contextWindow": spec.ContextWindow,
		"maxOutput":     spec.MaxOutput,
		"toolsCount":    len(a.tools.GetAll()),
	}
}

// GetStatus returns the full status information.
func (a *App) GetStatus() map[string]interface{} {
	spec := model.GetSpec(a.modelName)
	var sessions []session.SessionSummary
	if a.sessMgr != nil {
		sessions, _ = a.sessMgr.List(10)
	}
	return map[string]interface{}{
		"version":        "2.9.0",
		"provider":       a.cfg.DefaultProvider,
		"model":          a.modelName,
		"contextWindow":  spec.ContextWindow,
		"toolsCount":     len(a.tools.GetAll()),
		"sessionsCount":  len(sessions),
		"configDir":      config.ConfigDir(),
	}
}

// SendMessage sends a message to the AI and returns the response via events.
func (a *App) SendMessage(content string) map[string]interface{} {
	if a.prov == nil {
		return map[string]interface{}{"error": "Provider not configured"}
	}

	// Add user message
	a.conv.Messages = append(a.conv.Messages, types.Message{
		Role:      types.RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	})

	// Auto-title
	if a.conv.Title == "New Conversation" {
		if len(content) > 60 {
			a.conv.Title = content[:57] + "..."
		} else {
			a.conv.Title = content
		}
	}

	// Prune and compact
	a.conv.Messages = agent.PruneToolResults(a.conv.Messages)
	spec := model.GetSpec(a.modelName)
	a.conv.Messages, _ = agent.CompactIfNeeded(a.conv.Messages, model.CompactThreshold(spec.ContextWindow))

	// Create cancellable context
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelFunc = cancel

	systemPrompt := buildSystemPrompt()

	result, finalMsgs, err := agent.Run(ctx, agent.Options{
		Provider:       a.prov,
		Tools:          a.tools,
		SystemPrompt:   systemPrompt,
		ModelID:        a.modelName,
		PermissionMode: types.PermissionModeName(a.cfg.PermissionMode),
		Stream:         true,
		OnToken: func(chunk types.StreamChunk) {
			if chunk.Content != "" {
				wailsRuntime.EventsEmit(a.ctx, "stream:token", chunk.Content)
			}
		},
		OnToolStart: func(name string, args map[string]interface{}) {
			wailsRuntime.EventsEmit(a.ctx, "tool:start", map[string]interface{}{
				"name": name,
				"args": args,
			})
		},
		OnToolEnd: func(name string, tr *types.ToolResult) {
			wailsRuntime.EventsEmit(a.ctx, "tool:end", map[string]interface{}{
				"name":    name,
				"success": tr.Success,
				"output":  tr.Output,
				"error":   tr.Error,
			})
		},
	}, a.conv.Messages)

	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	a.conv.Messages = finalMsgs

	// Auto-save
	if a.sessMgr != nil {
		a.sessMgr.Save(a.conv)
	}

	// Calculate cost
	cost := float64(result.Usage.PromptTokens)*spec.PriceInput/1_000_000 +
		float64(result.Usage.CompletionTokens)*spec.PriceOutput/1_000_000

	return map[string]interface{}{
		"content":    result.Content,
		"thinking":   result.Thinking,
		"toolCalls":  len(result.ToolCalls),
		"tokens":     result.Usage.TotalTokens,
		"cost":       cost,
		"session":    a.conv.ID,
		"title":      a.conv.Title,
	}
}

// ListSessions returns recent sessions.
func (a *App) ListSessions() []map[string]interface{} {
	if a.sessMgr == nil {
		return nil
	}
	sessions, _ := a.sessMgr.List(20)
	var result []map[string]interface{}
	for _, s := range sessions {
		result = append(result, map[string]interface{}{
			"id":            s.ID,
			"title":         s.Title,
			"messageCount":  s.MessageCount,
			"updatedAt":     s.UpdatedAt,
		})
	}
	return result
}

// ResumeSession resumes a past session.
func (a *App) ResumeSession(sessionID string) map[string]interface{} {
	if a.sessMgr == nil {
		return map[string]interface{}{"error": "Session manager not available"}
	}
	conv, err := a.sessMgr.Load(sessionID)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	a.conv = conv
	return map[string]interface{}{
		"id":       conv.ID,
		"title":    conv.Title,
		"messages": len(conv.Messages),
	}
}

// NewConversation starts a new conversation.
func (a *App) NewConversation() {
	a.conv = &types.Conversation{
		ID:        fmt.Sprintf("conv-%d", time.Now().UnixMilli()),
		Title:     "New Conversation",
		Messages:  []types.Message{},
		CreatedAt: time.Now(),
		Model:     a.modelName,
		Provider:  a.cfg.DefaultProvider,
	}
}

// UpdateConfig updates the configuration.
func (a *App) UpdateConfig(patch map[string]interface{}) map[string]interface{} {
	if v, ok := patch["provider"].(string); ok {
		a.cfg.DefaultProvider = v
	}
	if v, ok := patch["model"].(string); ok {
		a.cfg.DefaultModel = v
		a.modelName = v
	}
	if v, ok := patch["apiKey"].(string); ok {
		a.cfg.APIKey = v
	}
	if v, ok := patch["effort"].(string); ok {
		a.cfg.Effort = v
	}
	if v, ok := patch["mode"].(string); ok {
		a.cfg.PermissionMode = v
	}

	// Recreate provider with new config
	if a.prov != nil {
		prov, modelName, err := provider.CreateProvider(a.cfg)
		if err == nil {
			a.prov = prov
			a.modelName = modelName
		}
	}

	return a.GetConfig()
}

// GetProviders returns the list of available providers.
func (a *App) GetProviders() []map[string]interface{} {
	providers := []struct{ Name, URL, Default string }{
		{"deepseek", "https://api.deepseek.com", "deepseek-v4"},
		{"claude", "https://api.anthropic.com", "claude-sonnet-4-6"},
		{"openai", "https://api.openai.com", "gpt-4o"},
		{"gemini", "https://generativelanguage.googleapis.com", "gemini-2.5-flash"},
		{"mimo", "https://api.xiaomimimo.com/v1", "mimo-v2.5-pro"},
		{"glm", "https://open.bigmodel.cn/api/paas/v4", "glm-5.1"},
		{"moonshot", "https://api.moonshot.cn/v1", "kimi-k2.7-code"},
		{"ollama", "http://localhost:11434", "llama3.1"},
	}
	var result []map[string]interface{}
	for _, p := range providers {
		result = append(result, map[string]interface{}{
			"name":    p.Name,
			"url":     p.URL,
			"default": p.Default,
		})
	}
	return result
}

// GetConversationMessages returns the current conversation messages.
func (a *App) GetConversationMessages() []map[string]interface{} {
	var msgs []map[string]interface{}
	for _, m := range a.conv.Messages {
		if m.Role == types.RoleSystem {
			continue
		}
		msgs = append(msgs, map[string]interface{}{
			"role":      string(m.Role),
			"content":   m.Content,
			"timestamp": m.Timestamp.Format(time.RFC3339),
		})
	}
	return msgs
}

// GetContextUsage returns current context window usage.
func (a *App) GetContextUsage() map[string]interface{} {
	spec := model.GetSpec(a.modelName)
	totalTokens := agent.EstimateConversationTokens(a.conv.Messages)
	threshold := model.CompactThreshold(spec.ContextWindow)
	percentage := float64(totalTokens) / float64(spec.ContextWindow) * 100

	return map[string]interface{}{
		"currentTokens":  totalTokens,
		"contextWindow":  spec.ContextWindow,
		"threshold":      threshold,
		"percentage":     fmt.Sprintf("%.1f%%", percentage),
		"compactAt":      fmt.Sprintf("%.0f%%", float64(threshold)/float64(spec.ContextWindow)*100),
	}
}

// SelectProvider opens a provider selection dialog.
func (a *App) SelectProvider() string {
	return a.cfg.DefaultProvider
}

func maskKey(key string) string {
	if key == "" {
		return "not set"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func buildSystemPrompt() string {
	cwd, _ := os.Getwd()
	return fmt.Sprintf(`You are NeoCode — an expert AI coding agent running in the terminal.

## Core Principles
- Be concise, accurate, and direct.
- Prefer showing code over describing it.
- When modifying files, use the provided tools rather than printing code blocks.
- Always explain what you are about to do before using a tool.
- Respect the user's existing code style and architecture.

## Working Directory
%s

## Output Style
- Use markdown for structured responses.
- Use fenced code blocks with language tags.
- Keep explanations brief — the user is a developer.`, cwd)
}

// generateID creates a unique ID.
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
