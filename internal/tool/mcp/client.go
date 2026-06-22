// Package mcp implements the Model Context Protocol (MCP) client.
// Supports stdio transport for local plugins.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WuSuBuDuoMing/aidev/internal/types"
)

// Client represents an MCP client connection.
type Client struct {
	name    string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Scanner
	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int64]chan *jsonRPCResponse
	tools   []MCPTool
}

// MCPTool describes a tool exposed by an MCP server.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	ServerName  string                 `json:"serverName"`
}

// jsonRPCRequest is a JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// jsonRPCError is a JSON-RPC 2.0 error.
type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPConfig defines an MCP server configuration.
type MCPConfig struct {
	Name    string   `toml:"name"`
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
	URL     string   `toml:"url"`
}

// Start launches an MCP server via stdio transport.
func Start(ctx context.Context, cfg MCPConfig) (*Client, error) {
	if cfg.Command == "" {
		return nil, fmt.Errorf("MCP server %q requires a command", cfg.Name)
	}

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start MCP server %q: %w", cfg.Name, err)
	}

	c := &Client{
		name:    cfg.Name,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewScanner(stdout),
		pending: make(map[int64]chan *jsonRPCResponse),
	}

	// Start response reader
	go c.readLoop()

	// Initialize the connection
	if err := c.initialize(); err != nil {
		c.Close()
		return nil, fmt.Errorf("initialize MCP %q: %w", cfg.Name, err)
	}

	// Discover tools
	if err := c.discoverTools(); err != nil {
		c.Close()
		return nil, fmt.Errorf("discover tools for %q: %w", cfg.Name, err)
	}

	return c, nil
}

// Close shuts down the MCP client and its server process.
func (c *Client) Close() {
	c.stdin.Close()
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
		c.cmd.Wait()
	}
}

// Name returns the server name.
func (c *Client) Name() string { return c.name }

// Tools returns the tools exposed by this server.
func (c *Client) Tools() []MCPTool { return c.tools }

// CallTool invokes a tool on the MCP server.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*types.ToolResult, error) {
	result, err := c.call(ctx, "tools/call", map[string]interface{}{
		"name":      name,
		"arguments": args,
	})
	if err != nil {
		return &types.ToolResult{Success: false, Error: err.Error()}, nil
	}

	var resp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return &types.ToolResult{Success: false, Error: "failed to parse tool result"}, nil
	}

	output := ""
	for _, c := range resp.Content {
		if c.Type == "text" {
			output += c.Text
		}
	}

	return &types.ToolResult{
		Success: !resp.IsError,
		Output:  output,
		Error: func() string {
			if resp.IsError {
				return output
			}
			return ""
		}(),
	}, nil
}

// initialize performs the MCP initialize handshake.
func (c *Client) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.call(ctx, "initialize", map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "neocode",
			"version": "2.6.0",
		},
	})
	if err != nil {
		return err
	}

	// Send initialized notification
	return c.notify("notifications/initialized", nil)
}

// discoverTools fetches the list of tools from the server.
func (c *Client) discoverTools() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return err
	}

	var resp struct {
		Tools []MCPTool `json:"tools"`
	}
	if err := json.Unmarshal(result, &resp); err != nil {
		return fmt.Errorf("parse tools: %w", err)
	}

	c.tools = resp.Tools
	for i := range c.tools {
		c.tools[i].ServerName = c.name
	}

	return nil
}

// call sends a JSON-RPC request and waits for the response.
func (c *Client) call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	ch := make(chan *jsonRPCResponse, 1)

	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	c.mu.Lock()
	_, err = c.stdin.Write(append(data, '\n'))
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("MCP error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify sends a JSON-RPC notification (no response expected).
func (c *Client) notify(method string, params interface{}) error {
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	_, err = c.stdin.Write(append(data, '\n'))
	return err
}

// readLoop reads JSON-RPC responses from the server's stdout.
func (c *Client) readLoop() {
	for c.stdout.Scan() {
		line := c.stdout.Bytes()
		var resp jsonRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		if resp.ID == 0 {
			continue // notification
		}

		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		c.mu.Unlock()

		if ok {
			ch <- &resp
		}
	}
}

// ToToolDefinition converts an MCPTool to a NeoCode ToolDefinition.
func (t MCPTool) ToToolDefinition() types.ToolDefinition {
	return types.ToolDefinition{
		Name:        fmt.Sprintf("mcp__%s__%s", t.ServerName, t.Name),
		Description: t.Description,
		Parameters:  t.InputSchema,
	}
}
