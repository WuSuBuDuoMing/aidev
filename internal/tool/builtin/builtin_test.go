// Package builtin provides tests for built-in tools.
package builtin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- helpers ---

func setupTempDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- ReadTool ---

func TestReadTool_NameAndReadOnly(t *testing.T) {
	tool := NewReadTool()
	if tool.Name() != "readFile" {
		t.Fatalf("expected readFile, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("ReadTool should be read-only")
	}
}

func TestReadTool_Schema(t *testing.T) {
	s := NewReadTool().Schema()
	req, ok := s["required"].([]string)
	if !ok || len(req) != 1 || req[0] != "path" {
		t.Fatal("schema should require 'path'")
	}
}

func TestReadTool_ExecuteFullFile(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "hello.txt", "line one\nline two\nline three")

	result, err := NewReadTool().Execute(context.Background(), map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Output == "(empty file)" {
		t.Fatal("file should not be empty")
	}
}

func TestReadTool_ExecuteLineRange(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "lines.txt", "aaa\nbbb\nccc\nddd\neee")

	result, err := NewReadTool().Execute(context.Background(), map[string]interface{}{
		"path":      path,
		"startLine": float64(2),
		"endLine":   float64(4),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	// Should contain lines 2-4 only
	for _, want := range []string{"2\tbbb", "3\tccc", "4\tddd"} {
		if !containsStr(result.Output, want) {
			t.Fatalf("output missing %q:\n%s", want, result.Output)
		}
	}
	if containsStr(result.Output, "aaa") {
		t.Fatal("should not contain line 1")
	}
}

func TestReadTool_MissingPath(t *testing.T) {
	result, err := NewReadTool().Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without path")
	}
}

func TestReadTool_NonexistentFile(t *testing.T) {
	result, err := NewReadTool().Execute(context.Background(), map[string]interface{}{
		"path": "/nonexistent/file.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail for nonexistent file")
	}
}

func TestReadTool_EmptyFile(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "empty.txt", "")

	result, err := NewReadTool().Execute(context.Background(), map[string]interface{}{
		"path": path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "(empty file)" {
		t.Fatalf("expected '(empty file)', got %q", result.Output)
	}
}

// --- WriteTool ---

func TestWriteTool_NameAndReadOnly(t *testing.T) {
	tool := NewWriteTool()
	if tool.Name() != "writeFile" {
		t.Fatalf("expected writeFile, got %s", tool.Name())
	}
	if tool.ReadOnly() {
		t.Fatal("WriteTool should not be read-only")
	}
}

func TestWriteTool_Execute(t *testing.T) {
	dir := setupTempDir(t)
	path := filepath.Join(dir, "output.txt")

	result, err := NewWriteTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"content": "hello world",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}
}

func TestWriteTool_CreatesDirectories(t *testing.T) {
	dir := setupTempDir(t)
	path := filepath.Join(dir, "deep", "nested", "file.txt")

	result, err := NewWriteTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"content": "deep file",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "deep file" {
		t.Fatalf("unexpected content: %q", string(data))
	}
}

func TestWriteTool_MissingPath(t *testing.T) {
	result, err := NewWriteTool().Execute(context.Background(), map[string]interface{}{
		"content": "hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without path")
	}
}

// --- EditTool ---

func TestEditTool_NameAndReadOnly(t *testing.T) {
	tool := NewEditTool()
	if tool.Name() != "editFile" {
		t.Fatalf("expected editFile, got %s", tool.Name())
	}
	if tool.ReadOnly() {
		t.Fatal("EditTool should not be read-only")
	}
}

func TestEditTool_Execute(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "edit.txt", "hello world")

	result, err := NewEditTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"oldText": "world",
		"newText": "gopher",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}

	data, _ := os.ReadFile(path)
	if string(data) != "hello gopher" {
		t.Fatalf("expected 'hello gopher', got %q", string(data))
	}
}

func TestEditTool_NotFound(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "edit.txt", "hello world")

	result, err := NewEditTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"oldText": "xyz",
		"newText": "abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail when oldText not found")
	}
}

func TestEditTool_MultipleMatches(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "edit.txt", "aaa bbb aaa")

	result, err := NewEditTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"oldText": "aaa",
		"newText": "ccc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail when oldText matches multiple times")
	}
}

func TestEditTool_MissingPath(t *testing.T) {
	result, err := NewEditTool().Execute(context.Background(), map[string]interface{}{
		"oldText": "a",
		"newText": "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without path")
	}
}

