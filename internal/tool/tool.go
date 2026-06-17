// Package tool defines the tool interface and registry for NeoCode.
package tool

import (
	"context"
	"encoding/json"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// Tool is the interface all tools must implement.
type Tool interface {
	// Name returns the tool's unique name.
	Name() string

	// Description returns a human-readable description for the AI.
	Description() string

	// Schema returns the JSON Schema for the tool's parameters.
	Schema() map[string]interface{}

	// Execute runs the tool with the given arguments.
	Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error)

	// ReadOnly returns true if the tool has no side effects.
	// Used for parallel dispatch optimization.
	ReadOnly() bool
}

// Registry manages all registered tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// GetAll returns all registered tools.
func (r *Registry) GetAll() []Tool {
	var tools []Tool
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetDefinitions returns tool definitions formatted for AI providers.
func (r *Registry) GetDefinitions() []types.ToolDefinition {
	var defs []types.ToolDefinition
	for _, t := range r.tools {
		defs = append(defs, types.ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.Schema(),
		})
	}
	return defs
}

// GetReadOnlyTools returns the names of all read-only tools.
func (r *Registry) GetReadOnlyTools() []string {
	var names []string
	for name, t := range r.tools {
		if t.ReadOnly() {
			names = append(names, name)
		}
	}
	return names
}

// AllReadOnly checks if all tool calls in a batch are read-only.
func (r *Registry) AllReadOnly(toolCalls []types.ToolCall) bool {
	for _, tc := range toolCalls {
		t, ok := r.tools[tc.Name]
		if !ok || !t.ReadOnly() {
			return false
		}
	}
	return true
}

// Execute runs a tool by name with permission checking.
func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (*types.ToolResult, error) {
	t, ok := r.tools[name]
	if !ok {
		return &types.ToolResult{
			Success: false,
			Error:   "Unknown tool: " + name,
		}, nil
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return &types.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	return result, nil
}

// StringArg extracts a string argument from the args map.
func StringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// IntArg extracts an integer argument from the args map.
func IntArg(args map[string]interface{}, key string) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		case json.Number:
			i, _ := n.Int64()
			return int(i)
		}
	}
	return 0
}

// BoolArg extracts a boolean argument from the args map.
func BoolArg(args map[string]interface{}, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
