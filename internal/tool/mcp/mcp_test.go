// Package mcp provides tests for the MCP tool adapter and utilities.
package mcp

import "testing"

func TestParseMCPServerName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"mcp__filesystem__readFile", "filesystem"},
		{"mcp__custom-server__search", "custom-server"},
		{"mcp__server__tool_name", "server"},
		{"mcp__s", "s"},
		{"not_mcp_tool", ""},
		{"mcp__", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseMCPServerName(tt.input)
			if got != tt.want {
				t.Errorf("ParseMCPServerName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMCPTool_ToToolDefinition(t *testing.T) {
	tool := MCPTool{
		Name:        "readFile",
		Description: "Read a file from the filesystem",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"path"},
		},
		ServerName: "filesystem",
	}

	def := tool.ToToolDefinition()

	if def.Name != "mcp__filesystem__readFile" {
		t.Errorf("expected name mcp__filesystem__readFile, got %s", def.Name)
	}
	if def.Description != "Read a file from the filesystem" {
		t.Errorf("unexpected description: %s", def.Description)
	}
	if def.Parameters == nil {
		t.Fatal("parameters should not be nil")
	}
}

func TestMCPToolAdapter_Name(t *testing.T) {
	client := &Client{name: "test-server"}
	mcpTool := MCPTool{Name: "doThing", Description: "Does a thing"}

	adapter := NewMCPToolAdapter(mcpTool, client)

	if adapter.Name() != "mcp__test-server__doThing" {
		t.Errorf("expected mcp__test-server__doThing, got %s", adapter.Name())
	}
}

func TestMCPToolAdapter_Description(t *testing.T) {
	client := &Client{name: "myserver"}
	mcpTool := MCPTool{Name: "tool", Description: "A useful tool"}

	adapter := NewMCPToolAdapter(mcpTool, client)

	expected := "[MCP:myserver] A useful tool"
	if adapter.Description() != expected {
		t.Errorf("expected %q, got %q", expected, adapter.Description())
	}
}

func TestMCPToolAdapter_ReadOnly(t *testing.T) {
	client := &Client{name: "s"}
	adapter := NewMCPToolAdapter(MCPTool{}, client)

	// MCP tools should be conservative (not read-only)
	if adapter.ReadOnly() {
		t.Error("MCPToolAdapter should not be read-only (conservative assumption)")
	}
}

func TestMCPToolAdapter_Schema(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file": map[string]interface{}{"type": "string"},
		},
	}
	client := &Client{name: "s"}
	adapter := NewMCPToolAdapter(MCPTool{InputSchema: schema}, client)

	got := adapter.Schema()
	if got["type"] != "object" {
		t.Error("schema type should be object")
	}
}

func TestManager_NewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager should not return nil")
	}
}

func TestManager_ListServersEmpty(t *testing.T) {
	m := NewManager()
	servers := m.ListServers()
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

func TestManager_CloseAllEmpty(t *testing.T) {
	m := NewManager()
	// Should not panic on empty manager
	m.CloseAll()
}

func TestManager_ConnectFromConfigMissingFile(t *testing.T) {
	m := NewManager()

	// ConnectFromConfig with nonexistent path should return nil (no error)
	err := m.ConnectFromConfig(nil, "/nonexistent/.mcp.json", nil)
	if err != nil {
		t.Errorf("missing config file should return nil error, got %v", err)
	}
}
