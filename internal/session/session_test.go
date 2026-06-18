package session

import (
	"testing"
)

func TestGenerateTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello world", "Hello world"},
		{"", "New Conversation"},
		{string(make([]byte, 100)), string(make([]byte, 57)) + "..."},
	}

	for _, tt := range tests {
		got := GenerateTitle(tt.input)
		if got != tt.want {
			t.Errorf("GenerateTitle(%q) = %q, want %q", truncate(tt.input, 30), got, tt.want)
		}
	}
}

func TestGenerateTitle_MaxLength(t *testing.T) {
	long := string(make([]byte, 200))
	title := GenerateTitle(long)
	if len(title) > 60 {
		t.Errorf("Title too long: %d chars", len(title))
	}
	if len(title) != 60 {
		t.Errorf("Title length = %d, want 60", len(title))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
