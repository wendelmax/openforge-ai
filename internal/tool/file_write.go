package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type WriteTool struct{}

func NewWriteTool() *WriteTool { return &WriteTool{} }
func (t *WriteTool) Name() string { return "write" }
func (t *WriteTool) Description() string { return "Create/overwrite a file. Args: {path, content}" }
func (t *WriteTool) InputSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"path": map[string]any{"type": "string"}, "content": map[string]any{"type": "string"}},
		"required": []string{"path", "content"},
	}
}
func (t *WriteTool) Run(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" { return ToolResult{Error: "path is required"}, nil }
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolResult{Error: fmt.Sprintf("write: %v", err)}, nil
	}
	return ToolResult{Content: fmt.Sprintf("Wrote %d bytes to %s", len(content), path)}, nil
}
