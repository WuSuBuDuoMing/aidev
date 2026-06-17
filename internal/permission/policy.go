// Package permission implements the permission policy system.
package permission

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// Decision represents a permission decision.
type Decision int

const (
	Deny  Decision = iota
	Allow
	Ask
)

// Policy evaluates tool calls against permission rules.
type Policy struct {
	Mode    types.PermissionModeName
	Rules   []types.PermissionEntry
	Project string // project root for path sandboxing
}

// NewPolicy creates a new permission policy.
func NewPolicy(mode types.PermissionModeName, rules []types.PermissionEntry, projectRoot string) *Policy {
	return &Policy{
		Mode:    mode,
		Rules:   rules,
		Project: projectRoot,
	}
}

// Evaluate checks if a tool call is allowed.
// Returns Allow, Deny, or Ask.
func (p *Policy) Evaluate(toolName string, args map[string]interface{}) Decision {
	// Check explicit rules first (last match wins)
	for i := len(p.Rules) - 1; i >= 0; i-- {
		rule := p.Rules[i]
		if matchRule(rule.ToolName, toolName) {
			switch rule.Mode {
			case "deny":
				return Deny
			case "allow":
				return Allow
			case "ask":
				return Ask
			}
		}
	}

	// Mode-based defaults
	switch p.Mode {
	case "edit":
		return Allow
	case "auto":
		if isReadOnlyTool(toolName) {
			return Allow
		}
		return p.defaultForWrite(toolName)
	case "plan":
		if isReadOnlyTool(toolName) {
			return Allow
		}
		return Deny // Plan mode blocks all writes
	default: // "ask"
		if isReadOnlyTool(toolName) {
			return Allow
		}
		return Ask
	}
}

// defaultForWrite determines the default for write tools in auto mode.
func (p *Policy) defaultForWrite(toolName string) Decision {
	// High-risk tools always ask
	if toolName == "runCommand" {
		return Ask
	}
	return Ask
}

// CheckPathSandbox verifies that a file path is within the project directory.
// Returns true if the path is safe.
func CheckPathSandbox(filePath, projectRoot string) bool {
	if projectRoot == "" {
		return true // No sandbox configured
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return false
	}
	// Path must be inside project root or equal to it
	return strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) || absPath == absRoot
}

// isReadOnlyTool returns true if the tool name is a read-only tool.
func isReadOnlyTool(name string) bool {
	readOnly := map[string]bool{
		"readFile":    true,
		"searchFiles": true,
		"listDir":     true,
		"gitStatus":   true,
		"gitDiff":     true,
		"glob":        true,
		"grep":        true,
	}
	return readOnly[name]
}

// matchRule checks if a rule pattern matches a tool name.
// Supports exact match and wildcard (*).
func matchRule(pattern, name string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == name {
		return true
	}
	// Simple wildcard: "git*" matches "gitStatus", "gitDiff", etc.
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, strings.TrimSuffix(pattern, "*"))
	}
	return false
}

// DecisionString returns a human-readable string for a decision.
func DecisionString(d Decision) string {
	switch d {
	case Allow:
		return "allow"
	case Deny:
		return "deny"
	case Ask:
		return "ask"
	default:
		return "unknown"
	}
}

// IsReadOnlyBashCommand checks if a shell command is read-only.
func IsReadOnlyBashCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	firstWord := strings.Fields(trimmed)[0]

	readOnlyCommands := []string{
		"ls", "cat", "head", "tail", "grep", "find", "which", "echo",
		"pwd", "whoami", "date", "wc", "sort", "uniq", "diff",
		"git status", "git log", "git diff", "git show", "git branch",
		"git remote", "git tag", "git blame", "git shortlog",
		"node --version", "npm --version", "go version", "python --version",
		"curl -s", "curl --head",
	}

	for _, ro := range readOnlyCommands {
		if firstWord == ro || strings.HasPrefix(trimmed, ro) {
			return true
		}
	}
	return false
}

// FormatToolCall returns a human-readable description of a tool call.
func FormatToolCall(name string, args map[string]interface{}) string {
	switch name {
	case "readFile":
		return fmt.Sprintf("readFile(%s)", StringArg(args, "path"))
	case "writeFile":
		return fmt.Sprintf("writeFile(%s)", StringArg(args, "path"))
	case "editFile":
		return fmt.Sprintf("editFile(%s)", StringArg(args, "path"))
	case "runCommand":
		return fmt.Sprintf("runCommand(%s)", StringArg(args, "command"))
	case "searchFiles":
		return fmt.Sprintf("searchFiles(%s)", StringArg(args, "pattern"))
	case "gitStatus":
		return "gitStatus()"
	case "gitDiff":
		return "gitDiff()"
	case "gitCommit":
		return fmt.Sprintf("gitCommit(%s)", StringArg(args, "message"))
	case "listDir":
		return fmt.Sprintf("listDir(%s)", StringArg(args, "path"))
	default:
		return name
	}
}

// StringArg is a helper to extract string args.
func StringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
