// Package builtin contains shared helper functions for all built-in tools.
package builtin

import (
	"os/exec"
	"runtime"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// ok creates a successful result.
func ok(output string) *types.ToolResult {
	return &types.ToolResult{Success: true, Output: output}
}

// fail creates a failed result.
func fail(err string) *types.ToolResult {
	return &types.ToolResult{Success: false, Error: err}
}

// stringArg extracts a string from args.
func stringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// intArg extracts an int from args.
func intArg(args map[string]interface{}, key string) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}

// boolArg extracts a bool from args.
func boolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// isWindows returns true on Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// runGit runs a git command and returns output.
func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
