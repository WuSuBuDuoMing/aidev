package types

import (
	"testing"
	"time"
)

func TestProviderConstants(t *testing.T) {
	providers := []ProviderID{
		ProviderClaude, ProviderOpenAI, ProviderDeepSeek, ProviderGemini,
		ProviderMiMo, ProviderQwen, ProviderGLM, ProviderMoonshot,
		ProviderStepFun, ProviderMiniMax, ProviderBaidu, ProviderVolcengine,
		ProviderBaiLing, ProviderOllama, ProviderCustom,
	}

	seen := make(map[ProviderID]bool)
	for _, p := range providers {
		if p == "" {
			t.Error("Provider constant should not be empty")
		}
		if seen[p] {
			t.Errorf("Duplicate provider constant: %s", p)
		}
		seen[p] = true
	}
}

func TestMessageRoles(t *testing.T) {
	roles := []MessageRole{RoleSystem, RoleUser, RoleAssistant, RoleTool}
	expected := []string{"system", "user", "assistant", "tool"}

	for i, role := range roles {
		if string(role) != expected[i] {
			t.Errorf("Role %d = %q, want %q", i, string(role), expected[i])
		}
	}
}

func TestToolCall(t *testing.T) {
	tc := ToolCall{
		ID:        "call-1",
		Name:      "readFile",
		Arguments: map[string]interface{}{"path": "/test"},
	}

	if tc.ID != "call-1" {
		t.Errorf("ID = %q, want %q", tc.ID, "call-1")
	}
	if tc.Name != "readFile" {
		t.Errorf("Name = %q, want %q", tc.Name, "readFile")
	}
	if tc.Arguments["path"] != "/test" {
		t.Errorf("Arguments[path] = %v, want %v", tc.Arguments["path"], "/test")
	}
}

func TestMessage(t *testing.T) {
	now := time.Now()
	msg := Message{
		Role:      RoleUser,
		Content:   "hello",
		Timestamp: now,
	}

	if msg.Role != RoleUser {
		t.Errorf("Role = %q, want %q", msg.Role, RoleUser)
	}
	if msg.Content != "hello" {
		t.Errorf("Content = %q, want %q", msg.Content, "hello")
	}
	if !msg.Timestamp.Equal(now) {
		t.Error("Timestamp should match")
	}
}

func TestConversation(t *testing.T) {
	conv := Conversation{
		ID:        "conv-1",
		Title:     "Test",
		Messages:  []Message{},
		CreatedAt: time.Now(),
		Model:     "deepseek-v4",
		Provider:  "deepseek",
	}

	if conv.ID != "conv-1" {
		t.Errorf("ID = %q, want %q", conv.ID, "conv-1")
	}
	if len(conv.Messages) != 0 {
		t.Errorf("Messages count = %d, want 0", len(conv.Messages))
	}
}

func TestToolDefinition(t *testing.T) {
	td := ToolDefinition{
		Name:        "readFile",
		Description: "Read a file",
		Parameters:  map[string]interface{}{"type": "object"},
	}

	if td.Name != "readFile" {
		t.Errorf("Name = %q, want %q", td.Name, "readFile")
	}
}

func TestToolResult(t *testing.T) {
	success := ToolResult{Success: true, Output: "ok"}
	if !success.Success {
		t.Error("Success should be true")
	}

	fail := ToolResult{Success: false, Error: "failed"}
	if fail.Success {
		t.Error("Success should be false")
	}
	if fail.Error != "failed" {
		t.Errorf("Error = %q, want %q", fail.Error, "failed")
	}
}

func TestTokenUsage(t *testing.T) {
	usage := TokenUsage{
		PromptTokens:     100,
		CompletionTokens: 200,
		TotalTokens:      300,
		EstimatedCost:    0.01,
	}

	if usage.PromptTokens != 100 {
		t.Errorf("PromptTokens = %d, want 100", usage.PromptTokens)
	}
	if usage.CompletionTokens != 200 {
		t.Errorf("CompletionTokens = %d, want 200", usage.CompletionTokens)
	}
	if usage.TotalTokens != 300 {
		t.Errorf("TotalTokens = %d, want 300", usage.TotalTokens)
	}
}

func TestGenerateResult(t *testing.T) {
	result := GenerateResult{
		Content:    "Hello, world!",
		Thinking:   "Let me think...",
		ToolCalls:  []ToolCall{{ID: "1", Name: "test"}},
		StopReason: "end_turn",
	}

	if result.Content != "Hello, world!" {
		t.Errorf("Content = %q", result.Content)
	}
	if len(result.ToolCalls) != 1 {
		t.Errorf("ToolCalls count = %d, want 1", len(result.ToolCalls))
	}
}

func TestPermissionModes(t *testing.T) {
	modes := []PermissionModeName{PermModeAsk, PermModeAuto, PermModePlan, PermModeEdit}
	expected := []string{"ask", "auto", "plan", "edit"}

	for i, mode := range modes {
		if string(mode) != expected[i] {
			t.Errorf("Mode %d = %q, want %q", i, string(mode), expected[i])
		}
	}
}

func TestEffortLevels(t *testing.T) {
	levels := []EffortLevel{EffortLow, EffortMedium, EffortHigh, EffortXHigh, EffortUltracode}
	for i, level := range levels {
		if int(level) != i+1 {
			t.Errorf("EffortLevel[%d] = %d, want %d", i, int(level), i+1)
		}
	}
}
