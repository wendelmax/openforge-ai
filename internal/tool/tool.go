package tool

import "context"

type ToolResult struct {
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

type Tool interface {
	Name() string
	Description() string
	Run(ctx context.Context, args map[string]any) (ToolResult, error)
	InputSchema() map[string]any
}
