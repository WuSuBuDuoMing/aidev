package permission

import (
	"testing"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

func TestNewPolicy(t *testing.T) {
	p := NewPolicy(types.PermModeAsk, nil, "")
	if p == nil {
		t.Fatal("NewPolicy returned nil")
	}
}

func TestPolicy_AskMode(t *testing.T) {
	p := NewPolicy(types.PermModeAsk, nil, "")

	// Read-only tools should be allowed
	if p.Evaluate("readFile", nil) != Allow {
		t.Error("readFile should be allowed in ask mode")
	}
	if p.Evaluate("grep", nil) != Allow {
		t.Error("grep should be allowed in ask mode")
	}

	// Write tools should ask
	if p.Evaluate("writeFile", nil) != Ask {
		t.Error("writeFile should ask in ask mode")
	}
	if p.Evaluate("runCommand", nil) != Ask {
		t.Error("runCommand should ask in ask mode")
	}
}

func TestPolicy_EditMode(t *testing.T) {
	p := NewPolicy(types.PermModeEdit, nil, "")

	if p.Evaluate("readFile", nil) != Allow {
		t.Error("readFile should be allowed in edit mode")
	}
	if p.Evaluate("writeFile", nil) != Allow {
		t.Error("writeFile should be allowed in edit mode")
	}
	if p.Evaluate("runCommand", nil) != Allow {
		t.Error("runCommand should be allowed in edit mode")
	}
}

func TestPolicy_PlanMode(t *testing.T) {
	p := NewPolicy(types.PermModePlan, nil, "")

	if p.Evaluate("readFile", nil) != Allow {
		t.Error("readFile should be allowed in plan mode")
	}
	if p.Evaluate("writeFile", nil) != Deny {
		t.Error("writeFile should be denied in plan mode")
	}
	if p.Evaluate("runCommand", nil) != Deny {
		t.Error("runCommand should be denied in plan mode")
	}
}

func TestPolicy_AutoMode(t *testing.T) {
	p := NewPolicy(types.PermModeAuto, nil, "")

	if p.Evaluate("readFile", nil) != Allow {
		t.Error("readFile should be allowed in auto mode")
	}
	// Write tools should ask in auto mode
	if p.Evaluate("writeFile", nil) != Ask {
		t.Error("writeFile should ask in auto mode")
	}
}

func TestPolicy_ExplicitRules(t *testing.T) {
	rules := []types.PermissionEntry{
		{ToolName: "runCommand", Mode: types.PermAllow},
		{ToolName: "writeFile", Mode: types.PermDeny},
	}

	p := NewPolicy(types.PermModeAsk, rules, "")

	// Explicit allow overrides mode default
	if p.Evaluate("runCommand", nil) != Allow {
		t.Error("runCommand should be allowed by explicit rule")
	}

	// Explicit deny overrides mode default
	if p.Evaluate("writeFile", nil) != Deny {
		t.Error("writeFile should be denied by explicit rule")
	}
}

func TestPolicy_WildcardRules(t *testing.T) {
	rules := []types.PermissionEntry{
		{ToolName: "git*", Mode: types.PermAllow},
	}

	p := NewPolicy(types.PermModeAsk, rules, "")

	if p.Evaluate("gitStatus", nil) != Allow {
		t.Error("gitStatus should match git* wildcard")
	}
	if p.Evaluate("gitDiff", nil) != Allow {
		t.Error("gitDiff should match git* wildcard")
	}
	if p.Evaluate("gitCommit", nil) != Allow {
		t.Error("gitCommit should match git* wildcard")
	}
	// Non-matching tool should not be affected
	if p.Evaluate("writeFile", nil) != Ask {
		t.Error("writeFile should not match git* wildcard")
	}
}

func TestCheckPathSandbox(t *testing.T) {
	tests := []struct {
		path      string
		root      string
		wantSafe  bool
	}{
		{"./test.txt", ".", true},
		{"subdir/file.txt", ".", true},
		{"", "", true}, // No sandbox
	}

	for _, tt := range tests {
		got := CheckPathSandbox(tt.path, tt.root)
		if got != tt.wantSafe {
			t.Errorf("CheckPathSandbox(%q, %q) = %v, want %v", tt.path, tt.root, got, tt.wantSafe)
		}
	}
}

func TestIsReadOnlyTool(t *testing.T) {
	readOnly := []string{"readFile", "searchFiles", "listDir", "gitStatus", "gitDiff", "glob", "grep"}
	for _, name := range readOnly {
		if !isReadOnlyTool(name) {
			t.Errorf("%s should be read-only", name)
		}
	}

	writeTools := []string{"writeFile", "editFile", "runCommand", "gitCommit"}
	for _, name := range writeTools {
		if isReadOnlyTool(name) {
			t.Errorf("%s should not be read-only", name)
		}
	}
}

func TestMatchRule(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"*", "anything", true},
		{"readFile", "readFile", true},
		{"readFile", "writeFile", false},
		{"git*", "gitStatus", true},
		{"git*", "gitDiff", true},
		{"git*", "readFile", false},
	}

	for _, tt := range tests {
		got := matchRule(tt.pattern, tt.name)
		if got != tt.want {
			t.Errorf("matchRule(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}

func TestDecisionString(t *testing.T) {
	if DecisionString(Allow) != "allow" {
		t.Errorf("Allow = %q", DecisionString(Allow))
	}
	if DecisionString(Deny) != "deny" {
		t.Errorf("Deny = %q", DecisionString(Deny))
	}
	if DecisionString(Ask) != "ask" {
		t.Errorf("Ask = %q", DecisionString(Ask))
	}
}

func TestFormatToolCall(t *testing.T) {
	tests := []struct {
		name string
		args map[string]interface{}
		want string
	}{
		{"readFile", map[string]interface{}{"path": "/test"}, "readFile(/test)"},
		{"writeFile", map[string]interface{}{"path": "/out"}, "writeFile(/out)"},
		{"runCommand", map[string]interface{}{"command": "ls"}, "runCommand(ls)"},
		{"gitStatus", nil, "gitStatus()"},
		{"unknown", nil, "unknown"},
	}

	for _, tt := range tests {
		got := FormatToolCall(tt.name, tt.args)
		if got != tt.want {
			t.Errorf("FormatToolCall(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
