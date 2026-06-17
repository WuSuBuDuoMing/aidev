// Package builtin implements the git tools (status, diff, commit).
package builtin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// GitStatusTool shows git status.
type GitStatusTool struct{}

func NewGitStatusTool() *GitStatusTool { return &GitStatusTool{} }

func (t *GitStatusTool) Name() string        { return "gitStatus" }
func (t *GitStatusTool) Description() string  { return "Show the current git status of the repository." }
func (t *GitStatusTool) ReadOnly() bool       { return true }
func (t *GitStatusTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *GitStatusTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	out, err := runGit("status", "--short", "--branch")
	if err != nil {
		return fail(fmt.Sprintf("Git error: %v\n%s", err, out)), nil
	}
	if strings.TrimSpace(out) == "" {
		return ok("Working tree clean."), nil
	}
	return ok(strings.TrimSpace(out)), nil
}

// GitDiffTool shows git diff.
type GitDiffTool struct{}

func NewGitDiffTool() *GitDiffTool { return &GitDiffTool{} }

func (t *GitDiffTool) Name() string        { return "gitDiff" }
func (t *GitDiffTool) Description() string  { return "Show git diff. Optionally diff staged changes or a specific file." }
func (t *GitDiffTool) ReadOnly() bool       { return true }
func (t *GitDiffTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"staged": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, show staged changes.",
			},
			"file": map[string]interface{}{
				"type":        "string",
				"description": "Optional file path to diff.",
			},
		},
	}
}

func (t *GitDiffTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	cmdArgs := []string{"diff"}
	if boolArg(args, "staged") {
		cmdArgs = append(cmdArgs, "--staged")
	}
	if file := stringArg(args, "file"); file != "" {
		cmdArgs = append(cmdArgs, "--", file)
	}

	out, err := runGit(cmdArgs...)
	if err != nil {
		return fail(fmt.Sprintf("Git error: %v\n%s", err, out)), nil
	}
	if strings.TrimSpace(out) == "" {
		return ok("No changes."), nil
	}
	return ok(strings.TrimSpace(out)), nil
}

// GitCommitTool creates a git commit.
type GitCommitTool struct{}

func NewGitCommitTool() *GitCommitTool { return &GitCommitTool{} }

func (t *GitCommitTool) Name() string { return "gitCommit" }
func (t *GitCommitTool) Description() string {
	return "Create a git commit with the given message. Stages all changes first."
}
func (t *GitCommitTool) ReadOnly() bool { return false }
func (t *GitCommitTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Commit message.",
			},
			"all": map[string]interface{}{
				"type":        "boolean",
				"description": "If true, run git add -A before committing.",
			},
		},
		"required": []string{"message"},
	}
}

func (t *GitCommitTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	message := stringArg(args, "message")
	if message == "" {
		return fail("message is required"), nil
	}

	if boolArg(args, "all") {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = "."
		if out, err := cmd.CombinedOutput(); err != nil {
			return fail(fmt.Sprintf("git add error: %v\n%s", err, out)), nil
		}
	}

	// Use exec.Command with separate args to avoid shell injection
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fail(fmt.Sprintf("git commit error: %v\n%s", err, string(out))), nil
	}

	return ok(strings.TrimSpace(string(out))), nil
}

// LsTool lists directory contents.
type LsTool struct{}

func NewLsTool() *LsTool { return &LsTool{} }

func (t *LsTool) Name() string        { return "listDir" }
func (t *LsTool) Description() string  { return "List files and directories at the given path." }
func (t *LsTool) ReadOnly() bool       { return true }
func (t *LsTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path. Default: current directory.",
			},
			"depth": map[string]interface{}{
				"type":        "integer",
				"description": "Recursion depth. Default: 1.",
			},
		},
	}
}

func (t *LsTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	dir := stringArg(args, "path")
	if dir == "" {
		dir = "."
	}
	depth := intArg(args, "depth")
	if depth == 0 {
		depth = 1
	}

	ignoreDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
	}

	var lines []string
	var listDir func(d string, currentDepth int)
	listDir = func(d string, currentDepth int) {
		if currentDepth > depth {
			return
		}
		entries, err := os.ReadDir(d)
		if err != nil {
			return
		}
		indent := strings.Repeat("  ", currentDepth-1)
		for _, e := range entries {
			if ignoreDirs[e.Name()] {
				continue
			}
			if e.IsDir() {
				lines = append(lines, fmt.Sprintf("%s%s/", indent, e.Name()))
				listDir(filepath_join(d, e.Name()), currentDepth+1)
			} else {
				lines = append(lines, fmt.Sprintf("%s%s", indent, e.Name()))
			}
		}
	}

	listDir(dir, 1)

	if len(lines) == 0 {
		return ok("(empty directory)"), nil
	}
	return ok(strings.Join(lines, "\n")), nil
}

func filepath_join(a, b string) string {
	return a + "/" + b
}
