// Package builtin implements the bash (shell command) tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// BashTool executes shell commands.
type BashTool struct{}

func NewBashTool() *BashTool { return &BashTool{} }

func (t *BashTool) Name() string        { return "runCommand" }
func (t *BashTool) Description() string  { return "Execute a shell command and return its output." }
func (t *BashTool) ReadOnly() bool       { return false }
func (t *BashTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to run.",
			},
			"cwd": map[string]interface{}{
				"type":        "string",
				"description": "Working directory. Default: current directory.",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in milliseconds. Default: 30000.",
			},
		},
		"required": []string{"command"},
	}
}

func (t *BashTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	command := stringArg(args, "command")
	if command == "" {
		return fail("command is required"), nil
	}

	cwd := stringArg(args, "cwd")
	timeout := intArg(args, "timeout")
	if timeout == 0 {
		timeout = 30000
	}

	// Determine shell
	shell := "bash"
	shellArgs := []string{"-c", command}
	if isWindows() {
		shell = "cmd"
		shellArgs = []string{"/C", command}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, shellArgs...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fail(fmt.Sprintf("Command timed out after %dms\n%s", timeout, outputStr)), nil
		}
		return fail(fmt.Sprintf("Exit error: %v\n%s", err, outputStr)), nil
	}

	return ok(outputStr), nil
}
