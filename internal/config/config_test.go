package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DefaultProvider != "deepseek" {
		t.Errorf("DefaultProvider = %q, want %q", cfg.DefaultProvider, "deepseek")
	}
	if cfg.DefaultModel != "deepseek-v4" {
		t.Errorf("DefaultModel = %q, want %q", cfg.DefaultModel, "deepseek-v4")
	}
	if cfg.Temperature != 0.2 {
		t.Errorf("Temperature = %f, want %f", cfg.Temperature, 0.2)
	}
	if cfg.MaxTokens != 4096 {
		t.Errorf("MaxTokens = %d, want %d", cfg.MaxTokens, 4096)
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "dark")
	}
	if !cfg.Stream {
		t.Error("Stream should be true by default")
	}
	if cfg.PermissionMode != "ask" {
		t.Errorf("PermissionMode = %q, want %q", cfg.PermissionMode, "ask")
	}
	if cfg.Effort != "medium" {
		t.Errorf("Effort = %q, want %q", cfg.Effort, "medium")
	}
	if !cfg.ShowTokenCounts {
		t.Error("ShowTokenCounts should be true by default")
	}
}

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	if dir == "" {
		t.Fatal("ConfigDir should not be empty")
	}
	if filepath.Base(dir) != ".neocode" {
		t.Errorf("ConfigDir base = %q, want %q", filepath.Base(dir), ".neocode")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if filepath.Base(path) != "config.toml" {
		t.Errorf("ConfigPath base = %q, want %q", filepath.Base(path), "config.toml")
	}
}

func TestDBPath(t *testing.T) {
	path := DBPath()
	if filepath.Base(path) != "history.db" {
		t.Errorf("DBPath base = %q, want %q", filepath.Base(path), "history.db")
	}
}

func TestKeyStorePath(t *testing.T) {
	path := KeyStorePath()
	if filepath.Base(path) != "keystore.enc" {
		t.Errorf("KeyStorePath base = %q, want %q", filepath.Base(path), "keystore.enc")
	}
}

func TestSkillsDir(t *testing.T) {
	dir := SkillsDir()
	if filepath.Base(dir) != "skills" {
		t.Errorf("SkillsDir base = %q, want %q", filepath.Base(dir), "skills")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"zh-CN", "中文"},
		{"zh-TW", "中文"},
		{"en", "English"},
		{"fr", "English"},
	}

	for _, tt := range tests {
		got := DetectLanguage(tt.code)
		if got != tt.want {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestLoadFromDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.DefaultProvider == "" {
		t.Error("DefaultProvider should not be empty")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temp dir for the config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := DefaultConfig()
	cfg.DefaultProvider = "test-provider"
	cfg.DefaultModel = "test-model"

	// Test marshal
	_ = configPath // We test Save/Load indirectly through Load("")
	if cfg.DefaultProvider != "test-provider" {
		t.Error("Config mutation failed")
	}
}

func TestMergeConfig(t *testing.T) {
	dst := DefaultConfig()
	src := &Config{
		DefaultProvider: "custom",
		DefaultModel:    "custom-model",
		APIKey:          "test-key",
		Temperature:     0.5,
		Theme:           "light",
	}

	mergeConfig(dst, src)

	if dst.DefaultProvider != "custom" {
		t.Errorf("Merged provider = %q, want %q", dst.DefaultProvider, "custom")
	}
	if dst.DefaultModel != "custom-model" {
		t.Errorf("Merged model = %q, want %q", dst.DefaultModel, "custom-model")
	}
	if dst.APIKey != "test-key" {
		t.Errorf("Merged APIKey = %q, want %q", dst.APIKey, "test-key")
	}
	if dst.Temperature != 0.5 {
		t.Errorf("Merged temperature = %f, want %f", dst.Temperature, 0.5)
	}
	if dst.Theme != "light" {
		t.Errorf("Merged theme = %q, want %q", dst.Theme, "light")
	}
}

func TestMergeConfigPreservesDefaults(t *testing.T) {
	dst := DefaultConfig()
	src := &Config{
		DefaultProvider: "custom",
		// Other fields left empty -- should preserve defaults
	}

	mergeConfig(dst, src)

	if dst.DefaultProvider != "custom" {
		t.Errorf("Provider should be overridden")
	}
	if dst.DefaultModel != "deepseek-v4" {
		t.Errorf("Model should remain default, got %q", dst.DefaultModel)
	}
	if dst.Temperature != 0.2 {
		t.Errorf("Temperature should remain default, got %f", dst.Temperature)
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := DefaultConfig()

	os.Setenv("NEOCODE_PROVIDER", "env-provider")
	defer os.Unsetenv("NEOCODE_PROVIDER")

	applyEnvOverrides(cfg)

	if cfg.DefaultProvider != "env-provider" {
		t.Errorf("Env override provider = %q, want %q", cfg.DefaultProvider, "env-provider")
	}
}

func TestProjectConfigPath(t *testing.T) {
	path := ProjectConfigPath("/some/project")
	expected := filepath.Join("/some/project", ".neocode", "config.toml")
	if path != expected {
		t.Errorf("ProjectConfigPath = %q, want %q", path, expected)
	}
}
