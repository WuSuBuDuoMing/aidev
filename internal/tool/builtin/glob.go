// Package builtin implements the searchFiles (glob) tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GlobTool finds files matching a glob pattern.
type GlobTool struct{}

func NewGlobTool() *GlobTool { return &GlobTool{} }

func (t *GlobTool) Name() string        { return "searchFiles" }
func (t *GlobTool) Description() string  { return "Find files matching a glob pattern in the project." }
func (t *GlobTool) ReadOnly() bool       { return true }
func (t *GlobTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern (e.g. '**/*.ts', '*.go').",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Root directory. Default: current directory.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	pattern := stringArg(args, "pattern")
	root := stringArg(args, "path")
	if root == "" {
		root = "."
	}

	ignoreDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
	}

	var matches []string
	count := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Simple glob matching
		matched, _ := filepath.Match(pattern, info.Name())
		if !matched {
			// Try matching the full relative path
			rel, _ := filepath.Rel(root, path)
			matched, _ = filepath.Match(pattern, rel)
		}

		if matched {
			matches = append(matches, path)
			count++
			if count >= 200 {
				return filepath.SkipDir
			}
		}
		return nil
	})

	if err != nil {
		return fail(fmt.Sprintf("Search error: %v", err)), nil
	}

	if len(matches) == 0 {
		return ok("No files found."), nil
	}

	return ok(strings.Join(matches, "\n")), nil
}
