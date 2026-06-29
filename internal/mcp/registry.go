package mcp

import (
	"context"
	"fmt"
	"sync"
)

type Registry struct {
	mu      sync.Mutex
	clients map[string]*Client
}

func NewRegistry() *Registry {
	return &Registry{clients: make(map[string]*Client)}
}

func (r *Registry) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, 0, len(r.clients))
	for name := range r.clients {
		names = append(names, name)
	}
	return names
}

func (r *Registry) Connect(ctx context.Context, name string, command string, args ...string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.clients[name]; exists {
		return fmt.Errorf("mcp server already connected: %s", name)
	}
	client := NewClient(command, args...)
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("mcp connect %s: %w", name, err)
	}
	r.clients[name] = client
	return nil
}

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

func (r *Registry) GetClient(name string) (*Client, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[name]
	return c, ok
}

func (r *Registry) GetAllTools() []MCPTool {
	r.mu.Lock()
	defer r.mu.Unlock()
	var tools []MCPTool
	for srv, cl := range r.clients {
		for _, t := range cl.Tools() {
			tools = append(tools, MCPTool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: convertInputSchema(t.InputSchema),
				ServerName:  srv,
			})
		}
	}
	return tools
}

func convertInputSchema(s InputSchema) map[string]any {
	props := make(map[string]any)
	for k, v := range s.Properties {
		props[k] = map[string]any{
			"type":        v.Type,
			"description": v.Description,
			"enum":        v.Enum,
		}
	}
	return map[string]any{
		"type":       s.Type,
		"properties": props,
		"required":   s.Required,
	}
}
