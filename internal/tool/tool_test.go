package tool

import (
	"context"
	"testing"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// mockTool implements Tool for testing.
type mockTool struct {
	name     string
	readOnly bool
}

func (t *mockTool) Name() string        { return t.name }
func (t *mockTool) Description() string  { return "mock tool " + t.name }
func (t *mockTool) Schema() map[string]interface{} { return map[string]interface{}{} }
func (t *mockTool) ReadOnly() bool       { return t.readOnly }
func (t *mockTool) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	return &types.ToolResult{Success: true, Output: "ok"}, nil
}

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(r.GetAll()) != 0 {
		t.Fatal("New registry should be empty")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &mockTool{name: "testTool", readOnly: true}
	r.Register(tool)

	got, ok := r.Get("testTool")
	if !ok {
		t.Fatal("Get should find registered tool")
	}
	if got.Name() != "testTool" {
		t.Errorf("Got name %q, want %q", got.Name(), "testTool")
	}
}

func TestRegistry_GetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("Get should not find unregistered tool")
	}
}

func TestRegistry_GetAll(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "a", readOnly: true})
	r.Register(&mockTool{name: "b", readOnly: false})

	all := r.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll returned %d tools, want 2", len(all))
	}
}

func TestRegistry_GetDefinitions(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "readFile", readOnly: true})
	r.Register(&mockTool{name: "writeFile", readOnly: false})

	defs := r.GetDefinitions()
	if len(defs) != 2 {
		t.Errorf("GetDefinitions returned %d defs, want 2", len(defs))
	}
}

func TestRegistry_GetReadOnlyTools(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "readFile", readOnly: true})
	r.Register(&mockTool{name: "writeFile", readOnly: false})
	r.Register(&mockTool{name: "grep", readOnly: true})

	roTools := r.GetReadOnlyTools()
	if len(roTools) != 2 {
		t.Errorf("GetReadOnlyTools returned %d, want 2", len(roTools))
	}
}

func TestRegistry_AllReadOnly(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "readFile", readOnly: true})
	r.Register(&mockTool{name: "grep", readOnly: true})
	r.Register(&mockTool{name: "writeFile", readOnly: false})

	// All read-only batch
	allReadOnlyCalls := []types.ToolCall{
		{Name: "readFile"},
		{Name: "grep"},
	}
	if !r.AllReadOnly(allReadOnlyCalls) {
		t.Error("AllReadOnly should return true for all read-only tools")
	}

	// Mixed batch
	mixedCalls := []types.ToolCall{
		{Name: "readFile"},
		{Name: "writeFile"},
	}
	if r.AllReadOnly(mixedCalls) {
		t.Error("AllReadOnly should return false for mixed tools")
	}

	// Unknown tool
	unknownCalls := []types.ToolCall{
		{Name: "nonexistent"},
	}
	if r.AllReadOnly(unknownCalls) {
		t.Error("AllReadOnly should return false for unknown tools")
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()
	r.Register(&mockTool{name: "testTool", readOnly: true})

	result, err := r.Execute(context.Background(), "testTool", nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if !result.Success {
		t.Error("Execute should succeed")
	}
}

func TestRegistry_ExecuteUnknown(t *testing.T) {
	r := NewRegistry()

	result, err := r.Execute(context.Background(), "unknown", nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Success {
		t.Error("Execute of unknown tool should fail")
	}
}

func TestStringArg(t *testing.T) {
	args := map[string]interface{}{
		"path":    "/some/path",
		"number":  42,
		"missing": nil,
	}

	if got := StringArg(args, "path"); got != "/some/path" {
		t.Errorf("StringArg(path) = %q, want %q", got, "/some/path")
	}
	if got := StringArg(args, "number"); got != "" {
		t.Errorf("StringArg(number) = %q, want empty", got)
	}
	if got := StringArg(args, "missing"); got != "" {
		t.Errorf("StringArg(missing) = %q, want empty", got)
	}
	if got := StringArg(args, "nonexistent"); got != "" {
		t.Errorf("StringArg(nonexistent) = %q, want empty", got)
	}
}

func TestIntArg(t *testing.T) {
	args := map[string]interface{}{
		"float": 3.14,
		"int":   42,
		"str":   "hello",
	}

	if got := IntArg(args, "float"); got != 3 {
		t.Errorf("IntArg(float) = %d, want 3", got)
	}
	if got := IntArg(args, "int"); got != 42 {
		t.Errorf("IntArg(int) = %d, want 42", got)
	}
	if got := IntArg(args, "str"); got != 0 {
		t.Errorf("IntArg(str) = %d, want 0", got)
	}
	if got := IntArg(args, "missing"); got != 0 {
		t.Errorf("IntArg(missing) = %d, want 0", got)
	}
}

func TestBoolArg(t *testing.T) {
	args := map[string]interface{}{
		"yes": true,
		"no":  false,
		"str": "true",
	}

	if got := BoolArg(args, "yes"); !got {
		t.Error("BoolArg(yes) should be true")
	}
	if got := BoolArg(args, "no"); got {
		t.Error("BoolArg(no) should be false")
	}
	if got := BoolArg(args, "str"); got {
		t.Error("BoolArg(str) should be false")
	}
	if got := BoolArg(args, "missing"); got {
		t.Error("BoolArg(missing) should be false")
	}
}
