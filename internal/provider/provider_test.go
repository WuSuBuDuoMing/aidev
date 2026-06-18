package provider

import (
	"testing"
)

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.BaseDelay <= 0 {
		t.Error("BaseDelay should be positive")
	}
	if cfg.MaxDelay <= 0 {
		t.Error("MaxDelay should be positive")
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		errMsg  string
		want    bool
	}{
		{"API error 429: rate limited", true},
		{"API error 500: internal server error", true},
		{"API error 502: bad gateway", true},
		{"API error 503: service unavailable", true},
		{"API error 504: gateway timeout", true},
		{"API error 529: overloaded", true},
		{"connection ECONNRESET", true},
		{"connection EPIPE broken", true},
		{"ETIMEDOUT error", true},
		{"EAI_AGAIN dns", true},
		{"connection refused", true},
		{"SSE read timed out", true},
		{"API error 400: bad request", false},
		{"API error 401: unauthorized", false},
		{"API error 403: forbidden", false},
		{"API error 404: not found", false},
		{"some random error", false},
	}

	for _, tt := range tests {
		err := &testError{msg: tt.errMsg}
		got := IsRetryable(err)
		if got != tt.want {
			t.Errorf("IsRetryable(%q) = %v, want %v", tt.errMsg, got, tt.want)
		}
	}

	// nil error
	if IsRetryable(nil) {
		t.Error("IsRetryable(nil) should be false")
	}
}

func TestIsRateLimited(t *testing.T) {
	if !IsRateLimited(&testError{"API error 429"}) {
		t.Error("Should detect 429 as rate limited")
	}
	if IsRateLimited(&testError{"API error 500"}) {
		t.Error("500 is not rate limited")
	}
	if IsRateLimited(nil) {
		t.Error("nil is not rate limited")
	}
}

func TestCalculateDelay(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 1000_000_000, // 1 second
		MaxDelay:  60_000_000_000, // 60 seconds
	}

	// First attempt: ~1s
	d := CalculateDelay(0, cfg, 0)
	if d <= 0 {
		t.Error("Delay should be positive")
	}

	// Retry-After should be honored
	d = CalculateDelay(0, cfg, 5_000_000_000) // 5 seconds
	if d != 5_000_000_000 {
		t.Errorf("Should honor Retry-After, got %v", d)
	}

	// Retry-After exceeding max should be capped
	d = CalculateDelay(0, cfg, 999_000_000_000) // 999 seconds
	if d != cfg.MaxDelay {
		t.Error("Should cap Retry-After to MaxDelay")
	}
}

func TestProviderKind(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"claude", "anthropic"},
		{"openai", "openai"},
		{"deepseek", "openai"},
		{"gemini", "gemini"},
		{"mimo", "openai"},
		{"qwen", "openai"},
		{"glm", "openai"},
		{"moonshot", "openai"},
		{"ollama", "openai"},
		{"custom", "openai"},
		{"unknown", "openai"},
	}

	for _, tt := range tests {
		got := ProviderKind(tt.name)
		if got != tt.want {
			t.Errorf("ProviderKind(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDefaultBaseURL(t *testing.T) {
	tests := []struct {
		provider string
		wantNon  bool
	}{
		{"deepseek", true},
		{"claude", true},
		{"openai", true},
		{"gemini", true},
		{"mimo", true},
		{"ollama", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		got := DefaultBaseURL(tt.provider)
		if tt.wantNon && got == "" {
			t.Errorf("DefaultBaseURL(%q) should not be empty", tt.provider)
		}
		if !tt.wantNon && got != "" {
			t.Errorf("DefaultBaseURL(%q) should be empty, got %q", tt.provider, got)
		}
	}
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.GetAll()) != 0 {
		t.Error("New registry should be empty")
	}
}

// testError implements error for testing.
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
