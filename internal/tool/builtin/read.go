// Package builtin implements the readFile tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// ReadTool reads file contents.
type ReadTool struct{}

func NewReadTool() *ReadTool { return &ReadTool{} }

func (t *ReadTool) Name() string        { return "readFile" }
func (t *ReadTool) Description() string  { return "Read the contents of a file at the given path." }
func (t *ReadTool) ReadOnly() bool       { return true }
func (t *ReadTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Absolute or relative path to the file.",
			},
			"startLine": map[string]interface{}{
				"type":        "integer",
				"description": "Start line (1-based, inclusive).",
			},
			"endLine": map[string]interface{}{
				"type":        "integer",
				"description": "End line (1-based, inclusive).",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	filePath := stringArg(args, "path")
	if filePath == "" {
		return fail("path is required"), nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fail(fmt.Sprintf("Cannot open file: %v", err)), nil
	}
	defer file.Close()

	startLine := intArg(args, "startLine")
	endLine := intArg(args, "endLine")

	scanner := bufio.NewScanner(file)
	var lines []string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if startLine > 0 && lineNum < startLine {
			continue
		}
		if endLine > 0 && lineNum > endLine {
			break
		}
		lines = append(lines, fmt.Sprintf("%d\t%s", lineNum, scanner.Text()))
	}

	if len(lines) == 0 {
		return ok("(empty file)"), nil
	}

	return ok(strings.Join(lines, "\n")), nil
}