func TestEditTool_MissingOldText(t *testing.T) {
	dir := setupTempDir(t)
	path := writeFile(t, dir, "edit.txt", "hello")

	result, err := NewEditTool().Execute(context.Background(), map[string]interface{}{
		"path":    path,
		"newText": "b",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without oldText")
	}
}

// --- GlobTool ---

func TestGlobTool_NameAndReadOnly(t *testing.T) {
	tool := NewGlobTool()
	if tool.Name() != "searchFiles" {
		t.Fatalf("expected searchFiles, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("GlobTool should be read-only")
	}
}

func TestGlobTool_FindGoFiles(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, dir, "main.go", "package main")
	writeFile(t, dir, "readme.md", "# Hello")

	result, err := NewGlobTool().Execute(context.Background(), map[string]interface{}{
		"pattern": "*.go",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if !containsStr(result.Output, "main.go") {
		t.Fatal("should find main.go")
	}
	if containsStr(result.Output, "readme.md") {
		t.Fatal("should not find readme.md")
	}
}

func TestGlobTool_NoMatches(t *testing.T) {
	dir := setupTempDir(t)

	result, err := NewGlobTool().Execute(context.Background(), map[string]interface{}{
		"pattern": "*.xyz",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "No files found." {
		t.Fatalf("expected 'No files found.', got %q", result.Output)
	}
}

// --- GrepTool ---

func TestGrepTool_NameAndReadOnly(t *testing.T) {
	tool := NewGrepTool()
	if tool.Name() != "grep" {
		t.Fatalf("expected grep, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("GrepTool should be read-only")
	}
}

func TestGrepTool_FindPattern(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, dir, "a.go", "package main\nimport \"fmt\"\nfunc main() { fmt.Println() }")
	writeFile(t, dir, "b.go", "package util\nfunc Helper() {}")

	result, err := NewGrepTool().Execute(context.Background(), map[string]interface{}{
		"pattern": "func main",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if !containsStr(result.Output, "a.go") {
		t.Fatal("should find match in a.go")
	}
	if containsStr(result.Output, "b.go") {
		t.Fatal("should not find match in b.go")
	}
}

func TestGrepTool_InvalidRegex(t *testing.T) {
	dir := setupTempDir(t)

	result, err := NewGrepTool().Execute(context.Background(), map[string]interface{}{
		"pattern": "[invalid",
		"path":    dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail with invalid regex")
	}
}

func TestGrepTool_WithGlobFilter(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, dir, "main.go", "package main")
	writeFile(t, dir, "main.py", "def main(): pass")

	result, err := NewGrepTool().Execute(context.Background(), map[string]interface{}{
		"pattern": "main",
		"path":    dir,
		"glob":    "*.go",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsStr(result.Output, "main.go") {
		t.Fatal("should find match in .go file")
	}
	if containsStr(result.Output, "main.py") {
		t.Fatal("should not find match in .py file due to glob filter")
	}
}

// --- BashTool ---

func TestBashTool_NameAndReadOnly(t *testing.T) {
	tool := NewBashTool()
	if tool.Name() != "runCommand" {
		t.Fatalf("expected runCommand, got %s", tool.Name())
	}
	if tool.ReadOnly() {
		t.Fatal("BashTool should not be read-only")
	}
}

func TestBashTool_ExecuteEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip echo test on windows shell differences")
	}
	result, err := NewBashTool().Execute(context.Background(), map[string]interface{}{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if result.Output != "hello" {
		t.Fatalf("expected 'hello', got %q", result.Output)
	}
}

func TestBashTool_MissingCommand(t *testing.T) {
	result, err := NewBashTool().Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without command")
	}
}

// --- GitTools (smoke tests) ---

func TestGitStatusTool_NameAndReadOnly(t *testing.T) {
	tool := NewGitStatusTool()
	if tool.Name() != "gitStatus" {
		t.Fatalf("expected gitStatus, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("GitStatusTool should be read-only")
	}
}

func TestGitDiffTool_NameAndReadOnly(t *testing.T) {
	tool := NewGitDiffTool()
	if tool.Name() != "gitDiff" {
		t.Fatalf("expected gitDiff, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("GitDiffTool should be read-only")
	}
}

func TestGitCommitTool_NameAndReadOnly(t *testing.T) {
	tool := NewGitCommitTool()
	if tool.Name() != "gitCommit" {
		t.Fatalf("expected gitCommit, got %s", tool.Name())
	}
	if tool.ReadOnly() {
		t.Fatal("GitCommitTool should not be read-only")
	}
}

func TestGitCommitTool_MissingMessage(t *testing.T) {
	result, err := NewGitCommitTool().Execute(context.Background(), map[string]interface{}{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Success {
		t.Fatal("should fail without message")
	}
}

// --- LsTool ---

func TestLsTool_NameAndReadOnly(t *testing.T) {
	tool := NewLsTool()
	if tool.Name() != "listDir" {
		t.Fatalf("expected listDir, got %s", tool.Name())
	}
	if !tool.ReadOnly() {
		t.Fatal("LsTool should be read-only")
	}
}

func TestLsTool_Execute(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, dir, "file1.txt", "a")
	writeFile(t, dir, "file2.txt", "b")

	result, err := NewLsTool().Execute(context.Background(), map[string]interface{}{
		"path": dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if !containsStr(result.Output, "file1.txt") || !containsStr(result.Output, "file2.txt") {
		t.Fatalf("should list both files, got: %s", result.Output)
	}
}

func TestLsTool_EmptyDir(t *testing.T) {
	dir := setupTempDir(t)

	result, err := NewLsTool().Execute(context.Background(), map[string]interface{}{
		"path": dir,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Output != "(empty directory)" {
		t.Fatalf("expected '(empty directory)', got %q", result.Output)
	}
}

func TestLsTool_WithSubdirectories(t *testing.T) {
	dir := setupTempDir(t)
	writeFile(t, dir, "top.txt", "top")
	writeFile(t, dir, "sub/deep.txt", "deep")

	result, err := NewLsTool().Execute(context.Background(), map[string]interface{}{
		"path":  dir,
		"depth": float64(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if !containsStr(result.Output, "top.txt") {
		t.Fatal("should list top.txt")
	}
	if !containsStr(result.Output, "deep.txt") {
		t.Fatal("should list deep.txt in subdirectory")
	}
}

// --- Helpers ---

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && contains(s, sub))
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
