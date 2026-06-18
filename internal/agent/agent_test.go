package agent

import (
	"testing"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

func TestNewStormBreaker(t *testing.T) {
	sb := NewStormBreaker()
	if sb == nil {
		t.Fatal("NewStormBreaker returned nil")
	}
	if len(sb.signatures) != 0 {
		t.Fatal("New StormBreaker should have empty signatures")
	}
}

func TestStormBreaker_Record_NoError(t *testing.T) {
	sb := NewStormBreaker()
	breached := sb.Record("readFile", "")
	if breached {
		t.Fatal("Empty error should not breach")
	}
	if len(sb.signatures) != 0 {
		t.Fatal("No error should not record a signature")
	}
}

func TestStormBreaker_Record_BreachThreshold(t *testing.T) {
	sb := NewStormBreaker()

	for i := 0; i < StormBreakThreshold-1; i++ {
		breached := sb.Record("runCommand", "exit status 1")
		if breached {
			t.Fatalf("Should not breach at attempt %d", i+1)
		}
	}

	breached := sb.Record("runCommand", "exit status 1")
	if !breached {
		t.Fatalf("Should breach at attempt %d", StormBreakThreshold)
	}
}

func TestStormBreaker_DifferentErrors(t *testing.T) {
	sb := NewStormBreaker()

	// Different error signatures should not combine
	for i := 0; i < StormBreakThreshold; i++ {
		sb.Record("readFile", "file not found")
		breached := sb.Record("writeFile", "permission denied")
		if breached {
			t.Fatal("Different error signatures should not combine")
		}
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 1},    // (0+3)/4 = 0, but min is 1 due to integer math
		{"hi", 1},  // (2+3)/4 = 1
		{"hello world test", 5}, // (16+3)/4 = 4
	}

	for _, tt := range tests {
		result := EstimateTokens(tt.input)
		if result != tt.expected {
			t.Errorf("EstimateTokens(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestEstimateConversationTokens(t *testing.T) {
	msgs := []types.Message{
		{Role: types.RoleUser, Content: "hello"},
		{Role: types.RoleAssistant, Content: "hi there"},
	}

	tokens := EstimateConversationTokens(msgs)
	if tokens <= 0 {
		t.Fatal("Should estimate positive tokens")
	}
}

func TestCompactIfNeeded_BelowThreshold(t *testing.T) {
	msgs := []types.Message{
		{Role: types.RoleUser, Content: "short message", Timestamp: time.Now()},
	}

	result, compacted := CompactIfNeeded(msgs, 100000)
	if compacted {
		t.Fatal("Short conversation should not trigger compaction")
	}
	if len(result) != len(msgs) {
		t.Fatal("Messages should be unchanged")
	}
}

func TestCompactIfNeeded_AboveThreshold(t *testing.T) {
	// Create a conversation large enough to trigger compaction
	var msgs []types.Message
	for i := 0; i < 20; i++ {
		msgs = append(msgs, types.Message{
			Role:      types.RoleUser,
			Content:   "This is a message with enough content to push the token count above the threshold for compaction testing purposes and needs to be sufficiently long",
			Timestamp: time.Now(),
		})
		msgs = append(msgs, types.Message{
			Role:      types.RoleAssistant,
			Content:   "This is a response with enough content to push the token count above the threshold for compaction testing purposes and needs to be sufficiently long as well",
			Timestamp: time.Now(),
		})
	}

	result, compacted := CompactIfNeeded(msgs, 100)
	if !compacted {
		t.Fatal("Long conversation should trigger compaction")
	}
	if len(result) >= len(msgs) {
		t.Fatal("Compacted conversation should be shorter")
	}
}

func TestPruneToolResults(t *testing.T) {
	msgs := []types.Message{
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
		{Role: types.RoleTool, ToolName: "readFile", Content: string(make([]byte, 1000))},
	}

	result := PruneToolResults(msgs)

	// First 2 messages (index 0-1, since cutoff = max(0, 10-8) = 2) should be pruned
	if len(result) != 10 {
		t.Fatal("Pruning should not change message count")
	}

	// Older messages should be pruned
	for i := 0; i < 2; i++ {
		if !isPruned(result[i].Content) {
			t.Errorf("Message %d should be pruned", i)
		}
	}

	// Recent messages should not be pruned
	for i := 2; i < 10; i++ {
		if isPruned(result[i].Content) {
			t.Errorf("Message %d should not be pruned", i)
		}
	}
}

func isPruned(content string) bool {
	return len(content) < 1000
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3, 5) should be 3")
	}
	if min(5, 3) != 3 {
		t.Error("min(5, 3) should be 3")
	}
	if min(0, 0) != 0 {
		t.Error("min(0, 0) should be 0")
	}
}

func TestMax(t *testing.T) {
	if max(3, 5) != 5 {
		t.Error("max(3, 5) should be 5")
	}
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
}
