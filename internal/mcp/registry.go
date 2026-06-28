package mcp

import (
	"context"
	"fmt"
	"sync"

	"github.com/openforge-ai/openforge/internal/config"
	"github.com/openforge-ai/openforge/internal/skill"
)

// ConnectFromConfig connects all MCP servers defined in the config.
func ConnectFromConfig(ctx context.Context, r *Registry, cfg *config.Config) error {
	for name, serverCfg := range cfg.MCP {
		mcpCfg := ServerConfig{
			Command: serverCfg.Command,
			Args:    serverCfg.Args,
			Env:     serverCfg.Env,
		}
		if err := r.Connect(ctx, name, mcpCfg); err != nil {
			return fmt.Errorf("mcp server %s: %w", name, err)
		}
	}
	return nil
}

// ServerConfig defines how to launch an MCP server.
type ServerConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,omitempty"`
	Env     []string `yaml:"env,omitempty"`
}

// Registry manages multiple MCP server connections.
type Registry struct {
	mu      sync.Mutex
	clients map[string]*Client
}

// NewRegistry creates an empty MCP registry.
func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]*Client),
	}
}

// Connect creates and starts an MCP client from config.
func (r *Registry) Connect(ctx context.Context, name string, cfg ServerConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clients[name]; exists {
		return fmt.Errorf("mcp server already connected: %s", name)
	}

	client := NewClient(cfg.Command, cfg.Args...)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("mcp connect %s: %w", name, err)
	}
	r.clients[name] = client
	return nil
}

// RegisterTools registers all tools from connected MCP servers with the executor.
func (r *Registry) RegisterTools(ctx context.Context, executor *skill.Executor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for serverName, cl := range r.clients {
		for _, tool := range cl.Tools() {
			toolName := serverName + "_" + tool.Name
			fn := func(c *Client, n string) skill.ToolFunc {
				return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
					return c.CallTool(ctx, n, args)
				}
			}(cl, tool.Name)
			executor.RegisterTool(toolName, fn)
		}
	}
	return nil
}

// CloseAll shuts down all MCP server connections.
func (r *Registry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, client := range r.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", name, err))
		}
		delete(r.clients, name)
	}
	if len(errs) > 0 {
		return fmt.Errorf("mcp close errors: %v", errs)
	}
	return nil
}
