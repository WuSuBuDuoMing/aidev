// Package mcp provides the MCP tool adapter that wraps MCP tools
// as NeoCode tools in the tool registry.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/WuSuBuDuoMing/aidev/internal/tool"
	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// MCPToolAdapter wraps an MCP tool as a NeoCode tool.Tool.
type MCPToolAdapter struct {
	mcpTool MCPTool
	client  *Client
}

// NewMCPToolAdapter creates a new adapter.
func NewMCPToolAdapter(mcpTool MCPTool, client *Client) *MCPToolAdapter {
	return &MCPToolAdapter{mcpTool: mcpTool, client: client}
}

func (a *MCPToolAdapter) Name() string {
	return fmt.Sprintf("mcp__%s__%s", a.client.Name(), a.mcpTool.Name)
}

func (a *MCPToolAdapter) Description() string {
	return fmt.Sprintf("[MCP:%s] %s", a.client.Name(), a.mcpTool.Description)
}

func (a *MCPToolAdapter) Schema() map[string]interface{} {
	return a.mcpTool.InputSchema
}

func (a *MCPToolAdapter) ReadOnly() bool {
	// Conservative: assume MCP tools are not read-only
	return false
}

func (a *MCPToolAdapter) Execute(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
	return a.client.CallTool(ctx, a.mcpTool.Name, args)
}

// Manager manages multiple MCP clients and their tools.
type Manager struct {
	clients []*Client
}

// NewManager creates a new MCP manager.
func NewManager() *Manager {
	return &Manager{}
}

// Connect starts an MCP server and registers its tools.
func (m *Manager) Connect(ctx context.Context, cfg MCPConfig, registry *tool.Registry) error {
	client, err := Start(ctx, cfg)
	if err != nil {
		return err
	}

	m.clients = append(m.clients, client)

	for _, mcpTool := range client.Tools() {
		adapter := NewMCPToolAdapter(mcpTool, client)
		registry.Register(adapter)
	}

	return nil
}

// ConnectFromConfig reads .mcp.json and connects all servers.
func (m *Manager) ConnectFromConfig(ctx context.Context, configPath string, registry *tool.Registry) error {
	// Read .mcp.json file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil // no .mcp.json, that's fine
	}

	var mcpConfig struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
			URL     string   `json:"url"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &mcpConfig); err != nil {
		return fmt.Errorf("parse .mcp.json: %w", err)
	}

	for name, serverCfg := range mcpConfig.MCPServers {
		cfg := MCPConfig{
			Name:    name,
			Command: serverCfg.Command,
			Args:    serverCfg.Args,
			URL:     serverCfg.URL,
		}
		if err := m.Connect(ctx, cfg, registry); err != nil {
			fmt.Printf("  ⚠ MCP server %q failed: %v\n", name, err)
		} else {
			fmt.Printf("  ✓ MCP server %q connected (%d tools)\n", name, len(m.clients[len(m.clients)-1].Tools()))
		}
	}

	return nil
}

// CloseAll shuts down all MCP clients.
func (m *Manager) CloseAll() {
	for _, c := range m.clients {
		c.Close()
	}
}

// ListServers returns info about connected MCP servers.
func (m *Manager) ListServers() []string {
	var names []string
	for _, c := range m.clients {
		names = append(names, fmt.Sprintf("%s (%d tools)", c.Name(), len(c.Tools())))
	}
	return names
}

// ParseMCPServerName extracts the server name from an MCP tool name (mcp__server__tool).
func ParseMCPServerName(toolName string) string {
	if strings.HasPrefix(toolName, "mcp__") {
		parts := strings.SplitN(toolName, "__", 3)
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	return ""
}
