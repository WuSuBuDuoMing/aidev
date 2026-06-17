// Package builtin implements the writeFile tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// WriteTool writes content to files.
type WriteTool struct{}

func NewWriteTool() *WriteTool { return &WriteTool{} }

func (t *WriteTool) Name() string        { return "writeFile" }
func (t *WriteTool) Description() string  { return "Write content to a file. Creates parent directories if needed." }
func (t *WriteTool) ReadOnly() bool       { return false }
func (t *WriteTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write.",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	filePath := stringArg(args, "path")
	content := stringArg(args, "content")

	if filePath == "" {
		return fail("path is required"), nil
	}

	// Create parent directories
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fail(fmt.Sprintf("Cannot create directory: %v", err)), nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return fail(fmt.Sprintf("Cannot write file: %v", err)), nil
	}

	return ok(fmt.Sprintf("Wrote %s (%d bytes)", filePath, len(content))), nil
}
