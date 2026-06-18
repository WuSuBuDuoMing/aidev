package model

import (
	"testing"
)

func TestGetSpec_Existing(t *testing.T) {
	spec := GetSpec("deepseek-v4")
	if spec.ID != "deepseek-v4" {
		t.Errorf("ID = %q, want %q", spec.ID, "deepseek-v4")
	}
	if spec.Provider != "deepseek" {
		t.Errorf("Provider = %q, want %q", spec.Provider, "deepseek")
	}
	if spec.ContextWindow != 128000 {
		t.Errorf("ContextWindow = %d, want 128000", spec.ContextWindow)
	}
	if !spec.SupportsTools {
		t.Error("deepseek-v4 should support tools")
	}
}

func TestGetSpec_Unknown(t *testing.T) {
	spec := GetSpec("unknown-model-xyz")
	if spec.ContextWindow != 128000 {
		t.Errorf("Unknown model should default to 128K context, got %d", spec.ContextWindow)
	}
	if !spec.SupportsTools {
		t.Error("Unknown model should default to SupportsTools=true")
	}
}

func TestGetSpec_AllModelsExist(t *testing.T) {
	expectedModels := []string{
		"claude-fable-5", "claude-opus-4-8", "claude-sonnet-4-6", "claude-haiku-4-5",
		"deepseek-v4", "deepseek-v4-pro", "deepseek-v4-flash",
		"gpt-5.5", "gpt-5.4-mini", "gpt-o3", "gpt-o3-mini", "gpt-4o",
		"gemini-2.5-pro", "gemini-2.5-flash", "gemini-3.5-flash",
		"mimo-v2.5", "mimo-v2.5-pro",
		"glm-5.1",
		"kimi-k2.7-code",
		"step-3.5-flash-2603",
		"minimax-m2.7",
		"qianfan-code-latest",
		"ark-code-latest", "doubao-seed-2-0-code",
		"ling-2.5-1t",
		"llama3.1", "qwen2.5", "deepseek-v3",
	}

	for _, id := range expectedModels {
		spec := GetSpec(id)
		if spec.ID != id {
			t.Errorf("Model %q not found in registry", id)
		}
	}
}

func TestDefaultModels(t *testing.T) {
	models := DefaultModels()

	expectedProviders := []string{
		"claude", "openai", "deepseek", "gemini", "mimo",
		"glm", "moonshot", "stepfun", "minimax", "baidu",
		"volcengine", "bailing", "ollama", "custom",
	}

	for _, p := range expectedProviders {
		if _, ok := models[p]; !ok {
			t.Errorf("DefaultModels missing provider %q", p)
		}
	}
}

func TestCompactThreshold(t *testing.T) {
	tests := []struct {
		ctxWindow int
		want      int
	}{
		{1000000, 400000},
		{500000, 400000},
		{200000, 80000},
		{128000, 80000},
		{100000, 80000},
		{64000, 40000},
		{32000, 40000},
	}

	for _, tt := range tests {
		got := CompactThreshold(tt.ctxWindow)
		if got != tt.want {
			t.Errorf("CompactThreshold(%d) = %d, want %d", tt.ctxWindow, got, tt.want)
		}
	}
}

func TestRegistryHasCorrectPricing(t *testing.T) {
	// DeepSeek v4 should be cheap
	ds := GetSpec("deepseek-v4")
	if ds.PriceInput >= 1.0 {
		t.Errorf("deepseek-v4 PriceInput = %f, should be < 1.0", ds.PriceInput)
	}

	// Gemini should support 1M context
	gemini := GetSpec("gemini-2.5-pro")
	if gemini.ContextWindow < 1000000 {
		t.Errorf("gemini-2.5-pro ContextWindow = %d, should be >= 1M", gemini.ContextWindow)
	}

	// MiMo should be free
	mimo := GetSpec("mimo-v2.5")
	if mimo.PriceInput != 0 || mimo.PriceOutput != 0 {
		t.Errorf("mimo-v2.5 should be free, got input=%f output=%f", mimo.PriceInput, mimo.PriceOutput)
	}
}

func TestGeminiVisionSupport(t *testing.T) {
	for _, id := range []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-3.5-flash"} {
		spec := GetSpec(id)
		if !spec.SupportsVision {
			t.Errorf("%s should support vision", id)
		}
	}
}
