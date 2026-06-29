package tool

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type ViewTool struct{}

func NewViewTool() *ViewTool { return &ViewTool{} }
func (t *ViewTool) Name() string { return "view" }
func (t *ViewTool) Description() string { return "Read a file with offset/limit. Args: {path, offset, limit}" }
func (t *ViewTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"path": map[string]any{"type": "string"}, "offset": map[string]any{"type": "integer", "default": 0}, "limit": map[string]any{"type": "integer", "default": 200}},
		"required": []string{"path"},
	}
}
func (t *ViewTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	if path == "" { return ToolResult{Error: "path is required"}, nil }
	offset := getInt(args, "offset", 0)
	limit := getInt(args, "limit", 200)
	if limit <= 0 { limit = 200 }
	data, err := os.ReadFile(path)
	if err != nil { return ToolResult{Error: fmt.Sprintf("read: %v", err)}, nil }
	lines := strings.Split(string(data), "\n")
	if offset >= len(lines) { return ToolResult{Content: ""}, nil }
	end := offset + limit
	if end > len(lines) { end = len(lines) }
	var b strings.Builder
	for i := offset; i < end; i++ {
		b.WriteString(fmt.Sprintf("%6d| %s\n", i+1, lines[i]))
	}
	return ToolResult{Content: b.String()}, nil
}
