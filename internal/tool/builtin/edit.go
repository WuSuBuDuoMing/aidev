// Package builtin implements the editFile tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EditTool performs targeted find-and-replace edits on files.
type EditTool struct{}

func NewEditTool() *EditTool { return &EditTool{} }

func (t *EditTool) Name() string { return "editFile" }
func (t *EditTool) Description() string {
	return "Apply a targeted edit to a file by replacing oldText with newText. The oldText must match exactly."
}
func (t *EditTool) ReadOnly() bool { return false }
func (t *EditTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path.",
			},
			"oldText": map[string]interface{}{
				"type":        "string",
				"description": "Exact text to find and replace.",
			},
			"newText": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text.",
			},
		},
		"required": []string{"path", "oldText", "newText"},
	}
}

func (t *EditTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	filePath := stringArg(args, "path")
	oldText := stringArg(args, "oldText")
	newText := stringArg(args, "newText")

	if filePath == "" {
		return fail("path is required"), nil
	}
	if oldText == "" {
		return fail("oldText is required"), nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fail(fmt.Sprintf("Cannot read file: %v", err)), nil
	}

	content := string(data)

	if !strings.Contains(content, oldText) {
		return fail("oldText not found in file"), nil
	}

	occurrences := strings.Count(content, oldText)
	if occurrences > 1 {
		return fail(fmt.Sprintf("oldText matches %d times. Please provide a more specific snippet.", occurrences)), nil
	}

	// Use Replace with function to avoid $ special chars in newText
	newContent := strings.Replace(content, oldText, newText, 1)

	if err := os.WriteFile(filePath, []byte(newContent), 0o644); err != nil {
		return fail(fmt.Sprintf("Cannot write file: %v", err)), nil
	}

	// Generate a simple diff summary
	removed := len(oldText)
	added := len(newText)
	return ok(fmt.Sprintf("Edited %s (-%d/+%d bytes)", filePath, removed, added)), nil
}
