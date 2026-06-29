package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type LSPConfig struct {
	Command    []string
	RootMarker string
	Language   string
	Extensions []string
}

var DefaultLSPs = map[string]LSPConfig{
	"gopls":                      {Command: []string{"gopls", "serve"}, RootMarker: "go.mod", Language: "go", Extensions: []string{".go"}},
	"typescript-language-server": {Command: []string{"typescript-language-server", "--stdio"}, RootMarker: "package.json", Language: "typescript", Extensions: []string{".ts", ".tsx"}},
	"rust-analyzer":              {Command: []string{"rust-analyzer"}, RootMarker: "Cargo.toml", Language: "rust", Extensions: []string{".rs"}},
	"pyright":                    {Command: []string{"pyright-langserver", "--stdio"}, RootMarker: "pyproject.toml", Language: "python", Extensions: []string{".py", ".pyi"}},
}

type Client struct {
	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	idSeq  int
}

func (c *Client) call(ctx context.Context, method string, params, result any) error {
	c.mu.Lock()
	c.idSeq++
	id := c.idSeq
	c.mu.Unlock()

	req := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		req["params"] = params
	}
	data, _ := json.Marshal(req)

	c.mu.Lock()
	_, err := fmt.Fprintf(c.stdin, "Content-Length: %d\r\n\r\n%s", len(data), data)
	c.mu.Unlock()
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	for c.stdout.Scan() {
		line := c.stdout.Text()
		if !strings.HasPrefix(line, "Content-Length:") {
			continue
		}
		l, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:")))
		c.stdout.Scan()
		buf := make([]byte, l)
		for i := 0; i < l; {
			if !c.stdout.Scan() {
				return fmt.Errorf("unexpected EOF")
			}
			lb := c.stdout.Bytes()
			n := copy(buf[i:], lb)
			i += n
			if i < l {
				buf[i] = '\n'
				i++
			}
		}
		var resp struct {
			ID     int            `json:"id,omitempty"`
			Error  *struct{ Message string } `json:"error,omitempty"`
			Result json.RawMessage `json:"result,omitempty"`
		}
		json.Unmarshal(buf, &resp)
		if resp.Error != nil {
			return fmt.Errorf("lsp error: %s", resp.Error.Message)
		}
		if result != nil && resp.Result != nil {
			json.Unmarshal(resp.Result, result)
		}
		return nil
	}
	return fmt.Errorf("no response from LSP")
}

func (c *Client) notify(ctx context.Context, method string, params any) error {
	req := map[string]any{"jsonrpc": "2.0", "method": method}
	if params != nil {
		req["params"] = params
	}
	data, _ := json.Marshal(req)
	c.mu.Lock()
	_, err := fmt.Fprintf(c.stdin, "Content-Length: %d\r\n\r\n%s", len(data), data)
	c.mu.Unlock()
	return err
}

func (c *Client) start(ctx context.Context) error {
	var err error
	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin: %w", err)
	}
	stdoutPipe, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout: %w", err)
	}
	c.stdout = bufio.NewScanner(stdoutPipe)
	c.stdout.Buffer(make([]byte, 0, 1<<20), 1<<20)
	stderrPipe, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr: %w", err)
	}
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	go io.Copy(io.Discard, stderrPipe)
	if err := c.call(ctx, "initialize", map[string]any{"processId": nil, "rootUri": nil, "capabilities": map[string]any{}}, nil); err != nil {
		c.cmd.Process.Kill()
		return fmt.Errorf("initialize: %w", err)
	}
	c.notify(ctx, "initialized", map[string]any{})
	return nil
}

func (c *Client) shutdown(ctx context.Context) error {
	c.call(ctx, "shutdown", nil, nil)
	c.notify(ctx, "exit", nil)
	if c.cmd != nil && c.cmd.Process != nil {
		return c.cmd.Process.Kill()
	}
	return nil
}

type Manager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewManager() *Manager { return &Manager{clients: make(map[string]*Client)} }

func (m *Manager) AutoDetect(rootPath string) error {
	for name, cfg := range DefaultLSPs {
		if _, err := os.Stat(filepath.Join(rootPath, cfg.RootMarker)); err == nil {
			m.mu.RLock()
			_, exists := m.clients[name]
			m.mu.RUnlock()
			if exists {
				continue
			}
			m.StartLSP(context.Background(), name, cfg)
		}
	}
	return nil
}

func (m *Manager) StartLSP(ctx context.Context, name string, cfg LSPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clients[name]; exists {
		return nil
	}
	client := &Client{cmd: exec.Command(cfg.Command[0], cfg.Command[1:]...)}
	if err := client.start(ctx); err != nil {
		return err
	}
	m.clients[name] = client
	return nil
}

func (m *Manager) StopLSP(ctx context.Context, name string) error {
	m.mu.Lock()
	client, ok := m.clients[name]
	if ok {
		delete(m.clients, name)
	}
	m.mu.Unlock()
	if !ok {
		return nil
	}
	return client.shutdown(ctx)
}

func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	names := make([]string, 0, len(m.clients))
	for name := range m.clients {
		names = append(names, name)
	}
	clients := m.clients
	m.clients = make(map[string]*Client)
	m.mu.Unlock()
	for _, name := range names {
		clients[name].shutdown(ctx)
	}
	return nil
}
