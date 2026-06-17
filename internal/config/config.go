// Package config handles configuration loading with 4-layer merging.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config is the application configuration.
type Config struct {
	DefaultProvider string            `toml:"default_provider"`
	DefaultModel    string            `toml:"default_model"`
	APIKey          string            `toml:"api_key"`
	BaseURL         string            `toml:"base_url"`
	Temperature     float64           `toml:"temperature"`
	MaxTokens       int               `toml:"max_tokens"`
	Theme           string            `toml:"theme"`
	Stream          bool              `toml:"stream"`
	ThinkingEnabled bool              `toml:"thinking_enabled"`
	Language        string            `toml:"language"`
	PermissionMode  string            `toml:"permission_mode"` // ask, auto, plan, edit
	Effort          string            `toml:"effort"`          // low, medium, high, xhigh, ultracode
	ShowTokenCounts bool              `toml:"show_token_counts"`
	Providers       []ProviderConfig  `toml:"providers"`
	Permissions     []PermissionEntry `toml:"permissions"`
	Sandbox         SandboxConfig     `toml:"sandbox"`
	Update          UpdateConfig      `toml:"update"`
}

// ProviderConfig defines a provider.
type ProviderConfig struct {
	Name     string `toml:"name"`
	Kind     string `toml:"kind"` // anthropic, openai, gemini, custom
	BaseURL  string `toml:"base_url"`
	APIKeyEnv string `toml:"api_key_env"`
	Model    string `toml:"model"`
	ContextWindow int  `toml:"context_window"`
	MaxOutput int    `toml:"max_output"`
}

// PermissionEntry defines a tool permission rule.
type PermissionEntry struct {
	ToolName string `toml:"tool_name"`
	Mode     string `toml:"mode"` // allow, deny, ask
}

// SandboxConfig defines sandbox settings.
type SandboxConfig struct {
	WriteRoots []string `toml:"write_roots"`
}

// UpdateConfig defines auto-update settings.
type UpdateConfig struct {
	AutoCheck bool   `toml:"auto_check"`
	Channel   string `toml:"channel"` // stable, canary
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "deepseek",
		DefaultModel:    "deepseek-v4",
		Temperature:     0.2,
		MaxTokens:       4096,
		Theme:           "dark",
		Stream:          true,
		ThinkingEnabled: false,
		Language:        "zh-CN",
		PermissionMode:  "ask",
		Effort:          "medium",
		ShowTokenCounts: true,
		Sandbox: SandboxConfig{
			WriteRoots: []string{"."},
		},
		Update: UpdateConfig{
			AutoCheck: true,
			Channel:   "stable",
		},
	}
}

// ConfigDir returns the NeoCode config directory (~/.neocode).
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".neocode")
}

// ConfigPath returns the path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// ProjectConfigPath returns the project-level config path.
func ProjectConfigPath(cwd string) string {
	return filepath.Join(cwd, ".neocode", "config.toml")
}

// Load loads configuration from all 4 layers.
func Load(cwd string) (*Config, error) {
	cfg := DefaultConfig()

	// Layer 2: User config (~/.neocode/config.toml)
	userCfg, err := loadTomlFile(ConfigPath())
	if err == nil {
		mergeConfig(cfg, userCfg)
	}

	// Layer 3: Project config (<project>/.neocode/config.toml)
	if cwd != "" {
		projCfg, err := loadTomlFile(ProjectConfigPath(cwd))
		if err == nil {
			mergeConfig(cfg, projCfg)
		}
	}

	// Layer 4: Environment variables (highest priority)
	applyEnvOverrides(cfg)

	return cfg, nil
}

// Save saves the configuration to the user config file.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(ConfigPath(), data, 0o644)
}

// loadTomlFile reads and parses a TOML file.
func loadTomlFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

// mergeConfig merges src into dst (src wins on conflict).
func mergeConfig(dst, src *Config) {
	if src.DefaultProvider != "" {
		dst.DefaultProvider = src.DefaultProvider
	}
	if src.DefaultModel != "" {
		dst.DefaultModel = src.DefaultModel
	}
	if src.APIKey != "" {
		dst.APIKey = src.APIKey
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.Temperature != 0 {
		dst.Temperature = src.Temperature
	}
	if src.MaxTokens != 0 {
		dst.MaxTokens = src.MaxTokens
	}
	if src.Theme != "" {
		dst.Theme = src.Theme
	}
	if src.Language != "" {
		dst.Language = src.Language
	}
	if src.PermissionMode != "" {
		dst.PermissionMode = src.PermissionMode
	}
	if src.Effort != "" {
		dst.Effort = src.Effort
	}
	// Stream and ThinkingEnabled are bools - only override if explicitly set
	if len(src.Providers) > 0 {
		dst.Providers = src.Providers
	}
	if len(src.Permissions) > 0 {
		dst.Permissions = src.Permissions
	}
}

// applyEnvOverrides applies environment variable overrides.
func applyEnvOverrides(cfg *Config) {
	// NeoCode-specific (highest priority)
	if v := os.Getenv("NEOCODE_PROVIDER"); v != "" {
		cfg.DefaultProvider = v
	}
	if v := os.Getenv("NEOCODE_MODEL"); v != "" {
		cfg.DefaultModel = v
	}
	if v := os.Getenv("NEOCODE_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("NEOCODE_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}

	// ccSwitch standard variables
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" && cfg.DefaultProvider == "claude" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" && cfg.DefaultProvider == "claude" {
		if cfg.BaseURL == "" {
			cfg.BaseURL = v
		}
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" && cfg.DefaultProvider == "openai" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" && cfg.DefaultProvider == "openai" {
		if cfg.BaseURL == "" {
			cfg.BaseURL = v
		}
	}

	// Provider-specific
	if v := os.Getenv("DEEPSEEK_API_KEY"); v != "" && cfg.DefaultProvider == "deepseek" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("MI_API_KEY"); v != "" && cfg.DefaultProvider == "mimo" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("QWEN_API_KEY"); v != "" && cfg.DefaultProvider == "qwen" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("GLM_API_KEY"); v != "" && cfg.DefaultProvider == "glm" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("MOONSHOT_API_KEY"); v != "" && cfg.DefaultProvider == "moonshot" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}
	if v := os.Getenv("GOOGLE_API_KEY"); v != "" && cfg.DefaultProvider == "gemini" {
		if cfg.APIKey == "" {
			cfg.APIKey = v
		}
	}

	// Theme
	if v := os.Getenv("NEOCODE_THEME"); v != "" {
		cfg.Theme = v
	}
}

// DetectLanguage returns the language name from the language code.
func DetectLanguage(code string) string {
	switch {
	case strings.HasPrefix(code, "zh"):
		return "中文"
	default:
		return "English"
	}
}

// HomeDir returns the home directory.
func HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir := ConfigDir()
	return os.MkdirAll(dir, 0o755)
}

// DataDir returns the data directory (~/.neocode).
func DataDir() string {
	return ConfigDir()
}

// DBPath returns the path to the SQLite database.
func DBPath() string {
	return filepath.Join(DataDir(), "history.db")
}

// KeyStorePath returns the path to the encrypted keystore.
func KeyStorePath() string {
	return filepath.Join(DataDir(), "keystore.enc")
}

// SkillsDir returns the global skills directory.
func SkillsDir() string {
	return filepath.Join(DataDir(), "skills")
}

// PluginsDir returns the global plugins directory.
func PluginsDir() string {
	return filepath.Join(DataDir(), "plugins")
}

// OS returns the current operating system.
func OS() string {
	return runtime.GOOS
}
