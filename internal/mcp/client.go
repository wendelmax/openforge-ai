package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client communicates with an MCP server subprocess over stdio.
type Client struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   *bufio.Scanner
	idSeq    atomic.Int64
	pending  map[int64]chan response
	capabilities ServerCapabilities
	tools    []Tool
}

// NewClient creates an MCP client that spawns the given command.
func NewClient(name string, args ...string) *Client {
	return &Client{
		cmd:     exec.Command(name, args...),
		pending: make(map[int64]chan response),
	}
}

// Start launches the MCP server and performs initialization.
func (c *Client) Start(ctx context.Context) error {
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	c.stdout = bufio.NewScanner(stdout)
	c.stdout.Buffer(make([]byte, 0, 1<<20), 1<<20)

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start mcp server: %w", err)
	}

	go drainStderr(stderr)
	go c.readResponses()

	if err := c.initialize(ctx); err != nil {
		c.cmd.Process.Kill()
		return fmt.Errorf("mcp init: %w", err)
	}

	if err := c.listTools(ctx); err != nil {
		c.cmd.Process.Kill()
		return fmt.Errorf("mcp list tools: %w", err)
	}

	return nil
}

func (c *Client) initialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ClientCapabilities{},
		ClientInfo: Implementation{
			Name:    "openforge",
			Version: "0.1.0",
		},
	}
	var result InitializeResult
	if err := c.call(ctx, "initialize", params, &result); err != nil {
		return err
	}
	c.capabilities = result.Capabilities
	return nil
}

func (c *Client) listTools(ctx context.Context) error {
	var result ListToolsResult
	if err := c.call(ctx, "tools/list", nil, &result); err != nil {
		return err
	}
	c.tools = result.Tools
	return nil
}

// Tools returns the list of tools from the server.
func (c *Client) Tools() []Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := make([]Tool, len(c.tools))
	copy(t, c.tools)
	return t
}

// CallTool invokes a tool by name with the given arguments.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	params := CallToolParams{
		Name:      name,
		Arguments: args,
	}
	var result CallToolResult
	if err := c.call(ctx, "tools/call", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Close shuts down the MCP server.
func (c *Client) Close() error {
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, params interface{}, result interface{}) error {
	id := c.idSeq.Add(1)
	ch := make(chan response, 1)

	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	req := request{
		JSONRPC: "2.0",
		ID:      int(id),
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	c.mu.Lock()
	_, err = fmt.Fprintf(c.stdin, "Content-Length: %d\r\n\r\n%s", len(data), data)
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return fmt.Errorf("mcp error: %s (code %d)", resp.Error.Message, resp.Error.Code)
		}
		if result == nil {
			return nil
		}
		data, err := json.Marshal(resp.Result)
		if err != nil {
			return fmt.Errorf("marshal response: %w", err)
		}
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("unmarshal result: %w", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) readResponses() {
	var buf []byte
	for c.stdout.Scan() {
		line := c.stdout.Text()
		if len(line) == 0 && len(buf) > 0 {
			var resp response
			if err := json.Unmarshal(buf, &resp); err != nil {
				buf = nil
				continue
			}
			c.mu.Lock()
			ch, ok := c.pending[int64(resp.ID)]
			c.mu.Unlock()
			if ok {
				ch <- resp
			}
			buf = nil
		} else if len(line) > 0 {
			buf = append(buf, []byte(line)...)
		}
	}
}

func drainStderr(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// stderr is drained but discarded
	}
}
