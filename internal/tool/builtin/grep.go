// Package builtin implements the grep (content search) tool.
package builtin

import "github.com/WuSuBuDuoMing/aidev/internal/types"

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GrepTool searches for patterns in file contents.
type GrepTool struct{}

func NewGrepTool() *GrepTool { return &GrepTool{} }

func (t *GrepTool) Name() string        { return "grep" }
func (t *GrepTool) Description() string  { return "Search for a regex pattern in file contents across the project." }
func (t *GrepTool) ReadOnly() bool       { return true }
func (t *GrepTool) Schema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regex pattern to search for.",
			},
			"glob": map[string]interface{}{
				"type":        "string",
				"description": "File pattern filter (e.g. '*.go'). Default: all files.",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Root directory. Default: current directory.",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GrepTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	pattern := stringArg(args, "pattern")
	globFilter := stringArg(args, "glob")
	root := stringArg(args, "path")
	if root == "" {
		root = "."
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fail(fmt.Sprintf("Invalid regex: %v", err)), nil
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

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Glob filter
		if globFilter != "" {
			matched, _ := filepath.Match(globFilter, info.Name())
			if !matched {
				return nil
			}
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			if re.MatchString(scanner.Text()) {
				rel, _ := filepath.Rel(root, path)
				matches = append(matches, fmt.Sprintf("%s:%d: %s", rel, lineNum, strings.TrimSpace(scanner.Text())))
				count++
				if count >= 200 {
					return filepath.SkipAll
				}
			}
		}
		return nil
	})

	if err != nil {
		return fail(fmt.Sprintf("Search error: %v", err)), nil
	}

	if len(matches) == 0 {
		return ok("No matches found."), nil
	}

	return ok(strings.Join(matches, "\n")), nil
}
